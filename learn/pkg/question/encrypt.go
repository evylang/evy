package question

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
)

// PublicKey is a 1024 bit RSA public key. This public key is the default
// sealing key used to encrypt Evy answer keys which are then decrypted and
// deployed by CI. If this key changes the key on CI needs to change for
// learn.evy.dev to work.
//
// For strong cryptography use 2048 or 4096 bit keys. This key is used to
// encrypt answers in plain text frontmatter, where a single characters
// answer, e.g. 'a' becomes an encrypted ~370 bytes for 2048 keys, versus
// ~190 bytes for 1024 key. We accept the weaker encryption as it still
// demonstrates encryption principles and is deemed a reasonable compromise
// in terms of space and security required for the answer key learning
// platform.
const PublicKey = "MIGJAoGBANipT8zrt3mgsU449ZQ5Z7MoP/wl4w1UMzLRBeI/GTNC4xqKXLL1fhvdxz0Vp39fSsGqRVS4kgz3n3aZpGY+YDYY6VoKP2h/zaIC+NO0oPo6eKQkjI+OTTpkg1a1Ymh+XTxl5KeLrslni5ygMVzwWVP9wZU6I+RJXxu2N4cosJD/AgMBAAE="

const sessionKeyBytes = 32

// KeyPair represents a pair of public and private keys.
type KeyPair struct {
	Public  string
	Private string
}

// Keygen generates a new RSA key pair for the given length. For strong
// cryptography use 2048 or 4096 bit keys.
func Keygen(length int) (KeyPair, error) {
	private, err := rsa.GenerateKey(rand.Reader, length)
	if err != nil {
		return KeyPair{}, err
	}
	return KeyPair{
		Private: base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(private)),
		Public:  base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&private.PublicKey)),
	}, nil
}

// Encrypt performs encryption using a combination of AES-GCM and RSA-OAEP
// algorithms:
//
//  1. Key Generation: It generates a random symmetric key for AES-GCM
//     encryption.
//  2. Message Encryption: The plaintext message is encrypted with the
//     generated symmetric key.
//  3. This symmetric key is then encrypted using the provided public key
//     associated with the RSA-OAEP asymmetric encryption algorithm.
//
// The output format is a byte string structured as follows:
//
//	   Version (1 byte) || RSA ciphertext length (2 bytes) || RSA ciphertext || AES ciphertext
//
//	- Version (1 byte): Identifies the encryption format version.
//	- RSA ciphertext length (2 bytes): Indicates the length of the RSA-OAEP
//	  encrypted symmetric key.
//	- RSA ciphertext: Encrypted symmetric key using the public key.
//	- AES ciphertext: Encrypted message using the generated symmetric key.
//
// The Encrypt function returns the ciphertext, i.e. an encrypted base64
// encoded byte string. This cipher text can be decrypt with
// [Decrypt] function.
//
// The Encrypt function is derived from the Apache 2 licensed Sealed Secret
// code by Bitnami Labs:
// https://github.com/bitnami-labs/sealed-secrets/blob/release/v0.20.5/pkg/crypto/crypto.go
func Encrypt(publicKeyB64, plaintext string) (string, error) {
	publicKey, err := parsePublicKey(publicKeyB64)
	if err != nil {
		return "", fmt.Errorf("bad key: %w", err)
	}
	ciphertext, err := hybridEncrypt(publicKey, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt recovers the original plaintext message from an encrypted byte string.
// It performs decryption using a combination of AES-GCM and RSA-OAEP algorithms.
//
//  1. Parse Input: It expects two arguments:
//     - privateKeyB64 (string): The base64 encoded representation of your
//     secret private key for RSA-OAEP decryption.
//     - ciphertext (string): The encrypted byte string containing the message
//     and associated data.
//
//  2. Decrypt RSA Ciphertext: The function first uses your private key to
//     decrypt the RSA-OAEP ciphertext, which retrieves the original symmetric
//     key used for encryption.
//
//  3. Decrypt Message: It then utilizes the recovered symmetric key to
//     decrypt the AES-GCM encrypted message within the ciphertext.
//     - On success, it returns the decrypted plaintext message as a string.
//     - On failure, it returns an error object describing the encountered
//     issue during decryption.
//
// The Decrypt function assumes that the ciphertext is generated by the
// [Encrypt] function.
//
// The Decrypt function is derived from the Apache 2 licensed Sealed Secret
// code by Bitnami Labs:
// https://github.com/bitnami-labs/sealed-secrets/blob/release/v0.20.5/pkg/crypto/crypto.go
func Decrypt(privateKeyB64, ciphertext string) (string, error) {
	cipher, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("cannot base64 decode: %w", err)
	}
	privateKey, err := parsePrivateKey(privateKeyB64)
	if err != nil {
		return "", fmt.Errorf("cannot parse private key: %w", err)
	}
	plaintext, err := hybridDecrypt(privateKey, cipher)
	if err != nil {
		return "", fmt.Errorf("cannot decrypt: %w", err)
	}
	return string(plaintext), nil
}

func parsePublicKey(key string) (*rsa.PublicKey, error) {
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PublicKey(b)
}

func parsePrivateKey(key string) (*rsa.PrivateKey, error) {
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return x509.ParsePKCS1PrivateKey(b)
}

// hybridEncrypt executes a regular AES-GCM + RSA-OAEP encryption. The output
// byte string is:
//
//	Version (1 byte) || RSA ciphertext length (2 bytes) || RSA ciphertext || AES ciphertext
//
// This function is derived from the Apache 2 licensed Sealed Secret code by
// Bitnami Labs:
// https://github.com/bitnami-labs/sealed-secrets/blob/release/v0.20.5/pkg/crypto/crypto.go#L36
func hybridEncrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	// Generate a random symmetric key.
	sessionKey := make([]byte, sessionKeyBytes)
	if _, err := io.ReadFull(rand.Reader, sessionKey); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Encrypt symmetric key.
	rsaCiphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, sessionKey, nil)
	if err != nil {
		return nil, err
	}

	// First 3 bytes are RSA ciphertext length, so we can separate all the
	// pieces later.
	ciphertext := make([]byte, 3)
	ciphertext[0] = 1 // Version
	binary.BigEndian.PutUint16(ciphertext[1:], uint16(len(rsaCiphertext)))
	ciphertext = append(ciphertext, rsaCiphertext...) //nolint:makezero // We want to initialize the first 3 bytes and then append, this is correct.

	// SessionKey is only used once, so zero nonce is ok.
	zeroNonce := make([]byte, gcm.NonceSize())
	// Append symmetrically encrypted Secret.
	ciphertext = gcm.Seal(ciphertext, zeroNonce, plaintext, nil)
	return ciphertext, nil
}

// hybridDecrypt performs a regular AES-GCM + RSA-OAEP decryption.
//
// This function is derived from the Apache 2 licensed Sealed Secret code by
// Bitnami Labs:
// https://github.com/bitnami-labs/sealed-secrets/blob/release/v0.20.5/pkg/crypto/crypto.go#L87
func hybridDecrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 3 {
		return nil, ErrSealedTooShort
	}
	rsaLen := int(binary.BigEndian.Uint16(ciphertext[1:]))
	if len(ciphertext) < rsaLen+3 {
		return nil, ErrSealedTooShort
	}

	rsaCiphertext := ciphertext[3 : rsaLen+3]
	aesCiphertext := ciphertext[rsaLen+3:]

	sessionKey, err := rsa.DecryptOAEP(sha256.New(), nil, privateKey, rsaCiphertext, nil)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Key is only used once, so zero nonce is ok.
	zeroNonce := make([]byte, gcm.NonceSize())
	plaintext, err := gcm.Open(nil, zeroNonce, aesCiphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
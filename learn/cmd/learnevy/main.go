package main

import (
	"encoding/json"
	"fmt"
	"os"

	"evylang.dev/evy/learn/pkg/answer"
	"github.com/alecthomas/kong"
)

const description = `
learnevy is a tool that manages practice and learn resources for Evy.
`

type app struct {
	Seal   sealCmd   `cmd:"" help:"Move 'answer' to 'sealed-answer' in source markdown."`
	Unseal unsealCmd `cmd:"" help:"Move 'sealed-answer' to 'answer' in source markdown."`
	Export exportCmd `cmd:"" help:"Export Answerkey File."`

	Verify verifyCmd `cmd:"" help:"Verify encryptedAnsers in markdown file. Ensure no plaintext answers."`
	Serve  serveCmd  `cmd:"" help:"Start HTTP server with contents."`

	Crypto cryptoCmd `cmd:"" help:"Encryption utilities."`
}

type cryptoCmd struct {
	Keygen keygenCryptoCmd `cmd:"" help:"Generate a new secret key."`
	Seal   sealCryptoCmd   `cmd:"" help:"Encrypt a string given on command line"`
	Unseal unsealCryptoCmd `cmd:"" help:"Decrypt string given on command line"`
}

func main() {
	kopts := []kong.Option{kong.Description(description)}
	kctx := kong.Parse(&app{}, kopts...)
	kctx.FatalIfErrorf(kctx.Run())
}

type keygenCryptoCmd struct {
	Length int `short:"l" default:"2048" help:"Length of key to generate."`
}
type sealCryptoCmd struct {
	Plaintext string `arg:"" help:"Plaintext to encrypt."`
}
type unsealCryptoCmd struct {
	Ciphertext string `arg:"" help:"Ciphertext to decrypt."`
	PrivateKey string `short:"s" help:"Secret private key to decrypt ciphertext." env:"EVY_LEARN_PRIVATE_KEY"`
}

func (c *keygenCryptoCmd) Run() error {
	keys, err := answer.Keygen(c.Length)
	if err != nil {
		return err
	}
	fmt.Printf("private: %s\n\npublic:  %s\n", keys.Private, keys.Public)
	return nil
}

func (c *sealCryptoCmd) Run() error {
	encrypted, err := answer.Encrypt(answer.PublicKey, c.Plaintext)
	if err != nil {
		return err
	}
	fmt.Println(encrypted)
	return nil
}

func (c *unsealCryptoCmd) Run() error {
	plaintext, err := answer.Decrypt(c.PrivateKey, c.Ciphertext)
	if err != nil {
		return err
	}
	fmt.Println(plaintext)
	return nil
}

type exportCmd struct {
	MDFile        string `arg:"" type:"markdownfile" help:"Markdown file with course, unit, exercise, or question." placeholder:"ANSWERFILE"`
	AnswerkeyFile string `arg:"" default:"-" help:"JSON output file for answer key (default: stdout)." placeholder:"ANSWERFILE"`
	UnsealedOnly  bool   `short:"u" help:"Only export files with unsealed answers. Suitable if private key not available."`
	PrivateKey    string `short:"k" help:"Secret private key to decrypt sealed answers." env:"EVY_LEARN_PRIVATE_KEY"`
}

type sealCmd struct {
	MDFile string `arg:"" type:"markdownfile" help:"Markdown file with course, unit, exercise, or question." placeholder:"ANSWERFILE"`
}

type unsealCmd struct {
	MDFile     string `arg:"" type:"markdownfile" help:"Markdown file with course, unit, exercise, or question." placeholder:"ANSWERFILE"`
	PrivateKey string `short:"k" help:"Secret private key to decrypt sealed answers." env:"EVY_LEARN_PRIVATE_KEY"`
}

type verifyCmd struct {
	MDFile       string `arg:"" type:"markdownfile" help:"Markdown file with course, unit, exercise, or question." placeholder:"ANSWERFILE"`
	Type         string `arg:"" default:"all" enum:"all,result,seal" help:"Type of verification to perform."`
	UnsealedOnly bool   `short:"u" help:"Only check result for files with unsealed answers. Suitable if private key not available."`
	PrivateKey   string `short:"k" help:"Secret private key to decrypt sealed answers." env:"EVY_LEARN_PRIVATE_KEY"`
}
type serveCmd struct {
	MDFile       string `arg:"" type:"markdownfile" help:"Markdown file with course, unit, exercise, or question." placeholder:"ANSWERFILE"`
	UnsealedOnly bool   `short:"u" help:"Only serve files with unsealed answers. Suitable if private key not available."`
	PrivateKey   string `short:"k" help:"Secret private key to decrypt sealed answers." env:"EVY_LEARN_PRIVATE_KEY"`
}

func (c *exportCmd) Run() error {
	md, err := answer.NewQuestionMarkdown(c.MDFile)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(c.AnswerkeyFile, os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close() //nolint

	fmt.Println("creating answerkey file", c.AnswerkeyFile)
	answerkey, err := md.ExportAnswerKey(c.PrivateKey)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(answerkey)
}

func (c *unsealCmd) Run() error {
	fmt.Println("unseal command not yet implemented")
	return nil
}
func (c *verifyCmd) Run() error {
	fmt.Println("verify command not yet implemented")
	return nil
}
func (c *serveCmd) Run() error {
	fmt.Println("serve command not yet implemented")
	return nil
}

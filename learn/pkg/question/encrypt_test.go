package question

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

const (
	// 2048 bit RSA keys.
	testKeyPrivate = "MIIEpQIBAAKCAQEAuNEufiuryg/OZPKVUbaIRam1UNqju5binwrRzsOGWkM6DYKqxW2tA+O7dhg9do/Jm0lr+rkVqf8CR/HejD08n9OTsHe0NeblLwZncQX1J3ayyGsu+xAFxQ0hvFfG+Vy8KXJAgug6CCsaiVgBwOWPdfEOqEDv5S5XlnwQh9dxWB8m/1CTDmqSdIhYnzQQp13ZyumCRgrIHKSYPR3KCZD8KLRvkoIrF0DU18f6ASO7wjv7FBhgQ2ZAR/Yud/h6ceQKvAW0W3MmPiJblZhbrsPQGi7eZZo4K8aAvuzQmcYq17/E/e6MnOweoyik4lIAG0uGa7FiY5f9NVuir7JPA2lCLwIDAQABAoIBAQCJHEcNu4BbC5bnNUCpum0moVyue0X1KV8+9lvotQ27cRxkYYgnp9IvjIfKePlAODQtTC8bdqwnzdP3Y+zixZtwRxrOVEARrRZh6LJdGzpg6KKCJWJZR+2/3pokjEpFPRMq/GP3uikzXib1taC3ZpcjvI5PLL3MnLDGJ4xr+t1Pral8BXSILhUSQzgMFAB8+5V+zWnUPuPzCeym3VeYpZSdbSsR+CZnxy4vbB4cSj97M1MgBTOPocduE5cRrE8mumAk93dzBmKH+/potjLOMhCiJJFVtPO9GXLLduLAH9qKwSk2vJytIX8KwYFTCve2EKhMB9ydBhk09zVoELUle0mhAoGBAOc9agk5CahkNOVO1E3Cw1zK+Da+2LXYk2HhjTpOTkr2lKji8v1eSDkk5R72ZfPrI5s8sBrkW0OqPJVXDnmho78quWHTwxvJrnrIcuZLa1Kn4H+cHN81J9jGcim7kLPTZUcnU0RMR7Xn3lT61H5lB3LSFplRq52tqS5AaxaksS0tAoGBAMybQjceAVTihCHKFkaFV8Ys2dm5p5ejCzYklY+jA0UdTmHT6kmr13KIA6k61+s8kyZDaGutZ6lRyHuCfotL6j6jr8rsn/EbDikZ4/XhhO9+B+xJMXolKLFA+/pBPxNs7KLSjZ3mH7N0qzxbQzyVWF4BhSxTxIjWEGAtc1ZUJN5LAoGBAMYPFzRhE0GU2q2RkEwuRnDDNEiHvEw8/Td4HiPTkEGq4/ens2KKj6fKTyju+LIsM6oyF9BgyT6yoAN1tmM9rGf/qxr8av/xBa4K5EcWUA1S1vnV9/DCsad9iajvC2jK5tND/pDgGQfYWtlEoh7EX9Xb1hlqF2kNpnuEF3UkiNDdAoGBAJzuMFlKAEd0/VdVQsSQHYR4fhbKmMprWXwLj1L9+tIV6jqKaVZcIQFNZVF1OorIiSx94ydDdxCdE6H3sstwTJgCwCBqYTpyP+gyXXAHqwhtp/IJKZO/0HgzmZCWXqStlMpFqC0FhicEQxol/WoIOiDQFa6sCT/Sv/iko6QBIc4FAoGAMSC5SUsgUiHo6gvp2put1ySmJIVj3roqI6mAndi2hLVMalF1Q5F4X4HVHWqOj7QA7zpf3ATotCI4AbmfOwpFCZ4rEP0QsbV2uZ/3NhxwAE1MWrv+ht2ONe74sOYg7Z+XAjD7TW7We3KTewerVnC/VotKZ+3Eq2FgelSYDvlNmoQ="
	testKeyPublic  = "MIIBCgKCAQEAuNEufiuryg/OZPKVUbaIRam1UNqju5binwrRzsOGWkM6DYKqxW2tA+O7dhg9do/Jm0lr+rkVqf8CR/HejD08n9OTsHe0NeblLwZncQX1J3ayyGsu+xAFxQ0hvFfG+Vy8KXJAgug6CCsaiVgBwOWPdfEOqEDv5S5XlnwQh9dxWB8m/1CTDmqSdIhYnzQQp13ZyumCRgrIHKSYPR3KCZD8KLRvkoIrF0DU18f6ASO7wjv7FBhgQ2ZAR/Yud/h6ceQKvAW0W3MmPiJblZhbrsPQGi7eZZo4K8aAvuzQmcYq17/E/e6MnOweoyik4lIAG0uGa7FiY5f9NVuir7JPA2lCLwIDAQAB"
)

func TestKeygen(t *testing.T) {
	keys, err := Keygen(1024)
	assert.NoError(t, err)

	_, err = parsePrivateKey(keys.Private)
	assert.NoError(t, err)

	_, err = parsePublicKey(keys.Public)
	assert.NoError(t, err)
}

func TestEncryptDecrypt(t *testing.T) {
	plaintexts := map[string]string{
		"short": "Hello, world!",
		"long":  "On the other hand, we denounce with righteous indignation and dislike men who are so beguiled and demoralized by the charms of pleasure of the moment, so blinded by desire, that they cannot foresee the pain and trouble that are bound to ensue; and equal blame belongs to those who fail in their duty through weakness of will, which is the same as saying through shrinking from toil and pain. These cases are perfectly simple and easy to distinguish. In a free hour, when our power of choice is untrammelled and when nothing prevents our being able to do what we like best, every pleasure is to be welcomed and every pain avoided. But in certain circumstances and owing to the claims of duty or the obligations of business it will frequently occur that pleasures have to be repudiated and annoyances accepted. The wise man therefore always holds in these matters to this principle of selection: he rejects pleasures to secure other greater pleasures, or else he endures pains to avoid worse pains.",
	}
	for label, plaintext := range plaintexts {
		t.Run(label, func(t *testing.T) {
			ciphertext, err := Encrypt(testKeyPublic, plaintext)
			assert.NoError(t, err)
			decrypted, err := Decrypt(testKeyPrivate, ciphertext)
			assert.NoError(t, err)
			assert.Equal(t, plaintext, decrypted)
		})
	}
}

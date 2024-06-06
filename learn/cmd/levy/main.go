// Levy is a tool for creating Evy practice and learn materials.
//
// Levy has the following sub-commands: export, verify, seal, unseal.
//
//	Usage: levy <command> [flags]
//
//	levy is a tool that manages learn and practice resources for Evy.
//
//	Flags:
//	  -h, --help       Show context-sensitive help.
//	  -V, --version    Print version information
//
//	Commands:
//	  export <export-type> <md-file> [<target>] [flags]
//	    Export answer key and HTML Files.
//
//	  verify <md-file> [<type>] [flags]
//	    Verify answers in markdown file.
//
//	  seal <md-file> [flags]
//	    Move 'answer' to 'sealed-answer' in source markdown.
//
//	  unseal <md-file> [flags]
//	    Move 'sealed-answer' to 'answer' in source markdown.
//
//	Run "levy <command> --help" for more information on a command.
package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"evylang.dev/evy/learn/pkg/learn"
	"github.com/alecthomas/kong"
)

const description = `
levy is a tool that manages learn and practice resources for Evy.
`

var version = "v0.0.0"

type app struct {
	Export exportCmd `cmd:"" help:"Export answer key and HTML Files."`
	Verify verifyCmd `cmd:"" help:"Verify answers in markdown file."`

	Seal   sealCmd   `cmd:"" help:"Move 'answer' to 'sealed-answer' in source markdown."`
	Unseal unsealCmd `cmd:"" help:"Move 'sealed-answer' to 'answer' in source markdown."`

	Version kong.VersionFlag `short:"V" help:"Print version information"`

	Crypto cryptoCmd `cmd:"" help:"Encryption utilities." hidden:""`
}

func main() {
	kopts := []kong.Option{
		kong.Description(description),
		kong.Vars{"version": version},
	}
	kctx := kong.Parse(&app{}, kopts...)
	kctx.FatalIfErrorf(kctx.Run())
}

type exportCmd struct {
	ExportType   string `arg:"" enum:"html,answerkey,all" help:"Export target: one of html, answerkey, all."`
	MDFile       string `arg:"" help:"Question markdown file." placeholder:"MDFILE"`
	Target       string `arg:"" default:"-" help:"Output directory or JSON/HTML output file (default: . | stdout)." placeholder:"TARGET"`
	IgnoreSealed bool   `short:"i" help:"Only export answerkey and add solution to unsealed answers. Suitable if private key not available."`
	PrivateKey   string `short:"k" help:"Secret private key to decrypt sealed answers." env:"EVY_LEARN_PRIVATE_KEY"`
	WithMarked   bool   `short:"m" help:"Include marked answers in HTML output. Cannot be used with export target answerkey."`

	htmlPath      string
	answerKeyPath string
}

type verifyCmd struct {
	MDFile       string `arg:"" help:"Question markdown file." placeholder:"MDFILE"`
	IgnoreSealed bool   `short:"i" help:"Only verify result for files with unsealed answers. Suitable if private key not available."`
	PrivateKey   string `short:"k" help:"Secret private key to decrypt sealed answers." env:"EVY_LEARN_PRIVATE_KEY"`

	// TODO
	Type string `arg:"" default:"all" enum:"all,result,seal" help:"Type of verification to perform (currently unused)." hidden:""`
}

type sealCmd struct {
	MDFile    string `arg:"" help:"Question markdown file." placeholder:"MDFILE"`
	PublicKey string `short:"k" help:"Public key to seal answers, default: built-in key"`
}

type unsealCmd struct {
	MDFile     string `arg:"" help:"Question markdown file." placeholder:"MDFILE"`
	PrivateKey string `short:"k" help:"Secret private key to decrypt sealed answers." env:"EVY_LEARN_PRIVATE_KEY"`
}

func (c *exportCmd) Run() error {
	if c.ExportType == "answerkey" && c.WithMarked {
		return errors.New(`--with-marked can only be used with all "all" and "html" export targets`) //nolint:err113 // dynamic errors in main are fine.
	}
	opts := getOptions(c.IgnoreSealed, c.PrivateKey)
	model, err := learn.NewQuestionModel(c.MDFile, opts...)
	if err != nil {
		return err
	}
	if err := c.setPaths(); err != nil {
		return err
	}
	if c.ExportType == "answerkey" || c.ExportType == "all" {
		answerKeyJSON, err := model.ExportAnswerKeyJSON()
		if err != nil {
			return err
		}
		if err := writeFileOrStdout(c.answerKeyPath, answerKeyJSON); err != nil {
			return err
		}
	}
	if c.ExportType == "html" || c.ExportType == "all" {
		html, err := model.ToHTML(c.WithMarked)
		if err != nil {
			return err
		}
		if err := writeFileOrStdout(c.htmlPath, html); err != nil {
			return err
		}
	}
	return nil
}

func writeFileOrStdout(filename, content string) error {
	if filename == "-" {
		fmt.Println(content)
		return nil
	}
	return os.WriteFile(filename, []byte(content), 0o666)
}

func (c *exportCmd) setPaths() error {
	c.htmlPath = c.Target
	c.answerKeyPath = c.Target
	if c.ExportType == "all" {
		if c.Target == "-" { // default
			c.Target = "."
		} else {
			if err := os.MkdirAll(c.Target, 0o755); err != nil {
				return err
			}
		}
		htmlFile := strings.TrimSuffix(filepath.Base(c.MDFile), filepath.Ext(c.MDFile)) + ".html"
		c.htmlPath = filepath.Join(c.Target, htmlFile)
		c.answerKeyPath = filepath.Join(c.Target, "answerkey.json")
	}
	return nil
}

func (c *verifyCmd) Run() error {
	opts := getOptions(c.IgnoreSealed, c.PrivateKey)
	model, err := learn.NewQuestionModel(c.MDFile, opts...)
	if err != nil {
		return err
	}
	return model.Verify()
}

func (c *sealCmd) Run() error {
	model, err := learn.NewQuestionModel(c.MDFile)
	if err != nil {
		return err
	}
	publicKey := cmp.Or(c.PublicKey, learn.PublicKey)
	if err := model.Seal(publicKey); err != nil {
		return err
	}
	return model.WriteFormatted()
}

func (c *unsealCmd) Run() error {
	model, err := learn.NewQuestionModel(c.MDFile, learn.WithPrivateKey(c.PrivateKey))
	if err != nil {
		return err
	}
	if err := model.Unseal(); err != nil {
		return err
	}
	fmt.Println("Unsealed answer:", model.Frontmatter.Answer)
	return nil
}

type cryptoCmd struct {
	Keygen keygenCryptoCmd `cmd:"" help:"Generate a new secret key."`
	Seal   sealCryptoCmd   `cmd:"" help:"Encrypt a string given on command line"`
	Unseal unsealCryptoCmd `cmd:"" help:"Decrypt string given on command line"`
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
	keys, err := learn.Keygen(c.Length)
	if err != nil {
		return err
	}
	fmt.Printf("private: %s\n\npublic:  %s\n", keys.Private, keys.Public)
	return nil
}

func (c *sealCryptoCmd) Run() error {
	encrypted, err := learn.Encrypt(learn.PublicKey, c.Plaintext)
	if err != nil {
		return err
	}
	fmt.Println(encrypted)
	return nil
}

func (c *unsealCryptoCmd) Run() error {
	plaintext, err := learn.Decrypt(c.PrivateKey, c.Ciphertext)
	if err != nil {
		return err
	}
	fmt.Println(plaintext)
	return nil
}

func getOptions(ignoreSealed bool, privateKey string) []learn.Option {
	var options []learn.Option
	if ignoreSealed {
		options = append(options, learn.WithIgnoreSealed())
	}
	if !ignoreSealed && privateKey != "" {
		options = append(options, learn.WithPrivateKey(privateKey))
	}
	return options
}

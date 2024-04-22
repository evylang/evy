package main

import (
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/alecthomas/kong"
	"google.golang.org/api/option"
)

var opts []option.ClientOption

func main() {
	var cli struct {
		CredentialFile string `type:"existingfile" env:"EVY_FIREBASE_CREDENTIAL_FILE" help:"Firebase credential file."`
	}
	kong.Parse(&cli)
	if cli.CredentialFile != "" {
		opts = append(opts, option.WithCredentialsFile(cli.CredentialFile))
	}
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// doc, err := client.Doc("answerkey/v1/courses/fundamentals").Get(ctx)
	// checkError(err)
	// answerkey := Answerkey{}
	// doc.DataTo(&answerkey)
	// b, err := json.MarshalIndent(answerkey, "", "  ")
	// checkError(err)
	// fmt.Println("answerkey/v0/courses/fundamentals size:", len(b))
	// answerkey, err = generateLargeAnswerKey(c.JSONFiles)
	// checkError(err)
	// b, err = json.MarshalIndent(answerkey, "", "  ")
	// checkError(err)
	// fmt.Println("answerkey/v0/courses/fundamentals:", len(b))
	// os.WriteFile("answerkey-large.json", b, 0644)
}

func handler(w http.ResponseWriter, r *http.Request) {
	client, err := firestore.NewClient(r.Context(), "evy-lang-test", opts...)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	defer client.Close()

	collection := client.Collection("default-collection")
	doc := collection.Doc("default-doc")
	snap, err := doc.Get(r.Context())
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	fmt.Fprintf(w, "V9 snap: %v\n", snap.Data())
}

func writeInternalServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("error: %v", err)))
}

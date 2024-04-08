package main

import (
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
)

//var opt = option.WithCredentialsFile("/Users/julia/Downloads/evy-lang-test-firebase-adminsdk-5ud3e-4ef53c5971.json")

func main() {
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	//client, err := firestore.NewClientWithDatabase(r.Context(), "evy-lang-test", "main", opt)
	//client, err := firestore.NewClient(r.Context(), "evy-lang-test", opt)
	client, err := firestore.NewClient(r.Context(), "evy-lang-test")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error: %v", err)))
		return
	}
	defer client.Close()

	collection := client.Collection("default-collection")
	doc := collection.Doc("default-doc")
	snap, err := doc.Get(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error: %v", err)))
		return
	}
	fmt.Fprintf(w, "V8 for Cam snap: %v\n", snap.Data())
}

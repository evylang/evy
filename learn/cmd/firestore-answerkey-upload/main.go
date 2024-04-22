package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/alecthomas/kong"
	"google.golang.org/api/option"
)

func main() {
	c := &cli{}
	kong.Parse(c)

	//answerkey, err := loadAnswerKey(c.File)
	answerkey, err := generateAnswerKey(c.JSONFiles)
	fmt.Println("answerkey size estimate (KB):", answerkey.estimateKB())

	f, err := os.Create("answerkey.json")
	checkError(err)
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	checkError(enc.Encode(answerkey))

	ctx := context.Background()
	client, err := newClient(ctx, c.DB, c.CredentialFile)
	checkError(err)
	defer client.Close()

	_, err = client.Doc("answerkey/v0/courses/fundamentals").Set(ctx, answerkey)
	checkError(err)
}

type Answerkey map[string]Answer

func (a Answerkey) estimateKB() int {
	b, _ := json.MarshalIndent(a, "", "  ")
	return len(b) / 1024
}

type Answer struct {
	Single  string   `firestore:"single,omitempty" json:"single,omitempty"`
	Multi   []string `firestore:"multi,omitempty" json:"multi,omitempty"`
	Texts   []string `firestore:"texts,omitempty" json:"texts,omitempty"`
	Text    string   `firestore:"text,omitempty" json:"text,omitempty"`
	Program string   `firestore:"program,omitempty" json:"program,omitempty"`
}

type cli struct {
	CredentialFile string `type:"existingfile" env:"EVY_FIREBASE_CREDENTIAL_FILE" help:"Firebase credential file."`
	DB             string `default:"evy-lang-test" help:"Firestore database name."`
	//File           string `arg:"" type:"existingfile" help:"Answerkey in JSON file."`
	JSONFiles []string `arg:"" type:"existingfile" help:"Unit structure JSON files."`
}

func loadAnswerKey(filename string) (Answerkey, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Answerkey doc around %d KB\n", stat.Size()/1024)
	answerkey := Answerkey{}
	err = json.NewDecoder(f).Decode(&answerkey)
	if err != nil {
		return nil, err
	}
	return answerkey, nil
}

func newClient(ctx context.Context, db, credentialFile string) (*firestore.Client, error) {
	opts := []option.ClientOption{}
	if credentialFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialFile))
	}
	return firestore.NewClient(ctx, db, opts...)
}

type Exercise struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Questions Questions `json:"questions"`
}

type Questions struct {
	Easy   map[string]Question `json:"easy"`
	Medium map[string]Question `json:"medium"`
	Hard   map[string]Question `json:"hard"`
}

func (q Questions) asMap() map[string]map[string]Question {
	return map[string]map[string]Question{
		"easy":   q.Easy,
		"medium": q.Medium,
		"hard":   q.Hard,
	}
}

type Question struct {
	Variants int    `json:"variants"`
	Type     string `json:"type"`
}

type Unit struct {
	ID        string     `json:"id"`
	Exercises []Exercise `json:"exercises"`
}

func generateAnswerKey(files []string) (Answerkey, error) {
	answerkey := Answerkey{}
	for _, file := range files {
		unit, err := loadUnit(file)
		if err != nil {
			return nil, err
		}
		if err := addToAnswerkey(unit.Exercises, answerkey); err != nil {
			return nil, err
		}
	}

	return answerkey, nil
}

func loadUnit(filename string) (Unit, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Unit{}, err
	}
	unit := Unit{}
	err = json.NewDecoder(f).Decode(&unit)
	if err != nil {
		return Unit{}, err
	}
	return unit, nil
}

func addToAnswerkey(exercises []Exercise, answerkey Answerkey) error {
	for _, ex := range exercises {
		for difficulty, questions := range ex.Questions.asMap() {
			for questionID, question := range questions {
				for i := range question.Variants {
					id := fmt.Sprintf("%s-%s-%s-%d", ex.ID, difficulty, questionID, i+1)
					answerkey[id] = generateAnswer(question.Type)
				}
			}
		}

	}
	return nil
}

func generateAnswer(questionType string) Answer {
	switch questionType {
	case "single":
		return Answer{Single: "a"}
	case "multi":
		return Answer{Multi: []string{"a", "b", "c"}}
	case "text":
		return Answer{Text: text}
	case "texts":
		return Answer{Texts: []string{text[:100], text[100:]}}
	case "program":
		return Answer{Program: program}
	}
	return Answer{}
}

func checkError(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

const text = "Imagine you're following a recipe to make slime.  In computer programming, things work kind of like that!  booluential execution means the computer follows the instructions in your code one by one, just like you follow the steps in the recipe."

const program = `while true
    print "╔════════════════════════╗"
    print "║ Guess my number (1-10) ║"
    print "╚════════════════════════╝"

    n := 1 + (rand 10)
    guess := readn

    while guess != n
        if guess < n
            print guess "Guess higher."
        else
            print guess "Guess lower."
        end
        guess = readn
    end
    print guess "You got it 🥳."
    sleep 2
    cls
end

// readn reads a number.
// If the input is not a number readn
// prints an error and repeats prompt.
func readn:num
    while true
        str := read
        n := str2num str
        if !err
            return n
        end
        print "💣 '"+str+"'is not a number. Try again."
    end
    return -1
end`

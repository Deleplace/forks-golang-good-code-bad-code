package parser

import (
	"io/ioutil"
	"testing"
)

func TestLex(t *testing.T) {
	b, err := ioutil.ReadFile("../resources/tests/adexp.txt")
	if err != nil {
		t.Fatal(err)
	}

	lexemes := lex(b)
	// nWords := len(lexemes)
	for _, word := range lexemes {
		switch {
		case word.isNewline():
			// fmt.Println("NEWLINE")
		case word.isCommand():
			// fmt.Printf("COMMAND %q \n", string(word))
		default:
			// fmt.Printf("  ARG %q \n", string(word))
		}
	}
}

package harness

import (
	"encoding/json"
	"fmt"
	"testing"
)

func ExampleSplit() {
	words, err := Split(`a-word 'cool' foo'bar'`)
	if err != nil {
		panic(err)
	}
	for _, word := range words {
		fmt.Println(word)
	}

	// Output:
	// a-word
	// cool
	// foo'bar'
}

func TestSimpleWord(t *testing.T) {
	var (
		words []string
		err   error
	)

	words, err = Split(`a-word`)
	equal(t, err, []string{"a-word"}, words)

	words, err = Split(`a-word foo bar`)
	equal(t, err, []string{"a-word", "foo", "bar"}, words)
}

func TestSingleQuoteWord(t *testing.T) {
	var (
		words []string
		err   error
	)

	words, err = Split(`'a-word'`)
	equal(t, err, []string{"a-word"}, words)

	words, err = Split(`'a-word' 'foo' 'bar'`)
	equal(t, err, []string{"a-word", "foo", "bar"}, words)

	words, err = Split(`'a-word' 'foo' cool'bar'`)
	equal(t, err, []string{"a-word", "foo", "cool'bar'"}, words)
}

func TestDoubleQuoteWord(t *testing.T) {
	var (
		words []string
		err   error
	)

	words, err = Split(`"a-word"`)
	equal(t, err, []string{"a-word"}, words)

	words, err = Split(`\a-word`)
	equal(t, err, []string{"\\a-word"}, words)

	words, err = Split(`"a-word" "foo" "bar"`)
	equal(t, err, []string{"a-word", "foo", "bar"}, words)

	words, err = Split(`"a \"word\"" "foo \\ baz" cool"bar"`)
	equal(t, err, []string{"a \"word\"", "foo \\ baz", "bar"}, words)
}

func equal(t *testing.T, e error, exp, act []string) {
	t.Helper()

	if e != nil {
		t.Fatalf("err: %s", e)
	}

	err := false

	if len(exp) != len(act) {
		err = true
	} else {
		for i, w := range exp {
			if w != act[i] {
				err = true
				break
			}
		}
	}

	if err {
		exp_json, _ := json.Marshal(exp)
		act_json, _ := json.Marshal(act)
		t.Fatalf("expected %s but got %s", exp_json, act_json)
	}
}

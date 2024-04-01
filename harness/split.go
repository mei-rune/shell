package harness

import (
	"unicode"

	"github.com/runner-mei/errors"
)

func readQuoteString(txt []rune, isDq bool) ([]rune, []rune, error) {
	var word []rune

	isEscape := false
	for idx := 0; idx < len(txt); idx++ {
		c := txt[idx]

		switch c {
		case '\\':
			if isEscape {
				isEscape = false
				word = append(word, c)
			} else {
				isEscape = true
			}
		case 't':
			if isEscape {
				isEscape = false
				word = append(word, '\t')
			} else {
				word = append(word, c)
			}
		case 'r':
			if isEscape {
				isEscape = false
				word = append(word, '\r')
			} else {
				word = append(word, c)
			}
		case 'n':
			if isEscape {
				isEscape = false
				word = append(word, '\n')
			} else {
				word = append(word, c)
			}
		case '\'':
			if !isEscape {
				if !isDq {
					return word, txt[idx+1:], nil
				}
			}

			isEscape = false
			word = append(word, c)
		case '"':
			if !isEscape {
				if isDq {
					return word, txt[idx+1:], nil
				}
			}

			isEscape = false
			word = append(word, c)
		default:
			word = append(word, c)
		}
	}

	return word, nil, errors.New("Expected a `\"` (double quote)")
}

func readIdentString(txt []rune) (string, []rune, []rune, error) {
	for idx, c := range txt {
		if unicode.IsSpace(c) {
			return "", txt[:idx], txt[idx:], nil
		}

		switch c {
		case '\'':
		case '"':
			word, line, err := readQuoteString(txt[idx+1:], true)
			if err != nil {
				return "", nil, nil, err
			}
			return string(txt[:idx]), word, line, nil
		}
	}
	return "", txt, nil, nil
}

func skipWhitespace(txt []rune) (bool, []rune) {
	for idx, c := range txt {
		if !unicode.IsSpace(c) {
			return idx != 0, txt[idx:]
		}
	}
	return false, nil
}

func split(line []rune) ([]int, []string, []string, error) {
	var types = make([]int, 0, 16)
	var charsets = make([]string, 0, 16)
	var words = make([]string, 0, 16)

	_, line = skipWhitespace(line)
	for len(line) > 0 {
		switch line[0] {
		case '"':
			word, remain, err := readQuoteString(line[1:], true)
			if err != nil {
				return nil, nil, nil, err
			}
			types = append(types, 0)
			charsets = append(charsets, "")
			words = append(words, string(word))
			line = remain
		case '\'':
			word, remain, err := readQuoteString(line[1:], false)
			if err != nil {
				return nil, nil, nil, err
			}
			types = append(types, 0)
			charsets = append(charsets, "")
			words = append(words, string(word))
			line = remain
		default:
			charset, word, remain, err := readIdentString(line)
			if err != nil {
				return nil, nil, nil, err
			}
			if charset == "" {
				types = append(types, 1)
			} else {
				types = append(types, 0)
			}
			charsets = append(charsets, charset)
			words = append(words, string(word))
			line = remain
		}
		_, line = skipWhitespace(line)
	}

	return types, charsets, words, nil
}

func Split(line string) ([]string, error) {
	_, _, ss, err := split([]rune(line))
	return ss, err
}

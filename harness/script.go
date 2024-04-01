package harness

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type CommandFunc func(ctx context.Context, script *Script, conn *Shell) error

type Command struct {
	LineNumber int
	LineText   string
	Command    string
	Run        CommandFunc
}

type Script struct {
	Cmds []Command
}

func (self *Script) Output() []byte {
	return []byte("")
}

type scriptResult struct {
	results []ExecuteResult
	err     error
}

func (se *scriptResult) Unwarp() error {
	return se.err
}

func (se *scriptResult) Error() string {
	return se.err.Error()
}

func (self *Script) Run(ctx context.Context, conn *Shell) ([]ExecuteResult, error) {
	var results = make([]ExecuteResult, 0, len(self.Cmds))
	for _, cmd := range self.Cmds {

		var in strings.Builder
		var out strings.Builder

		c1 := conn.SetTeeReader(&in)
		c2 := conn.SetTeeWriter(&out)

		err := cmd.Run(ctx, self, conn)

		c1()
		c2()

		results = append(results, ExecuteResult{
			LineNumber: cmd.LineNumber,
			LineText:   cmd.LineText,
			Command:    cmd.Command,
			Incomming:  in.String(),
			Outgoing:   out.String(),
		})

		if err == nil {
			incomming := results[len(results)-1].Incomming
			bs := []byte(incomming)
			for _, msg := range conn.FailStrings {
				if bytes.Contains(bs, msg) {
					err = errors.New(incomming)
					break
				}
			}
			if err != nil {
				return results, err
			}

		} else {
			if se, ok := err.(*scriptResult); ok {
				err = se.err

				results[len(results)-1].SubResults = se.results
			}
			return results, err
		}
	}
	return results, nil
}

func ParseScript(r io.Reader) (*Script, error) {
	scanner := bufio.NewScanner(r)

	_, script, err := parseScript(scanner, 0, false)
	return script, err
}

func parseScript(scanner *bufio.Scanner, line int, inBlock bool) (int, *Script, error) {

	start := line
	script := &Script{}

	for scanner.Scan() {
		line++
		bs := bytes.TrimSpace(scanner.Bytes())
		if len(bs) == 0 {
			continue
		}

		if bs[0] == '#' {
			continue
		}

		if bytes.HasSuffix(bs, []byte("}")) {
			if !inBlock {
				return line, nil, fmt.Errorf("%d: 非预期的块结束符 -- %s", line, bs)
			}

			if !bytes.Equal(bs, []byte("}")) {
				return line, nil, fmt.Errorf("%d: 块结束符必须独立一行 -- %s", line, bs)
			}
			return line, script, nil
		}

		if bytes.HasSuffix(bs, []byte("{")) {
			// NOTES: 必须 copy 一下， 因为 noPrefix 是 bs 的一部分
			// bs 是 Scanner 的， 在 parseScript 中读下一行时会被修改
			copyed := make([]byte, len(bs))
			copy(copyed, bs)

			newLine, subScript, err := parseScript(scanner, line, true)
			if err != nil {
				return newLine, script, err
			}

			args := bytes.TrimSuffix(copyed, []byte("{"))

			pos := bytes.IndexFunc(args, unicode.IsSpace)
			if pos < 0 {
				pos = len(args)
			}

			parse, ok := SubParsers[string(args[:pos])]
			if !ok {
				return line, nil, fmt.Errorf("%d: unknown command error -- %s", line, bs)
			}

			err = parse(script, line, newLine, string(copyed), bytes.TrimSpace(args[pos:]), subScript)
			if err != nil {
				return line, nil, fmt.Errorf("%d: %s -- %s", line, err, bs)
			}
			line = newLine
			continue
		}

		isOk := false
		for prefix, parse := range Parsers {
			prefixBytes := []byte(prefix)
			if bytes.HasPrefix(bs, prefixBytes) {
				if len(bs) != len(prefixBytes) && !unicode.IsSpace(rune(bs[len(prefixBytes)])) {
					continue
				}

				noPrefix := bytes.TrimSpace(bytes.TrimPrefix(bs, prefixBytes))

				// NOTES: 必须 copy 一下， 因为 noPrefix 是 bs 的一部分
				// bs 是 Scanner 的， 在读下一行时会被修改
				copyed := make([]byte, len(noPrefix))
				copy(copyed, noPrefix)

				if err := parse(script, line, string(bs), copyed); err != nil {
					return line, nil, fmt.Errorf("%d: %s -- %s", line, err, bs)
				}
				isOk = true
				break
			}
		}

		if !isOk {
			// NOTES: 必须 copy 一下， 因为 noPrefix 是 bs 的一部分
			// bs 是 Scanner 的， 在读下一行时会被修改
			copyed := make([]byte, len(bs))
			copy(copyed, bs)

			if err := defaultParse(script, line, copyed); err != nil {
				return line, nil, fmt.Errorf("%d: %s -- %s", line, err, bs)
			}
		}
	}

	if inBlock {
		return line, nil, fmt.Errorf("%d: 没有块结束符", start)
	}

	return line, script, scanner.Err()
}

// func parseRawScript(line int, copyed []byte, buf *bufio.Scanner) (*Script, error) {
// 	var raw bytes.Buffer
// 	for i := 1; i < line; i++ {
// 		raw.WriteString("\r\n")
// 	}

// 	raw.Write(copyed)
// 	raw.WriteString("\r\n")
// 	for buf.Scan() {
// 		line++
// 		raw.Write(buf.Bytes())
// 		raw.WriteString("\r\n")
// 	}

// 	if buf.Err() != nil {
// 		return nil, buf.Err()
// 	}

// 	return &Script{
// 		cmds: []Command{func(ctx context.Context, script *Script, conn *Shell) error {
// 			go func() {
// 				io.Copy(conn, &raw)
// 			}()
// 			_, err := io.Copy(ioutil.Discard, conn)
// 			return err
// 		}},
// 	}, nil
// }

func escapeBytes(bs []byte) []byte {
	var copyed = make([]byte, 0, len(bs))

	isEscaped := false
	for _, b := range bs {
		switch b {
		case '\\':
			if isEscaped {
				copyed = append(copyed, '\\')
			}
			isEscaped = !isEscaped
		case 'r':
			if isEscaped {
				copyed = append(copyed, '\r')
				isEscaped = false
			} else {
				copyed = append(copyed, 'r')
			}
		case 'n':
			if isEscaped {
				copyed = append(copyed, '\n')
				isEscaped = false
			} else {
				copyed = append(copyed, 'n')
			}
		case 't':
			if isEscaped {
				copyed = append(copyed, '\t')
				isEscaped = false
			} else {
				copyed = append(copyed, 't')
			}
		case 's':
			if isEscaped {
				copyed = append(copyed, ' ')
				isEscaped = false
			} else {
				copyed = append(copyed, 's')
			}
		default:
			if isEscaped {
				copyed = append(copyed, '\\')
			}
			isEscaped = false
			copyed = append(copyed, b)
		}
	}
	return copyed
}

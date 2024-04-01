package shell

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"unicode"
	"time"

	"github.com/runner-mei/errors"
)

const YesOrNo = "? [Y/N]:"

// MorePrompts 为 terminal 中 more 的各种格式
var MorePrompts = [][]byte{[]byte("- More -"),
	[]byte("-- More --"),
	[]byte("- more -"),
	[]byte("-- more --"),
	[]byte("-More-"),
	[]byte("--More--"),
	[]byte("-more-"),
	[]byte("--more--"),
	[]byte("-MORE-"),
	[]byte("--MORE--"),
	[]byte("- MORE -"),
	[]byte("-- MORE --"),
	[]byte("--More(CTRL+C break)--"),
	[]byte("-- More(CTRL+C break) --"),
	[]byte("-- More (CTRL+C break) --"),
	[]byte("--More (CTRL+C break)--"),
	[]byte("--more(CTRL+C break)--"),
	[]byte("-- more(CTRL+C break) --"),
	[]byte("-- more (CTRL+C break) --"),
	[]byte("--more (CTRL+C break)--"),
}

var (
	h3cSuperResponse  = []byte("User privilege level is")
	anonymousPassword = []byte("<<anonymous>>")
	nonePassword      = []byte("<<none>>")
	noneUsername      = []byte("<<none>>")
	emptyPassword     = []byte("<<empty>>")
	defaultEnableCmd  = []byte("enable")

	defaultUserPrompts = [][]byte{
		[]byte("Username:"),
		[]byte("username:"),
		[]byte("login:"),
		[]byte("Login:"),
		[]byte("login as:"),
		[]byte("Login as:"),
		[]byte("Login As:"),
		[]byte("login name:"),
		[]byte("Login Name:"),
	}
	defaultPasswordPrompts = [][]byte{[]byte("Password:"), []byte("password:")}
	defaultPrompts         = [][]byte{[]byte(">"), []byte("$"), []byte("#")}
	defaultErrorPrompts    = [][]byte{
		[]byte("Bad secrets"),
		[]byte("Login invalid"),
		[]byte("login invalid"),
		[]byte("Access denied"),
		[]byte("access denied"),
		[]byte("Login failed"),
		[]byte("Authorization fail"),
		[]byte("authorization fail"),
		[]byte("Authorizate fail"),
		[]byte("authorizate fail"),
		[]byte("Error:"),
		[]byte("found at '^' position"),
	}
	defaultPermissionPrompts = [][]byte{
		[]byte("Invalid input detected at '^' marker"),
		[]byte("Error: Too many parameters found at '^' position"),
		[]byte("Authorization failed"),
		[]byte("authorization failed"),
		[]byte("Authorizate fail"),
		[]byte("authorizate fail"),
		[]byte("Command authorization failed."),
		[]byte("Unrecognized command found"),
	}
)

var SayYesCRLF = func(conn Conn, bs []byte, idx int) (bool, error) {
	conn.Sendln([]byte("y"))
	return true, nil
}
var SayCRLF = func(conn Conn, bs []byte, idx int) (bool, error) {
	conn.Sendln([]byte(""))
	return true, nil
}
var SayNoCRLF = func(conn Conn, bs []byte, idx int) (bool, error) {
	conn.Sendln([]byte("N"))
	return true, nil
}
var SaySpace = func(conn Conn, bs []byte, idx int) (bool, error) {
	conn.Send([]byte(" "))
	return true, nil
}

var SayYes = func(conn Conn, bs []byte, idx int) (bool, error) {
	conn.Send([]byte("y"))
	return true, nil
}
var SayNo = func(conn Conn, bs []byte, idx int) (bool, error) {
	conn.Send([]byte("N"))
	return true, nil
}

var ReturnOK = func(conn Conn, bs []byte, idx int) (bool, error) {
	return false, nil
}

var ReturnErr = func(err error) func(conn Conn, bs []byte, idx int) (bool, error) {
	return func(conn Conn, bs []byte, idx int) (bool, error) {
		return false, err
	}
}

// 常见的提问
var (
	ChangeNow1Question     = Match("Change now? [Y/N]:", SayNoCRLF)
	ChangeNow2Question     = Match("Change now?[Y/N]:", SayNoCRLF)
	ChangePasswordQuestion = Match("change the password?", SayNoCRLF)
	StoreKeyInCache        = Match("Store key in cache? (y/n)", SayYes)
	ContinueWithConnection = Match("Continue with connection? (y/n)", SayYes)
	UpdateCachedKey        = Match("Update cached key? (y/n, Return cancels connection)", SayYes)
	More                   = Match(MorePrompts, SaySpace)

	DefaultMatchers = []Matcher{
		ChangeNow1Question,
		ChangeNow2Question,
		ChangePasswordQuestion,
		StoreKeyInCache,
		UpdateCachedKey,
		ContinueWithConnection,
		More,
	}
)

func init() {
	for _, prompt := range defaultPermissionPrompts {
		DefaultMatchers = append(DefaultMatchers,
			Match(prompt, ReturnErr(errors.WrapWithSuffix(errors.ErrPermission, string(prompt)))))
	}
}

func IsNoneUsername(username []byte) bool {
	return bytes.Equal(username, noneUsername)
}

func IsNonePassword(password []byte) bool {
	return bytes.Equal(password, nonePassword) || bytes.Equal(password, anonymousPassword)
}

func IsEmptyPassword(password []byte) bool {
	return bytes.Equal(password, emptyPassword)
}

type Matcher interface {
	Strings() []string
	Prompts() [][]byte
	Do() DoFunc
}

type stringMatcher struct {
	prompts []string
	do      DoFunc
}

func (s *stringMatcher) Strings() []string {
	return s.prompts
}

func (s *stringMatcher) Prompts() [][]byte {
	var prompts [][]byte
	for idx := range s.prompts {
		prompts = append(prompts, []byte(s.prompts[idx]))
	}
	return prompts
}

func (s *stringMatcher) Do() DoFunc {
	return s.do
}

type bytesMatcher struct {
	prompts [][]byte
	do      DoFunc
}

func (s *bytesMatcher) Strings() []string {
	var prompts []string
	for idx := range s.prompts {
		prompts = append(prompts, string(s.prompts[idx]))
	}
	return prompts
}

func (s *bytesMatcher) Prompts() [][]byte {
	return s.prompts
}

func (s *bytesMatcher) Do() DoFunc {
	return s.do
}

func Match(prompts interface{}, cb func(Conn, []byte, int) (bool, error)) Matcher {
	switch values := prompts.(type) {
	case []string:
		return &stringMatcher{
			prompts: values,
			do:      cb,
		}
	case [][]byte:
		return &bytesMatcher{
			prompts: values,
			do:      cb,
		}
	case []byte:
		return &bytesMatcher{
			prompts: [][]byte{values},
			do:      cb,
		}
	case string:
		return &bytesMatcher{
			prompts: [][]byte{[]byte(values)},
			do:      cb,
		}
	default:
		panic(fmt.Errorf("want []string or [][]byte got %T", prompts))
	}
}

const maxRetryCount = 1000

func Expect(ctx context.Context, conn Conn, matchs ...Matcher) error {
	var matchIdxs = make([]int, 0, len(matchs)+len(DefaultMatchers))
	var prompts = make([][]byte, 0, len(matchs)+len(DefaultMatchers))

	for idx := range matchs {
		matchIdxs = append(matchIdxs, len(prompts))
		prompts = append(prompts, matchs[idx].Prompts()...)
	}
	for idx := range DefaultMatchers {
		matchIdxs = append(matchIdxs, len(prompts))
		prompts = append(prompts, DefaultMatchers[idx].Prompts()...)
	}

	more := false
	for retryCount := 0; retryCount < maxRetryCount; retryCount++ {
		idx, recvBytes, err := conn.Expect(prompts)
		if err != nil {
			if bytes.Contains(recvBytes, []byte("Network error:")) {
				if bytes.Contains(recvBytes, []byte("Connection timed out")) {
					return &net.OpError{Op: "dial",
						Net: "tcp",
						Err: net.UnknownNetworkError(string(recvBytes))}
				}
				return errors.New(string(recvBytes))
			}
			err = errors.Wrap(err, "read util '"+string(bytes.Join(prompts, []byte(",")))+"' failed")
			return errors.WrapWithSuffix(err, "\r\n"+ToHexStringIfNeed(recvBytes))
		}

		foundMatchIndex := -1

		for i := 0; i < len(matchIdxs); i++ {
			if matchIdxs[i] <= idx && (i == len(matchIdxs)-1 || idx < matchIdxs[i+1]) {
				foundMatchIndex = i
				break
			}
		}

		if foundMatchIndex < 0 {
			return errors.New("read util '" + string(bytes.Join(prompts, []byte(","))) + "' failed, return index is '" + strconv.Itoa(idx) + "'")
		}

		var cb DoFunc
		if len(matchs) > foundMatchIndex {
			cb = matchs[foundMatchIndex].Do()
		} else {
			cb = DefaultMatchers[foundMatchIndex-len(matchs)].Do()
		}
		more, err = cb(conn, recvBytes, idx-matchIdxs[foundMatchIndex])
		if err != nil {
			return err
		}
		if !more {
			return nil
		}
	}

	return errors.New("read util '" + string(bytes.Join(prompts, []byte(","))) + "' failed, retry count > " + strconv.FormatInt(maxRetryCount, 10))
}

func UserLogin(ctx context.Context, conn Conn, userPrompts [][]byte, username []byte, passwordPrompts [][]byte, password []byte, prompts [][]byte, matchs ...Matcher) ([]byte, error) {
	if len(userPrompts) == 0 {
		userPrompts = defaultUserPrompts
	}
	if len(passwordPrompts) == 0 {
		passwordPrompts = defaultPasswordPrompts
	}
	if len(prompts) == 0 {
		prompts = defaultPrompts
	}

	var buf bytes.Buffer
	cancel := conn.SetTeeOutput(&buf)
	defer cancel()

	status := 0

	copyed := make([]Matcher, len(matchs)+5)
	copyed[0] = Match(userPrompts, func(c Conn, bs []byte, nidx int) (bool, error) {
		if e := conn.Sendln(username); e != nil {
			return false, errors.Wrap(e, "send username failed")
		}
		status = 1
		return false, nil
	})
	copyed[1] = Match(passwordPrompts, func(c Conn, bs []byte, nidx int) (bool, error) {
		if IsEmptyPassword(password) {
			password = []byte{}
		}
		if e := conn.SendPassword(password); e != nil {
			return false, errors.Wrap(e, "send user password failed")
		}

		status = 2
		return false, nil
	})
	copyed[2] = Match(prompts, func(c Conn, bs []byte, nidx int) (bool, error) {
		status = 3
		return false, nil
	})

	copyed[3] = Match(defaultErrorPrompts, func(c Conn, bs []byte, nidx int) (bool, error) {
		status = 4
		return false, nil
	})

	copyed[4] = Match(defaultPermissionPrompts, func(c Conn, bs []byte, nidx int) (bool, error) {
		status = 5
		return false, nil
	})

	copy(copyed[5:], matchs)

	for i := 0; ; i++ {
		if i >= 10 {
			return nil, errors.New("user login fail:\r\n" + ToHexStringIfNeed(buf.Bytes()))
		}
		err := Expect(ctx, conn, copyed...)
		if err != nil {
			return nil, errors.Wrap(err, "user login fail")
		}

		if status == 3 {
			if _, err := conn.DrainOff(1 * time.Second); err != nil {
				return nil, errors.New("read prompt failed, drain off, " + err.Error())
			}

			received := buf.Bytes()
			if len(received) == 0 {
				return nil, errors.New("read prompt failed, received is empty")
			}

			prompt := GetPrompt(received, prompts)
			if len(prompt) == 0 {
				return nil, errors.New("read prompt '" + string(bytes.Join(prompts, []byte(","))) + "' failed: \r\n" + ToHexStringIfNeed(received))
			}
			return prompt, nil
		}

		if status == 4 {
			received := buf.Bytes()
			if len(received) == 0 {
				return nil, errors.New("invalid password")
			}

			return nil, errors.New("invalid password: \r\n" + ToHexStringIfNeed(received))
		}

		if status == 5 {
			received := buf.Bytes()
			if len(received) == 0 {
				return nil, errors.ErrPermission
			}

			return nil, errors.WrapWithSuffix(errors.ErrPermission, "\r\n"+ToHexStringIfNeed(received))
		}
	}
}

func ReadPrompt(ctx context.Context, conn Conn, prompts [][]byte, matchs ...Matcher) ([]byte, error) {
	var buf bytes.Buffer
	cancel := conn.SetTeeOutput(&buf)
	defer cancel()

	return readPrompt(ctx, &buf, conn, prompts, matchs...)
}

func readPrompt(ctx context.Context, buf *bytes.Buffer, conn Conn, prompts [][]byte, matchs ...Matcher) ([]byte, error) {
	if len(prompts) == 0 {
		prompts = defaultPrompts
	}

	isPrompt := false

	copyed := make([]Matcher, len(matchs)+1)
	copyed[0] = Match(prompts, func(conn Conn, bs []byte, idx int) (bool, error) {
		isPrompt = true
		return false, nil
	})
	copy(copyed[1:], matchs)

	for retryCount := 0; ; retryCount++ {
		if retryCount >= 10 {
			return nil, errors.New("read prompt failed, retry count > 10")
		}
		e := Expect(ctx, conn, copyed...)
		if nil != e {
			return nil, e
		}

		if isPrompt {
			break
		}
	}

	if _, err := conn.DrainOff(1 * time.Second); err != nil {
		return nil, errors.New("read prompt failed, drain off, " + err.Error())
	}

	received := buf.Bytes()
	if len(received) == 0 {
		return nil, errors.New("read prompt failed, received is empty")
	}

	prompt := GetPrompt(received, prompts)
	if len(prompt) == 0 {
		return nil, errors.New("read prompt '" + string(bytes.Join(prompts, []byte(","))) + "' failed: \r\n" + ToHexStringIfNeed(received))
	}
	return prompt, nil
}

func GetPrompt(bs []byte, prompts [][]byte) []byte {
	if len(bs) == 0 {
		return nil
	}

	lines := SplitLines(bs)

	var fullPrompt []byte
	for i := len(lines) - 1; i >= 0; i-- {
		fullPrompt = bytes.TrimFunc(lines[i], func(r rune) bool {
			if r == 0 {
				return true
			}
			return unicode.IsSpace(r)
		})
		if len(fullPrompt) > 0 {
				for _, prompt := range prompts {
					if bytes.HasSuffix(fullPrompt, prompt) {
						if 2 <= len(fullPrompt) && ']' == fullPrompt[len(fullPrompt)-2] {
							idx := bytes.LastIndex(fullPrompt, []byte("["))
							if idx > 0 {
								fullPrompt = fullPrompt[idx:]
							}
						}
						return fullPrompt
					}
				}
		}
	}

	// fmt.Println("===", string(fullPrompt))



	return nil
}

func Exec(ctx context.Context, conn Conn, prompt, cmd []byte) ([]byte, error) {
	if len(prompt) == 0 {
		return nil, errors.New("prompt is missing")
	}
	if len(cmd) == 0 {
		return nil, errors.New("cmd is missing")
	}

	if bytes.HasPrefix(prompt, []byte("\\n")) {
		prompt[1] = '\n'
		prompt = prompt[1:]
	}

	var buf bytes.Buffer
	cancel := conn.SetTeeOutput(&buf)
	defer cancel()

	err := conn.Sendln(cmd)
	if err != nil {
		return nil, err
	}

	err = Expect(ctx, conn, Match(prompt, func(Conn, []byte, int) (bool, error) {
		return false, nil
	}))

	if err != nil {
		return nil, err
	}
	bs := buf.Bytes()
	bs = bs[:len(bs)-len(prompt)]

	for _, prompt := range defaultPermissionPrompts {
		if bytes.Contains(bs, prompt) {
			return nil, errors.WrapWithSuffix(errors.ErrPermission, string(prompt))
		}
	}
	return bs, nil
}

func WithEnable(ctx context.Context, conn Conn, enableCmd []byte, passwordPrompts [][]byte, password []byte, enablePrompts [][]byte) ([]byte, error) {
	if len(enableCmd) == 0 {
		enableCmd = defaultEnableCmd
	}

	if e := conn.Sendln(enableCmd); nil != e {
		return nil, errors.Wrap(e, "send enable '"+string(enableCmd)+"' failed")
	}

	if len(passwordPrompts) == 0 {
		passwordPrompts = defaultPasswordPrompts
	}
	if len(enablePrompts) == 0 {
		enablePrompts = defaultPrompts
	}

	// fmt.Println("send enable '" + string(enableCmd) + "' ok and read enable password prompt")

	var buf bytes.Buffer
	cancel := conn.SetTeeOutput(&buf)
	defer cancel()

	if !IsNonePassword(password) {

		var isPrompt bool
		err := Expect(ctx, conn,
			Match(enablePrompts, func(c Conn, bs []byte, nidx int) (bool, error) {
				isPrompt = true
				return false, nil
			}),
			Match(append(passwordPrompts, h3cSuperResponse), func(c Conn, bs []byte, nidx int) (bool, error) {
				if IsEmptyPassword(password) {
					password = []byte{}
				}
				if e := conn.SendPassword(password); e != nil {
					return false, errors.Wrap(e, "send enable password failed")
				}
				return false, nil
			}))
		if err != nil {
			return nil, err
		}

		if isPrompt {
			if _, err := conn.DrainOff(5 * time.Second); err != nil {
				return nil, errors.New("read prompt failed, drain off, " + err.Error())
			}

			output := buf.Bytes()
			if len(output) == 0 {
				return nil, errors.New("read prompt failed, received is empty")
			}

			prompt := GetPrompt(output, enablePrompts)
			if len(prompt) == 0 {
				return nil, errors.New("read prompt '" + string(bytes.Join(enablePrompts, []byte(","))) + "' failed: \r\n" + ToHexStringIfNeed(output))
			}
			return prompt, nil
		}
	}
	return readPrompt(ctx, &buf, conn, enablePrompts)
}

func WithView(ctx context.Context, conn Conn, cmd []byte, newPrompts [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	cancel := conn.SetTeeOutput(&buf)
	defer cancel()

	if e := conn.Sendln(cmd); nil != e {
		return nil, errors.Wrap(e, "send '"+string(cmd)+"' failed")
	}

	err := Expect(ctx, conn,
		Match(newPrompts, func(c Conn, bs []byte, nidx int) (bool, error) {
			return false, nil
		}),
	)
	if err != nil {
		return nil, err
	}

	if _, err := conn.DrainOff(5 * time.Second); err != nil {
		return nil, errors.New("read prompt failed, drain off, " + err.Error())
	}

	output := buf.Bytes()
	if len(output) == 0 {
		return nil, errors.New("read prompt failed, received is empty")
	}


	prompt := GetPrompt(output, newPrompts)
	if len(prompt) == 0 {
		return nil, errors.New("read prompt '" + string(bytes.Join(newPrompts, []byte(","))) + "' failed: \r\n" + ToHexStringIfNeed(output))
	}
	return prompt, nil
}

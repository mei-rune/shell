package shell

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	// "github.com/mei-rune/shell/sim/sshd"
	// "github.com/mei-rune/shell/sim/telnetd"
	"tech.hengwei.com.cn/go/private/sim/sshd"
	"tech.hengwei.com.cn/go/private/sim/telnetd"
)

func TestTelnetSimple(t *testing.T) {
	t.Skip("已有模拟的，跳过这个测试")

	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("192.168.1.172", "23"), 10*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)

	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("admin"), nil, []byte("admin"), nil)
	if err != nil {
		t.Error(err)
		return
	}

	if string(prompt) != "Switch>" {
		t.Errorf("want 'Switch>' got %s", prompt)
		return
	}

	output, err := Exec(ctx, conn, prompt, []byte("echo abcd"))
	if err != nil {
		if !strings.Contains(err.Error(), "Invalid input detected at '^' marker") {
			t.Error(err)
			return
		}
	} else if !strings.Contains(string(output), "print abcd") {
		t.Errorf("want 'print abcd' got %s", output)
	}
}

func TestTelnetSimSimple(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "","abc#", sshd.Echo)
	options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)

	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil)
	if err != nil {
		t.Error(err)
		return
	}

	testSimSimple(t, ctx, conn, prompt)
}


func TestTelnetSimPrompt(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "","abc#", sshd.Echo)
	options.WithNoEnable("ABC>>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)

	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(prompt))
	if string(prompt) != "ABC>>" {
		t.Error("want ABC>> got", string(prompt))
	}

	output, err := Exec(ctx, conn, prompt, []byte("echo ABC>abcd"))
	if err != nil {
		t.Error(err)
	} else if !strings.Contains(string(output), "print ABC>abcd") {
		t.Errorf("want 'print ABC>abcd' got %s", output)
	}

	// testSimSimple(t, ctx, conn, prompt)
}

func TestTelnetSimWithNoUserNoPassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("<<none>>", "<<none>>")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "","abc#", sshd.Echo)
	options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)

	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil)
	if err != nil {
		t.Error(err)
		return
	}

	testSimSimple(t, ctx, conn, prompt)
}

func TestTelnetSimSimpleWithLastLogin(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")
	options.Welcome = []byte("Last login: Tue Apr 7 10:11:21 from hostnamefortest 192.168.1.98\r\n")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "","abc#", sshd.Echo)
	options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)

	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil)
	if err != nil {
		t.Error(err)
		return
	}

	output := testSimSimple(t, ctx, conn, prompt)
	if len(output) == 0 {
		t.Error("output is empty")
	}
	if bytes.Contains(output, []byte("hostnamefortest")) {
		t.Error("want not hostnamefortest, but got", string(output))
	}
}

func TestTelnetSimWithEnable(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)
	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "testsx")
}

func TestTelnetSimWithNoUserNoPasswordWithEnable(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("<<none>>", "<<none>>")

	options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)
	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "testsx")
}

func TestTelnetSimWithEnableNonePassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "<<none>>", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}
	testSimWithEnable(t, ctx, conn, prompt, "enable", "<<none>>")
}

func TestTelnetSimWithEnableNonePasswordButUserInputEnPassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "<<none>>", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}
	testSimWithEnable(t, ctx, conn, prompt, "enable", "not_exist_password")
}

func TestTelnetSimWithEnableEmptyPassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "<<empty>>", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "<<empty>>")
}

func TestTelnetSimWithYesNo(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "","abc#", sshd.Echo)
	options.WithQuest("abc? [Y/N]:", "N", "ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimSimple(t, ctx, conn, prompt)
}

func TestTelnetSimWithEnableWithYesNo(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithQuest("abc? [Y/N]:", "N", "ABC>",
		sshd.WithEnable("enable", "password:", "testsx", "", "abc#", sshd.Echo))
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	telnetConn, err := DialTelnetTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 1*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	conn := TelnetWrap(telnetConn, nil, nil)
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "testsx")
}

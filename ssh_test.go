package shell

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	// "github.com/mei-rune/shell/sim/sshd"
	"tech.hengwei.com.cn/go/private/sim/sshd"
)

var answerNo = Match("abc? [Y/N]:", SayNoCRLF)
var AbcQuestion = answerNo

func TestSSHSimSimple1(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
	options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	//  conn, err := ConnectPlink(net.JoinHostPort("127.0.0.1", port), "abc", "123", "")
	conn, err := ConnectSSH(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := ReadPrompt(ctx, conn, [][]byte{[]byte(">")}, answerNo)
	if err != nil {
		t.Error(err)
		return
	}
	testSimSimple(t, ctx, conn, prompt)
}

func TestSSHSimSimplePrompt(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
	options.WithNoEnable("ABC>>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	//  conn, err := ConnectPlink(net.JoinHostPort("127.0.0.1", port), "abc", "123", "")
	conn, err := ConnectSSH(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := ReadPrompt(ctx, conn, [][]byte{[]byte(">")}, answerNo)
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
}


func TestSSHSimSimple2(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
	options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	//  conn, err := ConnectPlink(net.JoinHostPort("127.0.0.1", port), "abc", "123", "")
	conn, err := ConnectSSH(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
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

func TestSSHSimWithEnable(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	conn, err := ConnectSSH(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := ReadPrompt(ctx, conn, [][]byte{[]byte(">")}, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "testsx")
}


func TestSSHSimWithEnable2(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>>", "enable", "password:", "testsx", "", "abc##", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	conn, err := ConnectSSH(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := ReadPrompt(ctx, conn, [][]byte{[]byte(">")}, answerNo)
	if err != nil {
		t.Error(err)
		return
	}


	t.Log(string(prompt))
	if string(prompt) != "ABC>>" {
		t.Error("want ABC>> got", string(prompt))
	}

	// output, err := Exec(ctx, conn, prompt, []byte("echo ABC>abcd"))
	// if err != nil {
	// 	t.Error(err)
	// } else if !strings.Contains(string(output), "print ABC>abcd") {
	// 	t.Errorf("want 'print ABC>abcd' got %s", output)
	// }

	prompt, err = WithEnable(ctx, conn, []byte("enable"), nil, []byte("testsx"), nil)
	if err != nil {
		t.Error(err)
		return
	}
	if string(prompt) != "abc##" {
		t.Errorf("want 'abc##' got %s", prompt)
		return
	}

	output, err := Exec(ctx, conn, prompt, []byte("echo abc#abcd"))
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(string(output), "print abc#abcd") {
		t.Errorf("want 'print abc#abcd' got %s", output)
	}
}

func TestSSHSimWithEnableNonePassword(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "<<none>>", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	conn, err := ConnectSSH(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := ReadPrompt(ctx, conn, [][]byte{[]byte(">")}, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "<<none>>")
}

func TestSSHSimWithEnableEmptyPassword(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "<<empty>>", "", "abc#", sshd.Echo)
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	conn, err := ConnectSSH(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := ReadPrompt(ctx, conn, [][]byte{[]byte(">")}, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "<<empty>>")
}

func TestSSHSimWithYesNo(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "","abc#", sshd.Echo)
	options.WithQuest("abc? [Y/N]:", "N", "ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	conn, err := ConnectSSH(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := ReadPrompt(ctx, conn, [][]byte{[]byte(">")}, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimSimple(t, ctx, conn, prompt)
}

func TestSSHSimWithEnableWithYesNo(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithQuest("abc? [Y/N]:", "N", "ABC>",
		sshd.WithEnable("enable", "password:", "testsx", "", "abc#", sshd.Echo))
	//options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	conn, err := ConnectSSH(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(1 * time.Second)

	prompt, err := ReadPrompt(ctx, conn, [][]byte{[]byte(">")}, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "testsx")
}

func testSimSimple(t *testing.T, ctx context.Context, conn Conn, prompt []byte) []byte {
	if string(prompt) != "ABC>" {
		t.Errorf("want 'ABC>' got %s", prompt)
		return nil
	}

	output, err := Exec(ctx, conn, prompt, []byte("echo abcd"))
	if err != nil {
		t.Error(err)
		return nil
	}
	if !strings.Contains(string(output), "print abcd") {
		t.Errorf("want 'print abcd' got %s", output)
	}
	return output
}

func testSimWithEnable(t *testing.T, ctx context.Context, conn Conn, prompt []byte, enableCmd, enablePwd string) {
	var err error
	if string(prompt) != "ABC>" {
		t.Errorf("want 'ABC>' got %s", prompt)
		return
	}

	prompt, err = WithEnable(ctx, conn, []byte(enableCmd), [][]byte{[]byte("password:")}, []byte(enablePwd), [][]byte{[]byte("abc#")})
	if err != nil {
		t.Error(err)
		return
	}
	if string(prompt) != "abc#" {
		t.Errorf("want 'abc#' got %s", prompt)
		return
	}

	output, err := Exec(ctx, conn, prompt, []byte("echo abcd"))
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(string(output), "print abcd") {
		t.Errorf("want 'print abcd' got %s", output)
	}
}

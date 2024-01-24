package shell

import (
	"bytes"
	"context"
	"net"
	_ "net/http/pprof"
	"testing"
	"time"

	// "github.com/mei-rune/shell/sim/sshd"
	"tech.hengwei.com.cn/go/private/sim/sshd"
)

func TestPlinkSimSimple(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "","abc#", sshd.Echo)
	options.WithNoEnable("ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	var buf bytes.Buffer

	out := &SafeWriter{
		W: &buf,
	}
	defer func() {
		t.Log(buf.String())

		// time.Sleep(1 * time.Hour)
	}()

	conn, err := ConnectPlink(net.JoinHostPort("127.0.0.1", port), "abc", "123", "",
		PasswordWriter(WriteFunc(func(p []byte) (int, error) {
			return out.WriteWithTag("S:", p)
		})),
		PasswordWriter(WriteFunc(func(p []byte) (int, error) {
			return out.WriteWithTag("C:", p)
		})))
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(5 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}
	testSimSimple(t, ctx, conn, prompt)
}

func TestPlinkSimWithEnable(t *testing.T) {
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

	conn, err := ConnectPlink(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(5 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "testsx")
}

func TestPlinkSimWithEnableNonePassword(t *testing.T) {
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

	conn, err := ConnectPlink(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	conn.UseCRLF()
	conn.SetReadDeadline(5 * time.Second)

	prompt, err := UserLogin(ctx, conn, nil, []byte("abc"), nil, []byte("123"), nil, answerNo)
	if err != nil {
		t.Error(err)
		return
	}

	testSimWithEnable(t, ctx, conn, prompt, "enable", "<<none>>")
}

func TestPlinkSimWithEnableEmptyPassword(t *testing.T) {
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

	var buf bytes.Buffer
	out := safeIO(&buf)
	defer func() {
		t.Log(buf.String())
	}()

	conn, err := ConnectPlink(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", out, out)
	if err != nil {
		t.Error(err)
		return
	}
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

func TestPlinkSimWithYesNo(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "abc#", sshd.Echo)
	options.WithQuest("abc? [Y/N]:", "N", "ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	conn, err := ConnectPlink(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
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

func TestPlinkSimWithEnableWithYesNo(t *testing.T) {
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

	conn, err := ConnectPlink(net.JoinHostPort("127.0.0.1", port), "abc", "123", "", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
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

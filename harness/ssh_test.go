package harness

import (
	"bytes"
	"context"
	"strings"
	"testing"

	// "github.com/mei-rune/shell/sim/sshd"
	"tech.hengwei.com.cn/go/private/sim/sshd"

	"github.com/mei-rune/shell"
)

var AbcQuestion = shell.Match("abc? [Y/N]:", shell.SayNoCRLF)

func TestSSHSimSimple(t *testing.T) {
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

	params := &SSHParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		PrivateKey:          "",
		Prompt:              "",
		EnableCommand:       "",
		EnablePasswordQuest: "",
		EnablePassword:      "",
		EnablePrompt:        "",
		UseExternalSSH:      false,
		UseCRLF:             true,
	}
	testSSH(t, ctx, params)

	params.UseExternalSSH = true
	testSSH(t, ctx, params)
}

func TestSSHSimWithEnablePassword(t *testing.T) {
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

	params := &SSHParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		PrivateKey:          "",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "testsx",
		EnablePrompt:        "",
		UseExternalSSH:      false,
		UseCRLF:             true,
	}
	testSSH(t, ctx, params)

	params.UseExternalSSH = true
	testSSH(t, ctx, params)
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

	params := &SSHParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		PrivateKey:          "",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "<<none>>",
		EnablePrompt:        "",
		UseExternalSSH:      false,
		UseCRLF:             true,
	}
	testSSH(t, ctx, params)

	params.UseExternalSSH = true
	testSSH(t, ctx, params)
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

	params := &SSHParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		PrivateKey:          "",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "<<empty>>",
		EnablePrompt:        "",
		UseExternalSSH:      false,
		UseCRLF:             true,
	}
	testSSH(t, ctx, params)

	params.UseExternalSSH = true
	testSSH(t, ctx, params)
}

func TestSSHSimWithYesNo(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
	options.WithQuest("abc? [Y/N]:", "N", "ABC>", sshd.Echo)

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	params := &SSHParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		PrivateKey:          "",
		Prompt:              "",
		EnableCommand:       "",
		EnablePasswordQuest: "",
		EnablePassword:      "",
		EnablePrompt:        "",
		UseExternalSSH:      false,
		UseCRLF:             true,
	}
	testSSH(t, ctx, params)

	params.UseExternalSSH = true
	testSSH(t, ctx, params)
}

func TestSSHSimWithEnableWithYesNo(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", sshd.Echo)
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

	params := &SSHParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		PrivateKey:          "",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "testsx",
		EnablePrompt:        "",
		UseExternalSSH:      false,
		UseCRLF:             true,
	}
	testSSH(t, ctx, params)

	params.UseExternalSSH = true
	testSSH(t, ctx, params)
}

func testSSH(t *testing.T, ctx context.Context, params *SSHParam) {
	var buf bytes.Buffer
	c, prompt, err := DailSSH(ctx, params, ServerWriter(&buf), ClientWriter(&buf), Question(AbcQuestion.Prompts(), AbcQuestion.Do()))
	if err != nil {
		t.Error(err)
		return
	}
	conn := &Shell{Conn: c, Prompt: prompt}
	defer conn.Close()

	result, err := Exec(ctx, conn, "echo abcd")
	if err != nil {
		t.Error(err)
		return
	}

	if !strings.Contains(result.Incomming, "print abcd") {
		t.Errorf("want 'print abcd' got %s", result.Incomming)
	}
	t.Log(result.Incomming)
	t.Log(buf.String())
}

func TestSSHSimMore(t *testing.T) {
	options := &sshd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx","", "abc#", sshd.Echo)
	options.WithNoEnable("ABC>", sshd.OS(sshd.Commands{
		"show": sshd.WithMore([]string{
			"abcd",
			"efgh",
			"ijklmn",
		}, []byte("-- more --"), nil),
	}))

	listener, err := sshd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()

	ctx := context.Background()

	params := &SSHParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		PrivateKey:          "",
		Prompt:              "",
		EnableCommand:       "",
		EnablePasswordQuest: "",
		EnablePassword:      "",
		EnablePrompt:        "",
		UseExternalSSH:      false,
		UseCRLF:             true,
	}
	testSSHMore(t, ctx, params)

	params.UseExternalSSH = true
	testSSHMore(t, ctx, params)
}

func testSSHMore(t *testing.T, ctx context.Context, params *SSHParam) {
	var buf bytes.Buffer
	c, prompt, err := DailSSH(ctx, params, ServerWriter(&buf), ClientWriter(&buf), Question(AbcQuestion.Prompts(), AbcQuestion.Do()))

	if err != nil {
		t.Error(err)
		return
	}
	conn := &Shell{Conn: c, Prompt: prompt}
	defer conn.Close()

	result, err := Exec(ctx, conn, "show")
	if err != nil {
		t.Error(err)
		return
	}

	if !strings.Contains(result.Incomming, "abcd") {
		t.Errorf("want 'abcd' got %s", result.Incomming)
	}

	if !strings.Contains(result.Incomming, "efgh") {
		t.Errorf("want 'efgh' got %s", result.Incomming)
	}

	if !strings.Contains(result.Incomming, "ijklmn") {
		t.Errorf("want 'ijklmn' got %s", result.Incomming)
	}
	t.Log(result.Incomming)
	t.Log(buf.String())
}

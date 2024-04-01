package harness

import (
	"bytes"
	"context"
	"strings"
	"testing"

	// "github.com/mei-rune/shell/sim/telnetd"
	"tech.hengwei.com.cn/go/private/sim/telnetd"
)

func TestTelnetSimSimple(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "abc#", telnetd.Echo)
	options.WithNoEnable("ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()

	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		Prompt:              "",
		EnableCommand:       "",
		EnablePasswordQuest: "",
		EnablePassword:      "",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)
}

func TestTelnetSimWithNoUserNoPassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("<<none>>", "<<none>>")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "abc#", telnetd.Echo)
	options.WithNoEnable("ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()

	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		Prompt:              "",
		EnableCommand:       "",
		EnablePasswordQuest: "",
		EnablePassword:      "",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)

	params = &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "<<none>>",
		// PasswordQuest: "",
		Password:            "<<none>>",
		Prompt:              "",
		EnableCommand:       "",
		EnablePasswordQuest: "",
		EnablePassword:      "",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)
}

func TestTelnetSimWithEnablePassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", telnetd.Echo)
	//options.WithNoEnable("ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "testsx",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)
}

func TestTelnetSimWithNoUserNoPasswordWithEnablePassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("<<none>>", "<<none>>")

	options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", telnetd.Echo)
	//options.WithNoEnable("ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "testsx",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)

	params = &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "<<none>>",
		// PasswordQuest: "",
		Password:            "<<none>>",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "testsx",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)
}

func TestTelnetSimWithEnableNonePassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "<<none>>", "", "abc#", telnetd.Echo)
	//options.WithNoEnable("ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "<<none>>",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)
}

func TestTelnetSimWithEnableEmptyPassword(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithEnable("ABC>", "enable", "password:", "<<empty>>", "", "abc#", telnetd.Echo)
	//options.WithNoEnable("ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "<<empty>>",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)
}

func TestTelnetSimWithYesNo(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	//options.WithEnable("ABC>", "enable", "password:", "testsx","", "abc#", telnetd.Echo)
	options.WithQuest("abc? [Y/N]:", "N", "ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		Prompt:              "",
		EnableCommand:       "",
		EnablePasswordQuest: "",
		EnablePassword:      "",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)
}

func TestTelnetSimWithEnableWithYesNo(t *testing.T) {
	options := &telnetd.Options{}
	options.AddUserPassword("abc", "123")

	options.WithQuest("abc? [Y/N]:", "N", "ABC>",
		telnetd.WithEnable("enable", "password:", "testsx", "", "abc#", telnetd.Echo))
	//options.WithNoEnable("ABC>", telnetd.Echo)

	listener, err := telnetd.StartServer(":", options)
	if err != nil {
		t.Error(err)
		return
	}
	defer listener.Close()

	port := listener.Port()
	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "127.0.0.1",
		Port:    port,
		// UserQuest: "",
		Username: "abc",
		// PasswordQuest: "",
		Password:            "123",
		Prompt:              "",
		EnableCommand:       "enable",
		EnablePasswordQuest: "",
		EnablePassword:      "testsx",
		EnablePrompt:        "",
		UseCRLF:             true,
	}
	testTelnet(t, ctx, params)
}

func testTelnet(t *testing.T, ctx context.Context, params *TelnetParam) {
	var buf bytes.Buffer
	c, prompt, err := DailTelnet(ctx, params, ServerWriter(&buf), ClientWriter(&buf), Question(AbcQuestion.Prompts(), AbcQuestion.Do()))
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

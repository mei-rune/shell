package harness

import (
	"bytes"
	"context"
	"fmt"
	_ "net/http/pprof"
	"strings"
	"testing"

	// "github.com/mei-rune/shell/sim/telnetd"
	"tech.hengwei.com.cn/go/private/sim/telnetd"

	"github.com/mei-rune/shell"
)

var ciscoConfigurations = []string{
	`Using 2157 out of 393216 bytes
!
version 12.1
no service pad
service timestamps debug uptime
service timestamps log uptime
no service password-encryption
!
hostname Switch
!
enable password admin
!
username admin password 0 admin
ip subnet-zero
!
!
spanning-tree mode pvst
spanning-tree extend system-id
!
!
!
!
interface FastEthernet0/1
 no switchport
 no ip address
!
interface FastEthernet0/2
 switchport mode dynamic desirable
!
interface FastEthernet0/8
 switchport mode dynamic desirable`,
	`interface FastEthernet0/9
 switchport mode dynamic desirable
!
interface FastEthernet0/20
 switchport mode dynamic desirable
!
interface FastEthernet0/23
 switchport mode dynamic desirable
!`,
	`interface FastEthernet0/24
 switchport mode dynamic desirable
!
interface GigabitEthernet0/1
 switchport mode dynamic desirable
!
interface GigabitEthernet0/2
 switchport mode dynamic desirable
!
interface Vlan1
 ip address 192.168.1.172 255.255.255.0
!
ip classless
ip http server
!
snmp-server community public RO
snmp-server community private RW
!
line con 0
line vty 0 4
 login local
line vty 5 15
 login
!
!
end
`,
}

func TestCisco(t *testing.T) {

	// go http.ListenAndServe(":12445", nil)

	moreAfter := append(append(bytes.Repeat([]byte{byte('\b')}, 9), []byte("        ")...), bytes.Repeat([]byte{byte('\b')}, 9)...)

	welcome := []byte{0xFF, 0xFB, 0x01, 0xFF, 0xFB, 0x03, 0xFF, 0xFD, 0x18, 0xFF, 0xFD, 0x1F, 0x0D, 0x0A, 0x0D, 0x0A}
	welcome = append(welcome, []byte("*****************\r\n   <myuser> <mypassword> \r\nUser Access Verification\r\n\r\n")...)

	options := &telnetd.Options{}
	options.SetWelcome(welcome)
	options.SetUserQuest(append([]byte("Username: "), 0xFF, 0xFA, 0x18, 0x01, 0xFF, 0xF0), []byte("Password:"))
	options.SetUserPassword("admin1", "admin2")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", telnetd.Echo)
	options.WithPrompt([]byte("Switch>"),
		telnetd.WithEnable("enable", "Password: ", "admin", "", "Switch#", telnetd.OS(telnetd.Commands{
			"show": telnetd.WithCommands(telnetd.Commands{
				"configuration": telnetd.WithMore(ciscoConfigurations, []byte(" --More--"), moreAfter),
			}),
		})))

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
		Username: "admin1",
		// PasswordQuest: "",
		Password:            "admin2",
		Prompt:              "",
		EnableCommand:       "",
		EnablePasswordQuest: "",
		EnablePassword:      "admin",
		EnablePrompt:        "",
		UseCRLF:             true,
	}

	testTelnetCisco(t, ctx, params)
}

func testTelnetCisco(t *testing.T, ctx context.Context, params *TelnetParam) {
	var buf bytes.Buffer
	c, prompt, err := DailTelnet(ctx, params, ServerWriter(&buf), ClientWriter(&buf), Question(AbcQuestion.Prompts(), AbcQuestion.Do()))

	if err != nil {
		t.Error(err)
		// t.Error(buf.Len(), buf.String())

		s := shell.ToHexStringIfNeed(buf.Bytes())
		t.Log(s)
		fmt.Println(s)
		return
	}

	conn := &Shell{Conn: c, Prompt: prompt}
	defer conn.Close()

	result, err := Exec(ctx, conn, "show configuration")
	if err != nil {
		t.Error(err)
		return
	}

	if !strings.Contains(result.Incomming, "snmp-server community private RW") {
		t.Errorf("want 'print abcd' got %s", result.Incomming)
	}
	t.Log(result.Incomming)
	t.Log(buf.String())
}

func TestCiscoFail1(t *testing.T) {
	t.Skip("这个是实际设备，跳过, 开启动请修改正确的用户名和密码")
	ctx := context.Background()

	params := &TelnetParam{
		// Timeout: 30 * time.Second,
		Address: "192.168.1.173",
		Port:    "23",
		// UserQuest: "",
		Username: "admin",
		// PasswordQuest: "",
		Password:            "admin",
		Prompt:              "",
		EnableCommand:       "en",
		EnablePasswordQuest: "",
		EnablePassword:      "123456",
		EnablePrompt:        "",
		UseCRLF:             true,
	}

	var buf bytes.Buffer
	c, _, err := DailTelnet(ctx, params, ServerWriter(&buf), ClientWriter(&buf), Question(AbcQuestion.Prompts(), AbcQuestion.Do()))

	if err == nil {
		defer c.Close()

		t.Error("want error go ok")

		s := shell.ToHexStringIfNeed(buf.Bytes())
		t.Error(s)
		fmt.Println(s)
		return
	}

	if !strings.Contains(err.Error(), "invalid enable password") {
		t.Log(err)
		// t.Error(buf.Len(), buf.String())

		s := shell.ToHexStringIfNeed(buf.Bytes())
		t.Error(s)
		fmt.Println(s)
	}
}



func TestCiscoFail2(t *testing.T) {

	// go http.ListenAndServe(":12445", nil)

	moreAfter := append(append(bytes.Repeat([]byte{byte('\b')}, 9), []byte("        ")...), bytes.Repeat([]byte{byte('\b')}, 9)...)

	welcome := []byte{0xFF, 0xFB, 0x01, 0xFF, 0xFB, 0x03, 0xFF, 0xFD, 0x18, 0xFF, 0xFD, 0x1F, 0x0D, 0x0A, 0x0D, 0x0A}
	welcome = append(welcome, []byte("*****************\r\n   <myuser> <mypassword> \r\nUser Access Verification\r\n\r\n")...)

	options := &telnetd.Options{}
	options.SetWelcome(welcome)
	options.SetUserQuest(append([]byte("Username: "), 0xFF, 0xFA, 0x18, 0x01, 0xFF, 0xF0), []byte("Password:"))
	options.SetUserPassword("admin1", "admin2")

	//options.WithEnable("ABC>", "enable", "password:", "testsx", "", "abc#", telnetd.Echo)
	options.WithPrompt([]byte("Switch>"),
		telnetd.WithEnable("enable", "Password: ", "admin", "", "Switch#", telnetd.OS(telnetd.Commands{
			"show": telnetd.WithCommands(telnetd.Commands{
				"configuration": telnetd.WithMore(ciscoConfigurations, []byte(" --More--"), moreAfter),
			}),
		})))

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
		Username: "admin1",
		// PasswordQuest: "",
		Password:            "admin2",
		Prompt:              "",
		EnableCommand:       "",
		EnablePasswordQuest: "",
		EnablePassword:      "admintset",
		EnablePrompt:        "",
		UseCRLF:             true,
	}

	var buf bytes.Buffer
	c, _, err := DailTelnet(ctx, params, ServerWriter(&buf), ClientWriter(&buf), Question(AbcQuestion.Prompts(), AbcQuestion.Do()))

	if err == nil {
		defer c.Close()

		t.Error("want error go ok")

		s := shell.ToHexStringIfNeed(buf.Bytes())
		t.Error(s)
		fmt.Println(s)
		return
	}

	if !strings.Contains(err.Error(), "invalid enable password") {
		t.Log(err)
		// t.Error(buf.Len(), buf.String())

		s := shell.ToHexStringIfNeed(buf.Bytes())
		t.Error(s)
		fmt.Println(s)
	}
}
package shell

import (
	"bytes"
	"encoding/json"
	"testing"
	"unicode"
)

func TestParseCmdOutput(t *testing.T) {
	for idx, test := range []struct {
		data           string
		result         string
		prompt         string
		characteristic string
	}{{data: `--- JUNOS 11.4R7.5 built 2013-03-01 11:57:40 UTC
nEwtimer@151-J-M10I-P2> show confi
nEwtimer@151-J-M10I-P2> ...          uration 
nEwtimer@151-J-M10I-P2> ...         display
nEwtimer@151-J-M10I-P2> ...y       s
nEwtimer@151-J-M10I-P2> ...n | disp...
nEwtimer@151-J-M10I-P2> ...display             
nEwtimer@151-J-M10I-P2> ...uration ...
nEwtimer@151-J-M10I-P2> ...ion | di               
nEwtimer@151-J-M10I-P2> ...configur...
nEwtimer@151-J-M10I-P2> ...iguratio               
nEwtimer@151-J-M10I-P2> show config...                                    

nEwtimer@151-J-M10I-P2> 

nEwtimer@151-J-M10I-P2> show confi
nEwtimer@151-J-M10I-P2> ...          uration 
nEwtimer@151-J-M10I-P2> ...         display
nEwtimer@151-J-M10I-P2> ...y       set |
nEwtimer@151-J-M10I-P2> ...         tests.MO
nEwtimer@151-J-M10I-P2> ...t | tests.MO   
                               ^
syntax error, expecting <command>.

nEwtimer@151-J-M10I-P2> ...t | tests.MO   
nEwtimer@151-J-M10I-P2> ...display ...
nEwtimer@151-J-M10I-P2> ...lay set ...
nEwtimer@151-J-M10I-P2> ...set |tests.MO    tests.MO   
nEwtimer@151-J-M10I-P2> ...| displa...
nEwtimer@151-J-M10I-P2> ...splay se...
error: syntax error, expecting <command>: semo

nEwtimer@151-J-M10I-P2> show confi
nEwtimer@151-J-M10I-P2> ...          uration 
nEwtimer@151-J-M10I-P2> ...         display
nEwtimer@151-J-M10I-P2> ...y       set |
nEwtimer@151-J-M10I-P2> ...         tests.MO
nEwtimer@151-J-M10I-P2> ...m   
nEwtimer@151-J-M10I-P2> ...lay set ...
nEwtimer@151-J-M10I-P2> ...n          o-more
nEwtimer@151-J-M10I-P2> ...         
set version 11.4R7.5
set system host-name 151-J-M10I-P2
set system time-zone Asia/Shanghai
set firewall filter SH-COS-ACL term network-control-queue0-5percent then accept

nEwtimer@151-J-M10I-P2> `,
		result: `set version 11.4R7.5
set system host-name 151-J-M10I-P2
set system time-zone Asia/Shanghai
set firewall filter SH-COS-ACL term network-control-queue0-5percent then accept

`,
		prompt:         ">",
		characteristic: "set system"},
		{data: `nEwtimer@151-J-M10I-P2> show confi
nEwtimer@151-J-M10I-P2> ...          uration 
nEwtimer@151-J-M10I-P2> ...         display
nEwtimer@151-J-M10I-P2> ...y       set |
nEwtimer@151-J-M10I-P2> ...         tests.MO
nEwtimer@151-J-M10I-P2> ...m   
nEwtimer@151-J-M10I-P2> ...lay set ...
nEwtimer@151-J-M10I-P2> ...n          o-more
nEwtimer@151-J-M10I-P2> ...         
set version 11.4R7.5
set system host-name 151-J-M10I-P2
set system time-zone Asia/Shanghai
nEwtimer@151-J-M10I-P2>
set version 11.4R7.5
set system host-name 151-J-M10I-P2
set system time-zone Asia/Shanghai
set firewall filter SH-COS-ACL term network-control-queue0-5percent then accept

nEwtimer@151-J-M10I-P2> `,
			result: `set version 11.4R7.5
set system host-name 151-J-M10I-P2
set system time-zone Asia/Shanghai
set firewall filter SH-COS-ACL term network-control-queue0-5percent then accept

`,
			prompt:         ">",
			characteristic: "set system"},
		{data: `????

User Access Verification

Username: nEwtimer
Password: 
151-C-2950-BankOUT>en
Password: 
151-C-2950-BankOUT#sh run
Building configuration...

Current configuration : 5257 bytes
ntp source Vlan100
ntp server 10.151.100.99
end

151-C-2950-BankOUT#`,
			result: `Building configuration...

Current configuration : 5257 bytes
ntp source Vlan100
ntp server 10.151.100.99
end

`,
			prompt:         "#",
			characteristic: "Building configuration"},
		{data: `????Remote Management Console
login: nEwtimer
password: 
FW-HA:151-J-ISG1000-1(M)-> get conf
Total Config size 35092:
unset key protection enable
set clock timezone 8
set clock dst recurring start-weekday 2 0 3 02:00 end-weekday 1 0 11 02:00
set vrouter trust-vr sharable
set vrouter "trust-vr"
exit
FW-HA:151-J-ISG1000-1(M)->          `,
			result: `Total Config size 35092:
unset key protection enable
set clock timezone 8
set clock dst recurring start-weekday 2 0 3 02:00 end-weekday 1 0 11 02:00
set vrouter trust-vr sharable
set vrouter "trust-vr"
exit
`,
			prompt:         ">",
			characteristic: "Total Config"},
		{data: `show co
NM-UsrA@150-J-2320-Bank-1> ...on     fi
NM-UsrA@150-J-2320-Bank-1> ...ig    ur
NM-UsrA@150-J-2320-Bank-1> ...ra    ti
NM-UsrA@150-J-2320-Bank-1> ...io    n 
NM-UsrA@150-J-2320-Bank-1> ... |     d
NM-UsrA@150-J-2320-Bank-1> ...di    sp
NM-UsrA@150-J-2320-Bank-1> ...pl    ay
NM-UsrA@150-J-2320-Bank-1> ...y     se
NM-UsrA@150-J-2320-Bank-1> ...et     |
NM-UsrA@150-J-2320-Bank-1> ...|     no
NM-UsrA@150-J-2320-Bank-1> ...o-    tests.MO
NM-UsrA@150-J-2320-Bank-1> ...or    e
set services service-set szt1-nat nat-rules szt1-input
set services service-set szt1-nat nat-rules szt1-output
NM-UsrA@150-J-2320-Bank-1>`,
			result: `set services service-set szt1-nat nat-rules szt1-input
set services service-set szt1-nat nat-rules szt1-output
`,
			prompt:         ">",
			characteristic: "set services",
		},
		{data: ` show configu
NM-UsrA@157-J-2320-2> ...ation | d
NM-UsrA@157-J-2320-2> ...splay set
NM-UsrA@157-J-2320-2> ...| no-more
NM-UsrA@157-J-2320-2> ...
set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
NM-UsrA@157-J-2320-2>`,
			result: `set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
`,
			prompt:         ">",
			characteristic: "set system",
		},
		{data: ` abcabc
set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
NM-UsrA@157-J-2320-2>`,
			result: `set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
`,
			prompt:         ">",
			characteristic: "set system",
		},
		{data: ` abcabc
abcabc
set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
NM-UsrA@157-J-2320-2>`,
			result: `set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
`,
			prompt:         ">",
			characteristic: "set system",
		},
		{data: `abcabc
abcabc
set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
NM-UsrA@157-J-2320-2>`,
			result: `set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
`,
			prompt:         ">",
			characteristic: "set system",
		},

		//  test for  removeCtrlCharByLine
		{data: `abcabc
abcabc
set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
-- more -- set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
NM-UsrA@157-J-2320-2>`,
			result: `set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
`,
			prompt:         ">",
			characteristic: "set system",
		},
		{data: `abcabc
abcabc
set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
-- more -- set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
NM-UsrA@157-J-2320-2>`,
			result: `set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
`,
			prompt:         ">",
			characteristic: "set system",
		},
		{data: `abcabc
abcabc
set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
-- more -- set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
NM-UsrA@157-J-2320-2>`,
			result: `set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
`,
			prompt:         ">",
			characteristic: "set system",
		},
		{data: `abcabc
abcabc
set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
-- more -- ssset firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
NM-UsrA@157-J-2320-2>`,
			result: `set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
`,
			prompt:         ">",
			characteristic: "set system",
		},
		{data: `NM-UsrA@157-J-2320-2>
NM-UsrA@157-J-2320-2>
abcabc
abcabc
set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
-- more -- ssset firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
NM-UsrA@157-J-2320-2>`,
			result: `set version 9.3R4.4
set system host-name 157-J-2320-2
set system time-zone Asia/Shanghai
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 from source-prefix-list 89
set firewall filter 89-BJ-CaiWu-OUTBOUND term 1 then accept
`,
			prompt: ">",
		}} {
		t.Log("[", idx, "] begining")
		res, e := ParseCmdOutput([]byte(test.data), []byte("abcabc"), []byte(test.prompt), []byte(test.characteristic))
		if nil != e {
			t.Error("[", idx, "]", e)
			continue
		}
		if !bytes.Equal(res, []byte(test.result)) {
			t.Error("[", idx, "]", "result is diff.")
			t.Log("'" + string(res) + "'")
			t.Log("'" + test.result + "'")
		}
		t.Log("[", idx, "] end")
	}
}

// func TestParse(t testing.T) {
//  txt := `show running-config\r\n: Saved\r\n\r\n: \r\n: Serial Number: FCH2025J973\r\n: Hardware:   ASA5525, 8192 MB RAM, CPU Lynnfield 2394 MHz, 1 CPU (4 cores)\r\n:\r\nASA Version 9.4(4)5 \r\n!\r\nhostname TFSX-JY-FW5525\r\nenable password ienZVNT8yp0uLBVU encrypted\r\nnames\r\n!\r\ninterface GigabitEthernet0/0\r\n nameif outside_dx\r\n security-level 0\r\n ip address 101.95.12.18 255.255.255.252 \r\n!\r\ninterface GigabitEthernet0/1\r\n nameif outside_lt\r\n security-level 0\r\n ip address 112.65.130.226 255.255.255.0 \r\n!\r\ninterface GigabitEthernet0/2\r\n nameif inside\r\n security-level 100\r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000 ip address 192.168.111.254 255.255.255.0 \r\n!\r\ninterface GigabitEthernet0/3\r\n shutdown\r\n no nameif\r\n no security-level\r\n no ip address\r\n!\r\ninterface GigabitEthernet0/4\r\n shutdown\r\n no nameif\r\n no security-level\r\n no ip address\r\n!\r\ninterface GigabitEthernet0/5\r\n shutdown\r\n no nameif\r\n no security-level\r\n no ip address\r\n!\r\ninterface GigabitEthernet0/6\r\n shutdown\r\n no nameif\r\n no security-level\r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000 no ip address\r\n!\r\ninterface GigabitEthernet0/7\r\n description LAN/STATE Failover Interface\r\n!\r\ninterface Management0/0\r\n management-only\r\n nameif management\r\n security-level 100\r\n ip address 192.168.98.31 255.255.255.0 standby 192.168.98.32 \r\n!\r\nboot system disk0:/asa944-5-smp-k8.bin\r\nftp mode passive\r\nobject network 111.89\r\n host 192.168.111.89\r\nobject network ctp2trade1lt51\r\n host 192.168.111.51\r\nobject network ctp2trade1dx51\r\n host 192.168.111.51\r\nobject network ctp2market1lt51\r\n host 192.168.111.51\r\nobject network ctp2market1dx51\r\n host 192.168.111.51\r\nobject network ctp2trade2lt52\r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000 host 192.168.111.52\r\nobject network ctp2trade2dx52\r\n host 192.168.111.52\r\nobject network ctp2market1lt52\r\n host 192.168.111.52\r\nobject network ctp2market1dx52\r\n host 192.168.111.52\r\nobject network ctp2-pb-phone-lt-89\r\n host 192.168.111.89\r\nobject network ctp2-pb-phone-dx-89\r\n host 192.168.111.89\r\nobject network 192.168.100.89-5900\r\n host 192.168.111.89\r\nobject network 192.168.100.89-5900-lt\r\n host 192.168.111.89\r\nobject network ctp-operlt-88\r\n host 192.168.111.88\r\nobject network ctp-operdx-88\r\n host 192.168.111.88\r\nobject network ctp1jylt70\r\n host 192.168.111.70\r\nobject network ctp1jydx70\r\n host 192.168.111.70\r\nobject network ctp1mdlt70\r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000 host 192.168.111.70\r\nobject network ctp1mddx70\r\n host 192.168.111.70\r\nobject network ctp1jylt71\r\n host 192.168.111.71\r\nobject network ctp1jydx71\r\n host 192.168.111.71\r\nobject network ctp1mdlt71\r\n host 192.168.111.71\r\nobject network ctp1mddx71\r\n host 192.168.111.71\r\nobject network ctp1jylt72\r\n host 192.168.111.72\r\nobject network ctp1jydx72\r\n host 192.168.111.72\r\nobject network ctp1mdlt72\r\n host 192.168.111.72\r\nobject network ctp1mddx72\r\n host 192.168.111.72\r\nobject network ctpbeichajylt84\r\n host 192.168.111.84\r\nobject network ctpbeichajydx84\r\n host 192.168.111.84\r\nobject network ctpbeichamdlt84\r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000 host 192.168.111.84\r\nobject network ctpbeichamddx84\r\n host 192.168.111.84\r\nobject network nonghangceshi89yw\r\n host 192.168.111.89\r\nobject network nonghangceshi89ftp\r\n host 192.168.111.89\r\nobject network 58.32.236.38\r\n host 58.32.236.38\r\naccess-list outside-acl extended permit icmp any any \r\naccess-list outside-acl extended permit tcp any host 192.168.111.51 eq 41205 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.51 eq 41213 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.52 eq 41205 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.52 eq 41213 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.89 eq 20002 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.89 eq 5900 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.88 eq 5900 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.70 eq 41205 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.70 eq 41213 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.71 eq 41205 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.71 eq 41213 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.72 eq 41205 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.72 eq 41213 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.84 eq 41205 \r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000access-list outside-acl extended permit tcp any host 192.168.111.84 eq 41213 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.89 eq 6667 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.89 eq ftp \r\naccess-list outside-acl extended permit tcp any host 192.168.111.89 eq 5000 \r\naccess-list outside-acl extended permit tcp any host 192.168.111.89 eq 4000 \r\npager lines 24\r\nlogging asdm informational\r\nmtu outside_dx 1500\r\nmtu outside_lt 1500\r\nmtu inside 1500\r\nmtu management 1500\r\nfailover\r\nfailover lan unit secondary\r\nfailover lan interface fo GigabitEthernet0/7\r\nfailover key *****\r\nfailover link fo GigabitEthernet0/7\r\nfailover interface ip fo 1.1.1.1 255.255.255.252 standby 1.1.1.2\r\nicmp unreachable rate-limit 1 burst-size 1\r\nasdm image disk0:/asdm-7221.bin\r\nno asdm history enable\r\narp timeout 14400\r\nno arp permit-nonconnected\r\nnat (inside,outside_dx) source static 111.89 58.32.236.38\r\n!\r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000object network ctp2trade1lt51\r\n nat (inside,outside_lt) static 112.65.23.33 service tcp 41205 61205 \r\nobject network ctp2trade1dx51\r\n nat (inside,outside_dx) static 58.32.236.33 service tcp 41205 61205 \r\nobject network ctp2market1lt51\r\n nat (inside,outside_lt) static 112.65.23.33 service tcp 41213 61213 \r\nobject network ctp2market1dx51\r\n nat (inside,outside_dx) static 58.32.236.33 service tcp 41213 61213 \r\nobject network ctp2trade2lt52\r\n nat (inside,outside_lt) static 112.65.23.34 service tcp 41205 61205 \r\nobject network ctp2trade2dx52\r\n nat (inside,outside_dx) static 58.32.236.34 service tcp 41205 61205 \r\nobject network ctp2market1lt52\r\n nat (inside,outside_lt) static 112.65.23.34 service tcp 41213 61213 \r\nobject network ctp2market1dx52\r\n nat (inside,outside_dx) static 58.32.236.34 service tcp 41213 61213 \r\nobject network ctp2-pb-phone-lt-89\r\n nat (inside,outside_lt) static 112.65.23.35 service tcp 20002 20002 \r\nobject network ctp2-pb-phone-dx-89\r\n nat (inside,outside_dx) static 58.32.236.35 service tcp 20002 20002 \r\nobject network 192.168.100.89-5900\r\n nat (inside,outside_dx) static 58.32.236.36 service tcp 5900 59002 \r\nobject network 192.168.100.89-5900-lt\r\n nat (inside,outside_lt) static 112.65.23.36 service tcp 5900 59002 \r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000object network ctp-operlt-88\r\n nat (inside,outside_lt) static 112.65.23.36 service tcp 5900 59001 \r\nobject network ctp-operdx-88\r\n nat (inside,outside_dx) static 58.32.236.36 service tcp 5900 59001 \r\nobject network ctp1jylt70\r\n nat (inside,outside_lt) static 112.65.23.33 service tcp 41205 51205 \r\nobject network ctp1jydx70\r\n nat (inside,outside_dx) static 58.32.236.33 service tcp 41205 51205 \r\nobject network ctp1mdlt70\r\n nat (inside,outside_lt) static 112.65.23.33 service tcp 41213 51213 \r\nobject network ctp1mddx70\r\n nat (inside,outside_dx) static 58.32.236.33 service tcp 41213 51213 \r\nobject network ctp1jylt71\r\n nat (inside,outside_lt) static 112.65.23.34 service tcp 41205 51205 \r\nobject network ctp1jydx71\r\n nat (inside,outside_dx) static 58.32.236.34 service tcp 41205 51205 \r\nobject network ctp1mdlt71\r\n nat (inside,outside_lt) static 112.65.23.34 service tcp 41213 51213 \r\nobject network ctp1mddx71\r\n nat (inside,outside_dx) static 58.32.236.34 service tcp 41213 51213 \r\nobject network ctp1jylt72\r\n nat (inside,outside_lt) static 112.65.23.37 service tcp 41205 51205 \r\nobject network ctp1jydx72\r\n nat (inside,outside_dx) static 58.32.236.37 service tcp 41205 51205 \r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000object network ctp1mdlt72\r\n nat (inside,outside_lt) static 112.65.23.37 service tcp 41213 51213 \r\nobject network ctp1mddx72\r\n nat (inside,outside_dx) static 58.32.236.37 service tcp 41213 51213 \r\nobject network ctpbeichajylt84\r\n nat (inside,outside_lt) static 112.65.23.33 service tcp 41205 31205 \r\nobject network ctpbeichajydx84\r\n nat (inside,outside_dx) static 58.32.236.33 service tcp 41205 31205 \r\nobject network ctpbeichamdlt84\r\n nat (inside,outside_lt) static 112.65.23.33 service tcp 41213 31213 \r\nobject network ctpbeichamddx84\r\n nat (inside,outside_dx) static 58.32.236.33 service tcp 41213 31213 \r\nobject network nonghangceshi89yw\r\n nat (inside,outside_dx) static 58.32.236.38 service tcp 6667 6667 \r\nobject network nonghangceshi89ftp\r\n nat (inside,outside_dx) static 58.32.236.38 service tcp ftp ftp \r\n!\r\nnat (inside,outside_lt) after-auto source dynamic any interface\r\nnat (inside,outside_dx) after-auto source dynamic any interface\r\naccess-group outside-acl in interface outside_dx\r\naccess-group outside-acl in interface outside_lt\r\nroute outside_dx 0.0.0.0 0.0.0.0 101.95.12.17 90\r\nroute outside_lt 0.0.0.0 0.0.0.0 112.65.130.225 100\r\nroute management 192.168.99.0 255.255.255.0 192.168.98.254 1\r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000route management 192.168.100.0 255.255.255.0 192.168.98.254 1\r\ntimeout xlate 3:00:00\r\ntimeout pat-xlate 0:00:30\r\ntimeout conn 1:00:00 half-closed 0:10:00 udp 0:02:00 icmp 0:00:02\r\ntimeout sunrpc 0:10:00 h323 0:05:00 h225 1:00:00 mgcp 0:05:00 mgcp-pat 0:05:00\r\ntimeout sip 0:30:00 sip_media 0:02:00 sip-invite 0:03:00 sip-disconnect 0:02:00\r\ntimeout sip-provisional-media 0:02:00 uauth 0:05:00 absolute\r\ntimeout tcp-proxy-reassembly 0:01:00\r\ntimeout floating-conn 0:00:00\r\nuser-identity default-domain LOCAL\r\naaa authentication telnet console LOCAL \r\naaa authentication ssh console LOCAL \r\nhttp server enable\r\nhttp 192.168.1.0 255.255.255.0 management\r\nsnmp-server host management 192.168.98.77 community ***** version 2c\r\nsnmp-server host management 192.168.98.190 community ***** version 2c\r\nno snmp-server location\r\nno snmp-server contact\r\ncrypto ipsec security-association pmtu-aging infinite\r\ncrypto ca trustpool policy\r\ntelnet 0.0.0.0 0.0.0.0 management\r\ntelnet timeout 5\r\nno ssh stricthostkeycheck\r\nssh 0.0.0.0 0.0.0.0 management\r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000ssh timeout 5\r\nssh key-exchange group dh-group1-sha1\r\nconsole timeout 0\r\nthreat-detection basic-threat\r\nthreat-detection statistics access-list\r\nno threat-detection statistics tcp-intercept\r\nssl cipher default custom \"RC4-SHA:AES128-SHA:AES256-SHA:DES-CBC3-SHA\"\r\nssl cipher tlsv1 custom \"RC4-SHA:AES128-SHA:AES256-SHA:DES-CBC3-SHA\"\r\nssl cipher dtlsv1 custom \"RC4-SHA:AES128-SHA:AES256-SHA:DES-CBC3-SHA\"\r\ndynamic-access-policy-record DfltAccessPolicy\r\nusername tfadmin password OUXKFYFR8iMiidOd encrypted\r\n!\r\nclass-map inspection_default\r\n match default-inspection-traffic\r\n!\r\n!\r\npolicy-map type inspect dns preset_dns_map\r\n parameters\r\n  message-length maximum client auto\r\n  message-length maximum 512\r\npolicy-map global_policy\r\n class inspection_default\r\n  inspect dns preset_dns_map \r\n  inspect ftp \r\n\u003c--- More - --\u0026gt;\r\u0000              \r\u0000  inspect h323 h225 \r\n  inspect h323 ras \r\n  inspect rsh \r\n  inspect rtsp \r\n  inspect esmtp \r\n  inspect sqlnet \r\n  inspect skinny  \r\n  inspect sunrpc \r\n  inspect xdmcp \r\n  inspect sip  \r\n  inspect netbios \r\n  inspect tftp \r\n  inspect ip-options \r\n  inspect icmp \r\n!\r\nservice-policy global_policy global\r\nprompt hostname context \r\nno call-home reporting anonymous\r\nCryptochecksum:8cdaf3058057119e5b7f36e6955f4c58\r\n: end\r\n\r\u0000TFSX-JY-FW5525#`

//  ParseCmdOutput([]byte, cmd, prompt, characteristic)
// }

func TestTrimSpace(t *testing.T) {
	txt := `"\r\u0000TFSX-JY-FW5525#"`
	var value string
	json.Unmarshal([]byte(txt), &value)

	bs := bytes.TrimFunc([]byte(value), func(r rune) bool {
		t.Log("==", int(r), int('\r'))
		if r == 0 || r == '\r' {
			return true
		}
		return unicode.IsSpace(r)
	})

	t.Log(value[2:])
	t.Log(string(bs))
}

var raw_chars = []byte(`151-C-2950-BankOUT#sh run
Building configuration...
clock timezone ShangHai 8
ip subnet-zero
 --More--         no ip source-route
no ip domain-lookup
!
spanning-tree portfast bpdufilter default
spanning-tree extend system-id`)

var rs_chars = []byte(`151-C-2950-BankOUT#sh run
Building configuration...
clock timezone ShangHai 8
ip subnet-zero
no ip source-route
no ip domain-lookup
!
spanning-tree portfast bpdufilter default
spanning-tree extend system-id`)

var raw_chars2 = [][]byte{[]byte(`- `),
	[]byte(`        no ip source-route`),
	[]byte(`no ip domain-lookup`)}
var rs_chars2 = []byte("\nno ip source-route\nno ip domain-lookup\n")

var raw_chars3 = [][]byte{[]byte(`interface GigabitEthernet0/0/19`),
	[]byte(`  ---- More ----[42D                                          [42D#`),
	[]byte(`interface GigabitEthernet0/0/20`),
	[]byte(` user privilege level 5`),
	[]byte(`  ---- More ----[42D                                          [42D protocol inbound ssh`),
	[]byte(`#`),
	[]byte(`return`)}

var rs_chars3 = []byte("interface GigabitEthernet0/0/19\n#\ninterface GigabitEthernet0/0/20\n user privilege level 5\n protocol inbound ssh\n#\nreturn\n")

func TestRemoveChar(t *testing.T) {
	rs := RemoveCtrlChar(raw_chars)
	if !bytes.Equal(rs, rs_chars) {
		t.Error(string(rs))
	}

	rs = RemoveCtrlCharByLine(raw_chars2, 300)
	if !bytes.Equal(rs, rs_chars2) {
		t.Error("actal is", string(rs))
		t.Error("excepted is", string(rs_chars2))
	}

	rs = RemoveCtrlCharByLine(raw_chars3, 300)
	if !bytes.Equal(rs, rs_chars3) {
		t.Error("actal is", string(rs))
		t.Error("excepted is", string(rs_chars3))
	}
}

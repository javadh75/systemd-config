package systemdconfig

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func unitOf(sections ...*Section) *Unit {
	return &Unit{Sections: sections}
}

func sectionOf(name string, options ...*OptionValue) *Section {
	return &Section{Name: name, Options: options}
}

func optionOf(option, value string) *OptionValue {
	return &OptionValue{Option: option, Value: value}
}

// TestRealWorldConfigs deserializes real-world systemd config files from
// testdata/ and asserts the exact Unit each one must parse into: section
// and option order, duplicate sections and options, stripped comments,
// joined continuation lines, trimmed assignments, and CRLF handling.
func TestRealWorldConfigs(t *testing.T) {
	tests := []struct {
		file string
		want *Unit
	}{
		{
			file: "example.service",
			want: unitOf(
				sectionOf("Unit",
					optionOf("Description", "Example daemon"),
					optionOf("After", "network.target"),
				),
				sectionOf("Service",
					optionOf("ExecStart", "/usr/bin/example --flag=value"),
					optionOf("Environment", "FOO=bar"),
				),
				sectionOf("Install",
					optionOf("WantedBy", "multi-user.target"),
				),
			),
		},
		{
			file: "example.network",
			want: unitOf(
				sectionOf("Match",
					optionOf("Name", "eth0"),
				),
				sectionOf("Network",
					optionOf("DHCP", "no"),
					optionOf("DNS", "1.1.1.1"),
					optionOf("DNS", "8.8.8.8"),
				),
				sectionOf("Address",
					optionOf("Address", "10.0.0.2/24"),
				),
				sectionOf("Address",
					optionOf("Address", "2001:db8::2/64"),
				),
			),
		},
		{
			file: "sshd.service",
			want: unitOf(
				sectionOf("Unit",
					optionOf("Description", "OpenBSD Secure Shell server"),
					optionOf("Documentation", "man:sshd(8) man:sshd_config(5)"),
					optionOf("After", "network.target auditd.service"),
					optionOf("ConditionPathExists", "!/etc/ssh/sshd_not_to_be_run"),
				),
				sectionOf("Service",
					optionOf("EnvironmentFile", "-/etc/default/ssh"),
					optionOf("ExecStartPre", "/usr/sbin/sshd -t"),
					optionOf("ExecStart", "/usr/sbin/sshd -D $SSHD_OPTS"),
					optionOf("ExecReload", "/usr/sbin/sshd -t"),
					optionOf("ExecReload", "/bin/kill -HUP $MAINPID"),
					optionOf("KillMode", "process"),
					optionOf("Restart", "on-failure"),
					optionOf("RestartPreventExitStatus", "255"),
					optionOf("Type", "notify"),
					optionOf("RuntimeDirectory", "sshd"),
					optionOf("RuntimeDirectoryMode", "0755"),
				),
				sectionOf("Install",
					optionOf("WantedBy", "multi-user.target"),
					optionOf("Alias", "sshd.service"),
				),
			),
		},
		{
			file: "docker.service",
			want: unitOf(
				sectionOf("Unit",
					optionOf("Description", "Docker Application Container Engine"),
					optionOf("Documentation", "https://docs.docker.com"),
					optionOf("After", "network-online.target docker.socket firewalld.service containerd.service time-set.target"),
					optionOf("Wants", "network-online.target containerd.service"),
					optionOf("Requires", "docker.socket"),
				),
				sectionOf("Service",
					optionOf("Type", "notify"),
					optionOf("ExecStart", "/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock"),
					optionOf("ExecReload", "/bin/kill -s HUP $MAINPID"),
					optionOf("TimeoutStartSec", "0"),
					optionOf("RestartSec", "2"),
					optionOf("Restart", "always"),
					optionOf("StartLimitBurst", "3"),
					optionOf("StartLimitInterval", "60s"),
					optionOf("LimitNPROC", "infinity"),
					optionOf("LimitCORE", "infinity"),
					optionOf("TasksMax", "infinity"),
					optionOf("Delegate", "yes"),
					optionOf("KillMode", "process"),
					optionOf("OOMScoreAdjust", "-500"),
				),
				sectionOf("Install",
					optionOf("WantedBy", "multi-user.target"),
				),
			),
		},
		{
			file: "docker.socket",
			want: unitOf(
				sectionOf("Unit",
					optionOf("Description", "Docker Socket for the API"),
				),
				sectionOf("Socket",
					optionOf("ListenStream", "/run/docker.sock"),
					optionOf("SocketMode", "0660"),
					optionOf("SocketUser", "root"),
					optionOf("SocketGroup", "docker"),
				),
				sectionOf("Install",
					optionOf("WantedBy", "sockets.target"),
				),
			),
		},
		{
			file: "wg0.netdev",
			want: unitOf(
				sectionOf("NetDev",
					optionOf("Name", "wg0"),
					optionOf("Kind", "wireguard"),
					optionOf("Description", "WireGuard VPN tunnel"),
				),
				sectionOf("WireGuard",
					optionOf("ListenPort", "51820"),
					optionOf("PrivateKeyFile", "/etc/systemd/network/wg0.key"),
				),
				sectionOf("WireGuardPeer",
					optionOf("PublicKey", "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg="),
					optionOf("AllowedIPs", "10.192.122.3/32, 10.192.124.0/24"),
					optionOf("Endpoint", "vpn1.example.com:51820"),
					optionOf("PersistentKeepalive", "25"),
				),
				sectionOf("WireGuardPeer",
					optionOf("PublicKey", "TrMvSoP4jYQlY6RIzBgbssQqY3vxI2Pi+y71lOWWXX0="),
					optionOf("AllowedIPs", "10.192.122.4/32"),
					optionOf("Endpoint", "[2607:5300:60:6b0::c05f:543]:2468"),
				),
			),
		},
		{
			file: "static.network",
			want: unitOf(
				sectionOf("Match",
					optionOf("Name", "enp2s0"),
				),
				sectionOf("Network",
					optionOf("DHCP", "no"),
					optionOf("DNS", "192.168.0.1"),
					optionOf("DNS", "2001:4860:4860::8888"),
					optionOf("Domains", "example.com"),
				),
				sectionOf("Address",
					optionOf("Address", "192.168.0.15/24"),
					optionOf("Broadcast", "192.168.0.255"),
					optionOf("Label", "uplink"),
				),
				sectionOf("Address",
					optionOf("Address", "2001:db8:dead:beef::5/64"),
				),
				sectionOf("Route",
					optionOf("Gateway", "192.168.0.1"),
					optionOf("Metric", "100"),
				),
				sectionOf("Route",
					optionOf("Destination", "10.0.0.0/8"),
					optionOf("Gateway", "192.168.0.254"),
					optionOf("GatewayOnLink", "yes"),
				),
			),
		},
		{
			file: "home.mount",
			want: unitOf(
				sectionOf("Unit",
					optionOf("Description", "Home Directory"),
					optionOf("Before", "local-fs.target"),
				),
				sectionOf("Mount",
					optionOf("What", "/dev/disk/by-uuid/f5872a89-8a9c-4e42-a17e-6cb92b7e72a4"),
					optionOf("Where", "/home"),
					optionOf("Type", "ext4"),
					optionOf("Options", "defaults,noatime"),
				),
				sectionOf("Install",
					optionOf("WantedBy", "local-fs.target"),
				),
			),
		},
		{
			file: "backup.timer",
			want: unitOf(
				sectionOf("Unit",
					optionOf("Description", "Daily backup of /srv"),
					optionOf("Requires", "backup.service"),
				),
				sectionOf("Timer",
					optionOf("OnCalendar", "*-*-* 02:00:00"),
					optionOf("RandomizedDelaySec", "30min"),
					optionOf("Persistent", "true"),
					optionOf("AccuracySec", "1h"),
				),
				sectionOf("Install",
					optionOf("WantedBy", "timers.target"),
				),
			),
		},
		{
			file: "journald.conf",
			want: unitOf(
				sectionOf("Journal",
					optionOf("Storage", "persistent"),
					optionOf("Compress", "yes"),
					optionOf("SystemMaxUse", "500M"),
					optionOf("SystemKeepFree", "1G"),
					optionOf("MaxRetentionSec", "1month"),
					optionOf("ForwardToSyslog", "no"),
				),
			),
		},
		{
			file: "resolved.conf",
			want: unitOf(
				sectionOf("Resolve",
					optionOf("DNS", "9.9.9.9#dns.quad9.net 2620:fe::fe#dns.quad9.net"),
					optionOf("FallbackDNS", "1.1.1.1 8.8.8.8 2001:4860:4860::8888"),
					optionOf("Domains", "~."),
					optionOf("DNSSEC", "allow-downgrade"),
					optionOf("DNSOverTLS", "opportunistic"),
					optionOf("MulticastDNS", "no"),
					optionOf("LLMNR", "no"),
					optionOf("Cache", "yes"),
					optionOf("DNSStubListener", "yes"),
				),
			),
		},
		{
			file: "override.conf",
			want: unitOf(
				sectionOf("Service",
					optionOf("ExecStart", ""),
					optionOf("ExecStart", "/usr/sbin/nginx -c /etc/nginx/custom.conf -g 'daemon on; master_process on;'"),
					optionOf("LimitNOFILE", "65536"),
					optionOf("Nice", "-5"),
				),
			),
		},
		{
			file: "etcd.service",
			want: unitOf(
				sectionOf("Unit",
					optionOf("Description", "etcd key-value store"),
					optionOf("Documentation", "https://github.com/etcd-io/etcd"),
					optionOf("After", "network-online.target"),
					optionOf("Wants", "network-online.target"),
				),
				sectionOf("Service",
					optionOf("Type", "notify"),
					optionOf("User", "etcd"),
					optionOf("ExecStart", "/usr/local/bin/etcd --name=infra0 --cert-file=/etc/ssl/certs/etcd.pem --key-file=/etc/ssl/private/etcd-key.pem --initial-advertise-peer-urls=https://10.0.1.10:2380"),
					optionOf("Restart", "on-failure"),
					optionOf("RestartSec", "5"),
				),
				sectionOf("Install",
					optionOf("WantedBy", "multi-user.target"),
				),
			),
		},
		{
			file: "crlf.service",
			want: unitOf(
				sectionOf("Unit",
					optionOf("Description", "Telemetry agent"),
					optionOf("After", "network.target"),
				),
				sectionOf("Service",
					optionOf("Type", "simple"),
					optionOf("ExecStart", "/opt/agent/bin/agent --config /etc/agent.yaml"),
					optionOf("Restart", "always"),
				),
				sectionOf("Install",
					optionOf("WantedBy", "multi-user.target"),
				),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			got, err := Deserialize(bytes.NewReader(src))
			if err != nil {
				t.Fatalf("Deserialize(%s) error = %v", tt.file, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				var gotOut, wantOut bytes.Buffer
				_, _ = got.WriteTo(&gotOut)
				_, _ = tt.want.WriteTo(&wantOut)
				t.Errorf("Deserialize(%s) mismatch:\ngot:\n%s\nwant:\n%s", tt.file, gotOut.String(), wantOut.String())
			}
		})
	}
}

// TestRealWorldAccessors exercises the duplicate-section and
// last-assignment-wins accessors against the real-world fixtures.
func TestRealWorldAccessors(t *testing.T) {
	load := func(t *testing.T, file string) *Unit {
		t.Helper()
		src, err := os.ReadFile(filepath.Join("testdata", file))
		if err != nil {
			t.Fatal(err)
		}
		unit, err := Deserialize(bytes.NewReader(src))
		if err != nil {
			t.Fatalf("Deserialize(%s) error = %v", file, err)
		}
		return unit
	}

	t.Run("DuplicateSections", func(t *testing.T) {
		unit := load(t, "wg0.netdev")
		peers := unit.SectionsByName("WireGuardPeer")
		if len(peers) != 2 {
			t.Fatalf("SectionsByName(WireGuardPeer) returned %d sections, want 2", len(peers))
		}
		if got, _ := peers[1].Value("Endpoint"); got != "[2607:5300:60:6b0::c05f:543]:2468" {
			t.Errorf("second peer Endpoint = %q, want %q", got, "[2607:5300:60:6b0::c05f:543]:2468")
		}
	})

	t.Run("DuplicateOptions", func(t *testing.T) {
		unit := load(t, "static.network")
		want := []string{"192.168.0.1", "2001:4860:4860::8888"}
		if got := unit.SectionByName("Network").Values("DNS"); !reflect.DeepEqual(got, want) {
			t.Errorf("Values(DNS) = %v, want %v", got, want)
		}
	})

	t.Run("LastAssignmentWins", func(t *testing.T) {
		unit := load(t, "override.conf")
		want := "/usr/sbin/nginx -c /etc/nginx/custom.conf -g 'daemon on; master_process on;'"
		if got, ok := unit.SectionByName("Service").Value("ExecStart"); !ok || got != want {
			t.Errorf("Value(ExecStart) = %q, %v, want %q, true", got, ok, want)
		}
	})
}

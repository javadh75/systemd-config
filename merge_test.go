package systemdconfig

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		name    string
		base    *Unit
		dropins []*Unit
		want    *Unit
	}{
		{
			name: "NoDropinsCopiesBase",
			base: unitOf(
				sectionOf("Unit", optionOf("Description", "Test")),
				sectionOf("Install"),
			),
			want: unitOf(
				sectionOf("Unit", optionOf("Description", "Test")),
				sectionOf("Install"),
			),
		},
		{
			name: "SectionsConcatenateInOrder",
			base: unitOf(
				sectionOf("Service", optionOf("Restart", "always")),
			),
			dropins: []*Unit{
				unitOf(sectionOf("Service", optionOf("Restart", "on-failure"))),
			},
			want: unitOf(
				sectionOf("Service", optionOf("Restart", "always")),
				sectionOf("Service", optionOf("Restart", "on-failure")),
			),
		},
		{
			name: "EmptyAssignmentResetsAcrossUnits",
			base: unitOf(
				sectionOf("Service",
					optionOf("ExecStart", "/usr/sbin/nginx"),
					optionOf("Restart", "always"),
				),
			),
			dropins: []*Unit{
				unitOf(sectionOf("Service",
					optionOf("ExecStart", ""),
					optionOf("ExecStart", "/usr/sbin/nginx -c /etc/nginx/custom.conf"),
				)),
			},
			want: unitOf(
				sectionOf("Service", optionOf("Restart", "always")),
				sectionOf("Service", optionOf("ExecStart", "/usr/sbin/nginx -c /etc/nginx/custom.conf")),
			),
		},
		{
			name: "EmptyAssignmentResetsWithinOneUnit",
			base: unitOf(
				sectionOf("Service",
					optionOf("ExecStart", "/bin/a"),
					optionOf("ExecStart", ""),
					optionOf("ExecStart", "/bin/b"),
				),
			),
			want: unitOf(
				sectionOf("Service", optionOf("ExecStart", "/bin/b")),
			),
		},
		{
			name: "ResetOnlyTouchesSameSectionName",
			base: unitOf(
				sectionOf("Service", optionOf("ExecStart", "/bin/a")),
			),
			dropins: []*Unit{
				unitOf(sectionOf("Socket", optionOf("ExecStart", ""))),
			},
			want: unitOf(
				sectionOf("Service", optionOf("ExecStart", "/bin/a")),
				sectionOf("Socket"),
			),
		},
		{
			name: "ResetSpansDuplicateSections",
			base: unitOf(
				sectionOf("Network", optionOf("DNS", "1.1.1.1")),
				sectionOf("Network", optionOf("DNS", "8.8.8.8")),
			),
			dropins: []*Unit{
				unitOf(sectionOf("Network",
					optionOf("DNS", ""),
					optionOf("DNS", "9.9.9.9"),
				)),
			},
			want: unitOf(
				sectionOf("Network"),
				sectionOf("Network"),
				sectionOf("Network", optionOf("DNS", "9.9.9.9")),
			),
		},
		{
			name: "ListOptionsAccumulate",
			base: unitOf(
				sectionOf("Network", optionOf("DNS", "1.1.1.1")),
			),
			dropins: []*Unit{
				unitOf(sectionOf("Network", optionOf("DNS", "8.8.8.8"))),
			},
			want: unitOf(
				sectionOf("Network", optionOf("DNS", "1.1.1.1")),
				sectionOf("Network", optionOf("DNS", "8.8.8.8")),
			),
		},
		{
			name: "DuplicateSectionsSurviveMerging",
			base: unitOf(
				sectionOf("Address", optionOf("Address", "10.0.0.2/24")),
				sectionOf("Address", optionOf("Address", "10.0.0.3/24")),
			),
			dropins: []*Unit{
				unitOf(sectionOf("Address", optionOf("Address", "2001:db8::2/64"))),
			},
			want: unitOf(
				sectionOf("Address", optionOf("Address", "10.0.0.2/24")),
				sectionOf("Address", optionOf("Address", "10.0.0.3/24")),
				sectionOf("Address", optionOf("Address", "2001:db8::2/64")),
			),
		},
		{
			name: "MultipleDropinsApplyInOrder",
			base: unitOf(
				sectionOf("Service", optionOf("Nice", "0")),
			),
			dropins: []*Unit{
				unitOf(sectionOf("Service", optionOf("Nice", "-5"))),
				unitOf(sectionOf("Service", optionOf("Nice", "-10"))),
			},
			want: unitOf(
				sectionOf("Service", optionOf("Nice", "0")),
				sectionOf("Service", optionOf("Nice", "-5")),
				sectionOf("Service", optionOf("Nice", "-10")),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.base, tt.dropins...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() mismatch:\ngot:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestMergeDoesNotModifyInputs(t *testing.T) {
	base := unitOf(
		sectionOf("Service", optionOf("ExecStart", "/bin/a")),
	)
	dropin := unitOf(
		sectionOf("Service", optionOf("ExecStart", ""), optionOf("ExecStart", "/bin/b")),
	)
	wantBase := unitOf(
		sectionOf("Service", optionOf("ExecStart", "/bin/a")),
	)
	wantDropin := unitOf(
		sectionOf("Service", optionOf("ExecStart", ""), optionOf("ExecStart", "/bin/b")),
	)

	merged := Merge(base, dropin)

	if !reflect.DeepEqual(base, wantBase) {
		t.Errorf("Merge() modified base:\ngot:\n%s\nwant:\n%s", base, wantBase)
	}
	if !reflect.DeepEqual(dropin, wantDropin) {
		t.Errorf("Merge() modified dropin:\ngot:\n%s\nwant:\n%s", dropin, wantDropin)
	}

	// mutating the merged unit must not leak into the inputs; the
	// surviving option lives in the second Service section (the base's
	// ExecStart was reset away, leaving that section empty)
	merged.Sections[1].Options[0].Value = "/bin/changed"
	if dropin.Sections[0].Options[1].Value != "/bin/b" {
		t.Error("mutating merged unit changed the dropin unit")
	}
}

// TestMergeRealWorldDropin merges the override.conf fixture onto a base
// nginx service, mirroring what systemd does with
// /etc/systemd/system/nginx.service.d/override.conf.
func TestMergeRealWorldDropin(t *testing.T) {
	base := unitOf(
		sectionOf("Unit",
			optionOf("Description", "The nginx HTTP and reverse proxy server"),
			optionOf("After", "network-online.target"),
		),
		sectionOf("Service",
			optionOf("Type", "forking"),
			optionOf("ExecStart", "/usr/sbin/nginx"),
			optionOf("ExecReload", "/usr/sbin/nginx -s reload"),
		),
		sectionOf("Install",
			optionOf("WantedBy", "multi-user.target"),
		),
	)

	src, err := os.ReadFile(filepath.Join("testdata", "override.conf"))
	if err != nil {
		t.Fatal(err)
	}
	dropin, err := Deserialize(bytes.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	merged := Merge(base, dropin)

	wantExec := "/usr/sbin/nginx -c /etc/nginx/custom.conf -g 'daemon on; master_process on;'"
	if got := merged.Values("Service", "ExecStart"); !reflect.DeepEqual(got, []string{wantExec}) {
		t.Errorf("Values(Service, ExecStart) = %v, want [%s]", got, wantExec)
	}
	if got, _ := merged.Value("Service", "LimitNOFILE"); got != "65536" {
		t.Errorf("Value(Service, LimitNOFILE) = %q, want %q", got, "65536")
	}
	if got, _ := merged.Value("Service", "ExecReload"); got != "/usr/sbin/nginx -s reload" {
		t.Errorf("Value(Service, ExecReload) = %q, want the base value", got)
	}
}

func TestUnit_Values(t *testing.T) {
	unit := unitOf(
		sectionOf("Network", optionOf("DNS", "1.1.1.1"), optionOf("DHCP", "no")),
		sectionOf("Address", optionOf("Address", "10.0.0.2/24")),
		sectionOf("Network", optionOf("DNS", "8.8.8.8")),
	)

	if got, want := unit.Values("Network", "DNS"), []string{"1.1.1.1", "8.8.8.8"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Unit.Values(Network, DNS) = %v, want %v", got, want)
	}
	if got := unit.Values("Network", "Gateway"); got != nil {
		t.Errorf("Unit.Values(Network, Gateway) = %v, want nil", got)
	}
	if got := unit.Values("Route", "Gateway"); got != nil {
		t.Errorf("Unit.Values(Route, Gateway) = %v, want nil", got)
	}
}

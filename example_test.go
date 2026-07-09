package systemdconfig_test

import (
	"fmt"
	"log"
	"os"
	"strings"

	systemdconfig "github.com/javadh75/systemd-config"
)

func ExampleDeserialize() {
	const input = `# eth0 with static addressing
[Match]
Name=eth0

[Network]
DNS=1.1.1.1
DNS=8.8.8.8

[Address]
Address=10.0.0.2/24

[Address]
Address=2001:db8::2/64
`
	unit, err := systemdconfig.Deserialize(strings.NewReader(input))
	if err != nil {
		log.Fatal(err)
	}

	// Duplicate options are preserved and addressable.
	fmt.Println(unit.SectionByName("Network").Values("DNS"))

	// Duplicate sections too.
	for _, addr := range unit.SectionsByName("Address") {
		v, _ := addr.Value("Address")
		fmt.Println(v)
	}
	// Output:
	// [1.1.1.1 8.8.8.8]
	// 10.0.0.2/24
	// 2001:db8::2/64
}

func ExampleUnit_Value() {
	const input = `[Service]
ExecStart=
ExecStart=/usr/sbin/nginx
`
	unit, err := systemdconfig.Deserialize(strings.NewReader(input))
	if err != nil {
		log.Fatal(err)
	}

	// Last assignment wins, as in systemd.
	v, ok := unit.Value("Service", "ExecStart")
	fmt.Println(v, ok)
	// Output: /usr/sbin/nginx true
}

func ExampleMerge() {
	base, err := systemdconfig.Deserialize(strings.NewReader(`[Service]
ExecStart=/usr/sbin/nginx
Restart=always
`))
	if err != nil {
		log.Fatal(err)
	}
	// /etc/systemd/system/nginx.service.d/override.conf
	dropin, err := systemdconfig.Deserialize(strings.NewReader(`[Service]
ExecStart=
ExecStart=/usr/sbin/nginx -c /etc/nginx/custom.conf
`))
	if err != nil {
		log.Fatal(err)
	}

	effective := systemdconfig.Merge(base, dropin)

	// The empty assignment reset the base ExecStart; only the
	// drop-in's value remains. Untouched options survive.
	fmt.Println(effective.Values("Service", "ExecStart"))
	fmt.Println(effective.Values("Service", "Restart"))
	// Output:
	// [/usr/sbin/nginx -c /etc/nginx/custom.conf]
	// [always]
}

func ExampleUnit_WriteTo() {
	unit := systemdconfig.NewUnit()
	unit.AddSection("Match").AddOption("Name", "eth0")
	unit.AddSection("Network").AddOption("DHCP", "yes")

	if _, err := unit.WriteTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
	// Output:
	// [Match]
	// Name=eth0
	//
	// [Network]
	// DHCP=yes
}

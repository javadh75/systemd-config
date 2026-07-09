// Package systemdconfig serializes and deserializes systemd config/unit
// files. Unlike go-systemd/unit it supports duplicate sections (e.g. the
// [Address] section in .network files).
package systemdconfig

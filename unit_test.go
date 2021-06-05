package systemdconfig

import (
	"reflect"
	"testing"
)

func TestNewUnit(t *testing.T) {
	tests := []struct {
		name string
		want *Unit
	}{
		{
			"SimpleUnit",
			&Unit{Sections: []*Section{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUnit(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUnit() = %v, want %v", got, tt.want)
			}
		})
	}
}

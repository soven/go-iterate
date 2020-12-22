package go_text_templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_camelToSnakeCase(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"empty", "", ""},
		{"already_snake", "already_snake", "already_snake"},
		{"A", "A", "a"},
		{"AA", "AA", "aa"},
		{"AaAa", "AaAa", "aa_aa"},
		{"HTTPRequest", "HTTPRequest", "http_request"},
		{"BatteryLifeValue", "BatteryLifeValue", "battery_life_value"},
		{"Id0Value", "Id0Value", "id0_value"},
		{"ID0Value", "ID0Value", "id0_value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := camelToSnakeCase(tt.source)
			assert.Equal(t, tt.want, actual)
		})
	}
}

package stringx

import (
	"reflect"
	"testing"
)

func TestTrimSplit(t *testing.T) {
	tests := []struct {
		name string
		in   string
		sep  string
		want []string
	}{
		{
			name: "trims and removes empty parts",
			in:   " 127.0.0.1:6379, , 127.0.0.1:6380 ",
			sep:  ",",
			want: []string{"127.0.0.1:6379", "127.0.0.1:6380"},
		},
		{
			name: "returns empty slice for blank input",
			in:   "   ",
			sep:  ",",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimSplit(tt.in, tt.sep)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("TrimSplit() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

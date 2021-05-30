package fstesting

import (
	"bytes"
	"io"
	"testing"
)

func TestCompareReader(t *testing.T) {
	type args struct {
		a io.Reader
		b io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"same files",
			args{
				bytes.NewReader([]byte("abcdef")),
				bytes.NewReader([]byte("abcdef")),
			},
			true,
			false,
		},
		{
			"different files",
			args{
				bytes.NewReader([]byte("abcdef")),
				bytes.NewReader([]byte("uvwxyz")),
			},
			false,
			false,
		},
		{
			"different length",
			args{
				bytes.NewReader([]byte("abcdefghijkl")),
				bytes.NewReader([]byte("uvwxyz")),
			},
			false,
			false,
		},
		{
			"empty files",
			args{
				bytes.NewReader([]byte("")),
				bytes.NewReader([]byte("")),
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareReader(tt.args.a, tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareReader() = %v, want %v", got, tt.want)
			}
		})
	}
}

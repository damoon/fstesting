package fstesting

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"
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

func mustWriteFile(t *testing.T, fs afero.Fs, filename string, data []byte, perm os.FileMode) {
	err := afero.WriteFile(fs, filename, data, perm)
	if err != nil {
		t.Fatalf("write file %s: %v", filename, err)
	}
}

func TestCompareFile(t *testing.T) {
	fsA := afero.NewMemMapFs()
	mustWriteFile(t, fsA, "/file", []byte("abcdef"), 0644)
	mustWriteFile(t, fsA, "/empty", []byte(""), 0644)

	fsB := afero.NewMemMapFs()
	mustWriteFile(t, fsB, "/filesame", []byte("abcdef"), 0644)
	mustWriteFile(t, fsB, "/permissions", []byte("abcdef"), 0600)
	mustWriteFile(t, fsB, "/fileother", []byte("uvwxyz"), 0644)
	mustWriteFile(t, fsB, "/longer", []byte("abcdefghijkl"), 0644)
	mustWriteFile(t, fsB, "/empty", []byte(""), 0644)

	type args struct {
		fsA   afero.Fs
		fsB   afero.Fs
		pathA string
		pathB string
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
				fsA,
				fsB,
				"/file",
				"/filesame",
			},
			true,
			false,
		},
		{
			"different content",
			args{
				fsA,
				fsB,
				"/file",
				"/fileother",
			},
			false,
			false,
		},
		{
			"different permissions",
			args{
				fsA,
				fsB,
				"/file",
				"/permissions",
			},
			false,
			false,
		},
		{
			"different length",
			args{
				fsA,
				fsB,
				"/file",
				"/longer",
			},
			false,
			false,
		},
		{
			"empty files",
			args{
				fsA,
				fsB,
				"/empty",
				"/empty",
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareFile(tt.args.fsA, tt.args.fsB, tt.args.pathA, tt.args.pathB)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mustMkdirAll(t *testing.T, fs afero.Fs, path string, perm os.FileMode) {
	err := fs.MkdirAll(path, perm)
	if err != nil {
		t.Fatalf("crate directory %s: %v", path, err)
	}
}

func TestCompareDir(t *testing.T) {
	fsA := afero.NewMemMapFs()
	mustMkdirAll(t, fsA, "/same", 0755)
	mustWriteFile(t, fsA, "/same/file", []byte("abcdef"), 0644)

	mustMkdirAll(t, fsA, "/with/subdirectory", 0755)
	mustWriteFile(t, fsA, "/with/subdirectory", []byte("abcdef"), 0644)

	mustMkdirAll(t, fsA, "/empty", 0755)

	fsB := afero.NewMemMapFs()
	mustMkdirAll(t, fsB, "/othersame", 0755)
	mustWriteFile(t, fsB, "/othersame/file", []byte("abcdef"), 0644)

	mustMkdirAll(t, fsB, "/different", 0755)
	mustWriteFile(t, fsB, "/different/file", []byte("uvwxyz"), 0644)

	mustMkdirAll(t, fsB, "/with/subdirectory", 0755)
	mustWriteFile(t, fsB, "/with/subdirectory", []byte("abcdef"), 0644)

	mustMkdirAll(t, fsB, "/empty", 0755)

	type args struct {
		fsA   afero.Fs
		fsB   afero.Fs
		pathA string
		pathB string
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
				fsA,
				fsB,
				"/same",
				"/othersame",
			},
			true,
			false,
		},
		{
			"different files",
			args{
				fsA,
				fsB,
				"/same",
				"/different",
			},
			false,
			false,
		},
		{
			"with subdirectory",
			args{
				fsA,
				fsB,
				"/with",
				"/with",
			},
			true,
			false,
		},
		{
			"empty directories",
			args{
				fsA,
				fsB,
				"/empty",
				"/empty",
			},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareDir(tt.args.fsA, tt.args.fsB, tt.args.pathA, tt.args.pathB)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

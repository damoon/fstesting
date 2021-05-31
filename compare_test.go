package fstesting

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestDiffReader(t *testing.T) {
	type args struct {
		a io.Reader
		b io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		want1   string
		wantErr bool
	}{
		{
			name: "same files",
			args: args{
				bytes.NewReader([]byte("abcdef")),
				bytes.NewReader([]byte("abcdef")),
			},
			want: true,
		},
		{
			name: "different files",
			args: args{
				bytes.NewReader([]byte("abcdef")),
				bytes.NewReader([]byte("uvwxyz")),
			},
			want1: "content differs between abcdef and uvwxyz",
		},
		{
			name: "different length",
			args: args{
				bytes.NewReader([]byte("abcdefghijkl")),
				bytes.NewReader([]byte("uvwxyz")),
			},
			want1: "content differs between abcdefghijkl and uvwxyz",
		},
		{
			name: "empty files",
			args: args{
				bytes.NewReader([]byte("")),
				bytes.NewReader([]byte("")),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := DiffReader(tt.args.a, tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DiffReader() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DiffReader() got1 = %v, want %v", got1, tt.want1)
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

func TestDiffFile(t *testing.T) {
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
		want1   string
		wantErr bool
	}{
		{
			name: "same files",
			args: args{
				fsA,
				fsB,
				"/file",
				"/filesame",
			},
			want: true,
		},
		{
			name: "different content",
			args: args{
				fsA,
				fsB,
				"/file",
				"/fileother",
			},
			want1: "content differs between abcdef and uvwxyz",
		},
		{
			name: "different permissions",
			args: args{
				fsA,
				fsB,
				"/file",
				"/permissions",
			},
			want1: "permissions differ between 0644 and 0600",
		},
		{
			name: "different length",
			args: args{
				fsA,
				fsB,
				"/file",
				"/longer",
			},
			want1: "size differs between 6 and 12",
		},
		{
			name: "empty files",
			args: args{
				fsA,
				fsB,
				"/empty",
				"/empty",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := DiffFile(tt.args.fsA, tt.args.fsB, tt.args.pathA, tt.args.pathB)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DiffFile() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DiffFile() got1 = %v, want %v", got1, tt.want1)
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

func TestDiffDir(t *testing.T) {
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
		want1   string
		wantErr bool
	}{
		{
			name: "same files",
			args: args{
				fsA,
				fsB,
				"/same",
				"/othersame",
			},
			want: true,
		},
		{
			name: "different files",
			args: args{
				fsA,
				fsB,
				"/same",
				"/different",
			},
			want1: "files /same/file and /different/file differ: content differs between abcdef and uvwxyz",
		},
		{
			name: "with subdirectory",
			args: args{
				fsA,
				fsB,
				"/with",
				"/with",
			},
			want: true,
		},
		{
			name: "empty directories",
			args: args{
				fsA,
				fsB,
				"/empty",
				"/empty",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := DiffDir(tt.args.fsA, tt.args.fsB, tt.args.pathA, tt.args.pathB)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DiffDir() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DiffDir() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

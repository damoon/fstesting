package fstesting

import (
	"testing"

	"github.com/spf13/afero"
)

func TestInMemoryCopy(t *testing.T) {
	memFs, err := InMemoryCopy("testdata", "testdata")
	if err != nil {
		t.Errorf("InMemoryCopy() errored: %v", err)
		return
	}

	osFs := afero.NewOsFs()

	same, err := CompareDir(osFs, memFs, "testdata", "testdata")
	if err != nil {
		t.Errorf("CompareDir() returned error: %v", err)
		return
	}

	if !same {
		t.Errorf("CompareDir() returned not equal, but should be a copy")
		return
	}
}

func TestCopyFile(t *testing.T) {
	memFs := afero.NewMemMapFs()
	osFs := afero.NewOsFs()

	err := CopyFile(osFs, memFs, "testdata/copytest/directory/other-file.txt", "file.txt")
	if err != nil {
		t.Errorf("CopyFile() errored: %v", err)
		return
	}

	same, err := CompareFile(osFs, memFs, "testdata/copytest/directory/other-file.txt", "file.txt")
	if err != nil {
		t.Errorf("CompareFile() errored: %v", err)
		return
	}

	if !same {
		t.Errorf("CompareFile() returned not equal, but should be a copy")
		return
	}
}

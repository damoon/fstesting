package fstesting

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/spf13/afero"
)

func CompareDir(fsA, fsB afero.Fs, pathA, pathB string) (bool, error) {
	same, _, err := DiffDir(fsA, fsB, pathA, pathB)
	return same, err
}

func DiffDir(fsA, fsB afero.Fs, pathA, pathB string) (bool, string, error) {
	filesA, err := afero.ReadDir(fsA, pathA)
	if err != nil {
		return false, "", err
	}

	filesB, err := afero.ReadDir(fsB, pathB)
	if err != nil {
		return false, "", err
	}

	fileNamesA := filesnames(filesA)
	fileNamesB := filesnames(filesB)

	if !listsEqual(fileNamesA, fileNamesB) {
		return false, fmt.Sprintf("files differ in %s %v and %s %v", pathA, fileNamesA, pathB, fileNamesB), nil
	}

	for i, fileA := range filesA {
		fileB := filesB[i]

		fileName := fileA.Name()

		filepathA := filepath.Join(pathA, fileName)
		filepathB := filepath.Join(pathB, fileName)

		if fileA.IsDir() && fileB.IsDir() {
			same, diff, err := DiffDir(fsA, fsB, filepathA, filepathB)
			if err != nil {
				return false, "", err
			}

			if !same {
				return false, diff, nil
			}
		}

		if !fileA.IsDir() && !fileB.IsDir() {
			same, err := CompareFile(fsA, fsB, filepathA, filepathB)
			if err != nil {
				return false, "", err
			}

			if !same {
				return false, fmt.Sprintf("files %s and %s differ", filepathA, filepathB), nil
			}
		}
	}

	return true, "", nil
}

func filesnames(infos []fs.FileInfo) []string {
	names := []string{}
	for _, info := range infos {
		names = append(names, info.Name())
	}

	return names
}

func listsEqual(listA, listB []string) bool {
	if len(listA) != len(listB) {
		return false
	}

	for i, item := range listA {
		if item != listB[i] {
			return false
		}
	}

	return true
}

func CompareFile(fsA, fsB afero.Fs, pathA, pathB string) (bool, error) {
	sa, err := fsA.Stat(pathA)
	if err != nil {
		return false, err
	}

	sb, err := fsB.Stat(pathB)
	if err != nil {
		return false, err
	}

	if sa.Size() != sb.Size() {
		return false, nil
	}

	aMod := sa.Mode()
	bMod := sb.Mode()

	if aMod != bMod {
		return false, nil
	}

	fA, err := fsA.Open(pathA)
	if err != nil {
		return false, err
	}
	defer fA.Close()

	fB, err := fsB.Open(pathB)
	if err != nil {
		return false, err
	}
	defer fB.Close()

	return CompareReader(fA, fB)
}

func CompareReader(a, b io.Reader) (bool, error) {
	bufA := make([]byte, 512)
	bufB := make([]byte, 512)

	for {
		_, errA := a.Read(bufA)
		if errA != nil && errA != io.EOF {
			return false, errA
		}

		_, errB := b.Read(bufB)
		if errB != nil && errB != io.EOF {
			return false, errB
		}

		if !bytes.Equal(bufA, bufB) {
			return false, nil
		}

		if errA == io.EOF && errB == io.EOF {
			return true, nil
		}

		if errA == io.EOF || errB == io.EOF {
			return false, nil
		}
	}
}

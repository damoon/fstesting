package fstesting

import (
	"bytes"
	"io"
	"path/filepath"

	"github.com/spf13/afero"
)

func CompareDir(fsA, fsB afero.Fs, pathA, pathB string) (bool, error) {
	filesA, err := afero.ReadDir(fsA, pathA)
	if err != nil {
		return false, err
	}

	filesB, err := afero.ReadDir(fsB, pathB)
	if err != nil {
		return false, err
	}

	if len(filesA) != len(filesB) {
		return false, nil
	}

	for i, fileA := range filesA {
		fileB := filesB[i]

		if fileA.Name() != fileB.Name() {
			return false, nil
		}

		filepathA := filepath.Join(pathA, fileA.Name())
		filepathB := filepath.Join(pathB, fileB.Name())

		if fileA.IsDir() && fileB.IsDir() {
			same, err := CompareDir(fsA, fsB, filepathA, filepathB)
			if err != nil {
				return false, err
			}

			if !same {
				return false, nil
			}
		}

		if !fileA.IsDir() && !fileB.IsDir() {
			same, err := CompareFile(fsA, fsB, filepathA, filepathB)
			if err != nil {
				return false, err
			}

			if !same {
				return false, nil
			}
		}
	}

	return true, nil
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

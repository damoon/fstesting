package fstesting

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func InMemoryCopy(osPath, memPath string) (afero.Fs, error) {
	memFs := afero.NewMemMapFs()
	osFs := afero.NewOsFs()

	err := CopyDir(osFs, memFs, osPath, memPath)
	if err != nil {
		return memFs, err
	}

	return memFs, nil
}

func CopyDir(srcFs, dstFs afero.Fs, src, dst string) error {
	si, err := srcFs.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	found, err := afero.Exists(dstFs, dst)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("destination %s exists already", dst)
	}

	err = dstFs.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := afero.ReadDir(srcFs, src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Symlink
		if entry.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("copy symlink %s: symlinks are not generally supported for afero filesystem", srcPath)
		}

		if entry.IsDir() {
			err = CopyDir(srcFs, dstFs, srcPath, dstPath)
			if err != nil {
				return err
			}

			continue
		}

		err = CopyFile(srcFs, dstFs, srcPath, dstPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func CopyFile(srcFs, dstFs afero.Fs, src, dst string) error {
	in, err := srcFs.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	si, err := srcFs.Stat(src)
	if err != nil {
		return err
	}

	out, err := dstFs.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, si.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Close()
	if err != nil {
		return err
	}

	return nil
}

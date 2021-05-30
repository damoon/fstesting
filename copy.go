package fstesting

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func InMemoryCopy(osPath string) (afero.Fs, error) {
	appFs := afero.NewMemMapFs()

	err := CopyDir(afero.NewOsFs(), appFs, osPath, osPath)
	if err != nil {
		return appFs, err
	}

	return appFs, nil
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

func CopyFile(srcFs, dstFs afero.Fs, src, dst string) (err error) {
	in, err := srcFs.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := dstFs.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	si, err := srcFs.Stat(src)
	if err != nil {
		return
	}

	err = dstFs.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	return
}

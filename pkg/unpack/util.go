package unpack

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alauda/kube-supv/pkg/utils/untar"
	"github.com/pkg/errors"
)

func ReadFile(srcRoot, path string) (io.ReadCloser, error) {
	info, err := os.Stat(srcRoot)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		path = filepath.FromSlash(path)
		path = filepath.Join(srcRoot, path)
		return os.Open(filepath.Join(srcRoot, path))
	}
	return untar.ReadFile(srcRoot, path)
}

func MakeParentDir(path string) error {
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, fs.FileMode(0755)); err != nil {
				return errors.Wrapf(err, `make dir "%s"`, dir)
			}
		} else {
			return errors.Wrapf(err, `stat "%s"`, dir)
		}
	}
	if !info.IsDir() {
		return fmt.Errorf(`%s is not a directory`, dir)
	}
	return nil
}

func FindFileBySrc(files []File, src string) *File {
	for i, n := 0, len(files); i < n; i++ {
		if files[i].Src == src {
			return &files[i]
		}
	}
	return nil
}

func FindInstallFileByDest(installFiles []InstallFile, dest string) *InstallFile {
	for i, n := 0, len(installFiles); i < n; i++ {
		if installFiles[i].Dest == dest {
			return &installFiles[i]
		}
	}
	return nil
}

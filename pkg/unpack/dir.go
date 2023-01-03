package unpack

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func init() {
	AddInstallerFactory(Directory, NewDirInstaller)
}

type dirInstaller struct {
	destRoot string
}

func NewDirInstaller(srcRoot, destRoot string, values map[string]interface{}) Installer {
	return &dirInstaller{
		destRoot: filepath.FromSlash(destRoot),
	}
}

func (i *dirInstaller) Install(f *File) (*InstallFile, error) {
	if f.Type != Directory {
		return nil, fmt.Errorf(`need FileType "%s", but got "%s"`, Directory, f.Type)
	}

	dest := filepath.Join(i.destRoot, filepath.FromSlash(f.Dest))
	info, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dest, fs.FileMode(0755)); err != nil {
				return nil, errors.Wrapf(err, `make all dir for "%s"`, dest)
			}
		} else {
			return nil, errors.Wrapf(err, `stat "%s"`, dest)
		}
	}
	if !info.IsDir() {
		return nil, fmt.Errorf(`"%s" is not a directory`, dest)
	}

	if f.Mode != 0 {
		mode := f.Mode & os.ModePerm
		if err := os.Chmod(dest, mode); err != nil {
			return nil, fmt.Errorf(`chmod "%s" to %v`, dest, mode)
		}
	}

	if err := os.Chown(dest, f.Uid, f.Gid); err != nil {
		return nil, fmt.Errorf(`chown "%s" to %d:%d`, dest, f.Uid, f.Gid)
	}
	return &InstallFile{
		Dest:         dest,
		Type:         f.Type,
		Uid:          f.Uid,
		Gid:          f.Gid,
		Mode:         f.Mode,
		DeletePolicy: f.DeletePolicy,
	}, nil
}

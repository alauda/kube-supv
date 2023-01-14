package unpack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alauda/kube-supv/pkg/utils"
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
	if err := utils.MakeDir(dest); err != nil {
		return nil, err
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

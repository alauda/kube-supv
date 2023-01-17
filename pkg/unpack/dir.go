package unpack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/pkg/errors"
)

func init() {
	AddInstallerFactory(Directory, NewDirInstaller)
}

type dirInstaller struct {
	fileInstaller
}

func NewDirInstaller(m *Manifest, destRoot string) Installer {
	return &dirInstaller{
		fileInstaller: fileInstaller{
			srcRoot:  filepath.FromSlash(m.srcRoot),
			destRoot: filepath.FromSlash(destRoot),
		},
	}
}

func (i *dirInstaller) Install(f *File) ([]InstallFile, error) {
	var r []InstallFile

	if f.Type != Directory {
		return nil, fmt.Errorf(`need FileType "%s", but got "%s"`, Directory, f.Type)
	}

	destDir := filepath.Join(i.destRoot, filepath.FromSlash(f.Dest))
	destDirExist, err := utils.IsDirExist(destDir)
	if err != nil {
		return nil, err
	}

	uid := os.Getuid()
	gid := os.Getgid()
	if f.Uid != nil {
		uid = *f.Uid
	}
	if f.Gid != nil {
		gid = *f.Gid
	}

	if !destDirExist {
		if err := utils.MakeDir(destDir); err != nil {
			return nil, err
		}
	}

	r = append(r, InstallFile{
		Dest:         destDir,
		Type:         f.Type,
		Uid:          uid,
		Gid:          gid,
		Mode:         f.Mode,
		DeletePolicy: f.DeletePolicy,
	})

	if f.Src != "" {
		srcDir := filepath.Join(i.srcRoot, filepath.FromSlash(f.Src))
		r2, err := i.copyFiles(srcDir, destDir)
		r = append(r, r2...)
		if err != nil {
			return r, err
		}
	}

	if err := utils.ChOwnMod(destDir, uid, gid, f.Mode); err != nil {
		return r, err
	}
	return r, nil
}

func (i *dirInstaller) copyFiles(srcPath, destPath string) ([]InstallFile, error) {
	var r []InstallFile

	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return r, errors.Wrapf(err, `read dir "%s"`, srcPath)
	}
	for _, entry := range entries {
		srcFile := filepath.Join(srcPath, entry.Name())
		destFile := filepath.Join(destPath, entry.Name())
		if entry.IsDir() {
			r2, err := i.copyFiles(srcFile, destFile)
			r = append(r, r2...)
			if err != nil {
				return r, err
			}
		} else {
			uid, gid, mode, hash, err := utils.CopyFile(destFile, srcFile)
			if err != nil {
				return r, err
			}

			r = append(r, InstallFile{
				Dest:         destFile,
				Type:         NormalFile,
				Uid:          uid,
				Gid:          gid,
				Mode:         mode,
				Hash:         hash,
				DeletePolicy: DeletePolicyDelete,
			})
		}
	}
	return r, nil
}

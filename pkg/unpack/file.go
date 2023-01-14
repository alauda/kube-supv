package unpack

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/pkg/errors"
)

func init() {
	AddInstallerFactory(NormalFile, NewFileInstaller)
}

type fileInstaller struct {
	srcRoot  string
	destRoot string
}

func NewFileInstaller(srcRoot, destRoot string, values map[string]interface{}) Installer {
	return &fileInstaller{
		srcRoot:  filepath.FromSlash(srcRoot),
		destRoot: filepath.FromSlash(destRoot),
	}
}

func (i *fileInstaller) Install(f *File) (*InstallFile, error) {
	if f.Type != NormalFile {
		return nil, fmt.Errorf(`need FileType "%s", but got "%s"`, NormalFile, f.Type)
	}

	return i.InstallWithHandler(f, nil)
}

type srcHandler func(srcReader io.ReadCloser, filename string) (io.ReadCloser, error)

func (i *fileInstaller) InstallWithHandler(f *File, h srcHandler) (installFile *InstallFile, retErr error) {
	var srcReader io.ReadCloser
	var err error

	src := filepath.Join(i.srcRoot, filepath.FromSlash(f.Src))
	srcReader, err = os.Open(src)
	if err != nil {
		retErr = errors.Wrapf(err, `open "%s"`, src)
		return
	}
	defer srcReader.Close()

	if h != nil {
		srcReader, err = h(srcReader, f.Src)
		if err != nil {
			retErr = errors.Wrapf(err, `handle "%s"`, src)
			return
		}
		defer srcReader.Close()
	}

	dest := filepath.Join(i.destRoot, filepath.FromSlash(f.Dest))

	destFile, retErr := utils.OpenFileToWrite(dest, fs.FileMode(0600))
	if retErr != nil {
		return
	}

	defer func() {
		err := destFile.Close()
		if err != nil {
			err = errors.Wrapf(err, `close "%s"`, dest)
		}
		if retErr == nil {
			retErr = err
		}
	}()

	if _, err := io.Copy(destFile, srcReader); err != nil {
		retErr = errors.Wrapf(err, `copy from "%s" to "%s"`, src, dest)
		return
	}

	if _, err := destFile.Seek(0, io.SeekStart); err != nil {
		retErr = errors.Wrapf(err, `seek dest "%s" to begin`, dest)
		return
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, destFile); err != nil {
		retErr = errors.Wrapf(err, `compute hash for "%s"`, dest)
		return
	}
	hashResult := fmt.Sprintf("sha256:%s", hex.EncodeToString(hash.Sum(nil)))

	if f.Mode != 0 {
		mode := f.Mode & os.ModePerm
		if err := destFile.Chmod(mode); err != nil {
			retErr = fmt.Errorf(`chmod "%s" to %v`, dest, mode)
			return
		}
	}

	if err := destFile.Chown(f.Uid, f.Gid); err != nil {
		retErr = fmt.Errorf(`chown "%s" to %d:%d`, dest, f.Uid, f.Gid)
		return
	}

	installFile = &InstallFile{
		Dest:         dest,
		Type:         f.Type,
		Uid:          f.Uid,
		Gid:          f.Gid,
		Mode:         f.Mode,
		Hash:         hashResult,
		DeletePolicy: f.DeletePolicy,
	}
	return
}

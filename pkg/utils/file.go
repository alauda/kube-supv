package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

func IsFileExist(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if info.IsDir() {
		return true, fmt.Errorf(`"%s" is dir`, path)
	}
	return true, nil
}

func IsDirExist(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return true, fmt.Errorf(`"%s" is notdir`, path)
	}
	return true, nil
}

func MakeDir(path string) error {
	exist, err := IsDirExist(path)
	if err != nil {
		return err
	}
	if !exist {
		if err := os.MkdirAll(path, fs.FileMode(0755)); err != nil {
			return errors.Wrapf(err, `make all dir for "%s"`, path)
		}
	}
	return nil
}

func MakeParentDir(path string) error {
	return MakeDir(filepath.Dir(path))
}

func OpenFileToWrite(path string, mode fs.FileMode) (*os.File, error) {
	path = filepath.FromSlash(path)
	if err := MakeParentDir(path); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return nil, errors.Wrapf(err, `open file "%s" to write`, path)
	}
	return file, nil
}

func ChOwn(path string, uid, gid int) error {
	if uid != os.Getuid() || gid != os.Getgid() {
		if err := os.Chown(path, uid, gid); err != nil {
			return errors.Wrapf(err, `chown "%s" to %d:%d`, path, uid, gid)
		}
	}
	return nil
}

func ChOwnMod(path string, uid, gid int, mode fs.FileMode) error {
	mode = mode & os.ModePerm

	if mode != 0 {
		if err := os.Chmod(path, mode); err != nil {
			return errors.Wrapf(err, `chmod "%s" to %v`, path, mode)
		}
	}

	if err := ChOwn(path, uid, gid); err != nil {
		return err
	}

	return nil
}

func CopyFile(dest, src string) (uid, gid int, mode fs.FileMode, hash string, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, `copy from "%s" to from "%s"`, src, dest)
		}
	}()

	info, err := os.Stat(src)
	if err != nil {
		err = errors.Wrapf(err, `start "%s"`, src)
		return
	}
	mode = info.Mode().Perm()

	srcReader, err := os.Open(src)
	if err != nil {
		err = errors.Wrapf(err, `open "%s"`, src)
		return
	}
	defer srcReader.Close()

	if err = func() error {
		destWrite, err := OpenFileToWrite(dest, mode)
		if err != nil {
			return errors.Wrapf(err, `open "%s" to write`, dest)
		}
		defer destWrite.Close()

		if _, err := io.Copy(destWrite, srcReader); err != nil {
			return errors.Wrapf(err, `copy from "%s" to "%s"`, src, dest)
		}

		if _, err := destWrite.Seek(0, io.SeekStart); err != nil {
			return errors.Wrapf(err, `seek dest "%s" to begin`, dest)
		}

		sha := sha256.New()
		if _, err := io.Copy(sha, destWrite); err != nil {
			return errors.Wrapf(err, `compute hash for "%s"`, dest)
		}
		hash = fmt.Sprintf("sha256:%s", hex.EncodeToString(sha.Sum(nil)))
		return nil
	}(); err != nil {
		return
	}

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		uid = int(stat.Uid)
		gid = int(stat.Gid)
		if err = ChOwn(dest, uid, gid); err != nil {
			return
		}
	}
	return
}

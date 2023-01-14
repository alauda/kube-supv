package renderer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

func init() {
	AddFunc("file", &FileFuncs{
		fs: afero.NewOsFs(),
	})
}

type FileFuncs struct {
	fs afero.Fs
}

func (f *FileFuncs) Read(path string) (string, error) {
	bytes, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		return "", errors.Wrapf(err, "read file %s", path)
	}
	return string(bytes), nil
}

func (f *FileFuncs) Stat(path string) (os.FileInfo, error) {
	return f.fs.Stat(path)
}

func (f *FileFuncs) Exists(path string) bool {
	_, err := f.Stat(path)
	return err == nil
}

func (f *FileFuncs) IsDir(path string) bool {
	i, err := f.Stat(path)
	return err == nil && i.IsDir()
}

func (f *FileFuncs) ReadDir(path string) ([]string, error) {
	file, err := f.fs.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return file.Readdirnames(0)
	}
	return nil, fmt.Errorf(`"%s" is not a directory`, path)
}

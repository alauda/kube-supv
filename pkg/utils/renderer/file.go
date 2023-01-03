package renderer

import (
	"fmt"
	"io"
	"os"

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
	inFile, err := f.fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open %s", path)
	}
	defer inFile.Close()
	bytes, err := io.ReadAll(inFile)
	if err != nil {
		err = errors.Wrapf(err, "read failed for %s", path)
		return "", err
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

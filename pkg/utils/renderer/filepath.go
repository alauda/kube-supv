package renderer

import (
	"path/filepath"
)

func init() {
	AddFunc("filepath", &FilePathFuncs{})
}

type FilePathFuncs struct {
}

func (f *FilePathFuncs) Base(path string) string {
	return filepath.Base(path)
}

func (f *FilePathFuncs) Clean(path string) string {
	return filepath.Clean(path)
}

func (f *FilePathFuncs) Dir(path string) string {
	return filepath.Dir(path)
}

func (f *FilePathFuncs) Ext(path string) string {
	return filepath.Ext(path)
}

func (f *FilePathFuncs) FromSlash(path string) string {
	return filepath.FromSlash(path)
}

func (f *FilePathFuncs) IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

func (f *FilePathFuncs) Join(path ...string) string {
	return filepath.Join(path...)
}

func (f *FilePathFuncs) Match(pattern, name string) (matched bool, err error) {
	return filepath.Match(pattern, name)
}

func (f *FilePathFuncs) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
}

func (f *FilePathFuncs) Split(path string) []string {
	dir, file := filepath.Split(path)
	return []string{dir, file}
}

func (f *FilePathFuncs) ToSlash(path string) string {
	return filepath.ToSlash(path)
}

func (f *FilePathFuncs) VolumeName(path string) string {
	return filepath.VolumeName(path)
}

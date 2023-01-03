package untar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

func ReadFile(tgzFile, path string) (io.ReadCloser, error) {
	r, err := os.Open(tgzFile)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return readFile(r, path)
}

func readFile(r io.Reader, path string) (io.ReadCloser, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("requires gzip-compressed body: %v", err)
	}
	tr := tar.NewReader(zr)
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar error: %v", err)
		}
		if f.Name != path {
			continue
		}
		fi := f.FileInfo()
		mode := fi.Mode()
		if mode.IsRegular() {
			return &tarCloseWrapper{zr: zr, tr: tr}, nil
		}
		return nil, fmt.Errorf("file mode is %v, not regular file", mode)
	}
	return nil, fmt.Errorf("can not find path: %s", path)
}

type tarCloseWrapper struct {
	zr *gzip.Reader
	tr *tar.Reader
}

func (w *tarCloseWrapper) Read(p []byte) (n int, err error) {
	return w.tr.Read(p)
}

func (w *tarCloseWrapper) Close() error {
	return w.zr.Close()
}

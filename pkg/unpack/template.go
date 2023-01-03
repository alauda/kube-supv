package unpack

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"text/template"

	"github.com/alauda/kube-supv/pkg/utils/renderer"
	"github.com/pkg/errors"
)

func init() {
	AddInstallerFactory(Template, NewTemplateInstaller)
}

type templateInstaller struct {
	fileInstaller
	values   map[string]interface{}
	renderer *template.Template
}

func NewTemplateInstaller(srcRoot, destRoot string, values map[string]interface{}) Installer {
	return &templateInstaller{
		fileInstaller: fileInstaller{
			srcRoot:  filepath.FromSlash(srcRoot),
			destRoot: filepath.FromSlash(destRoot),
		},
		values:   values,
		renderer: renderer.NewRenderer(),
	}
}

func (i *templateInstaller) Install(f *File) (*InstallFile, error) {
	if f.Type != Template {
		return nil, fmt.Errorf(`need FileType "%s", but got "%s"`, Template, f.Type)
	}

	return i.fileInstaller.InstallWithHandler(f, i.render)
}

func (i *templateInstaller) render(srcReader io.ReadCloser, filename string) (io.ReadCloser, error) {
	template, err := io.ReadAll(srcReader)
	if err != nil {
		return nil, errors.Wrapf(err, `read template "%s"`, filename)
	}
	parsedTempalte, err := i.renderer.Parse(string(template))
	if err != nil {
		return nil, errors.Wrapf(err, `parse template "%s", contents: "%s"`, filename, string(template))
	}

	var buf bytes.Buffer
	if err := parsedTempalte.Execute(&buf, i.values); err != nil {
		return nil, errors.Wrapf(err, `rendering template "%s"`, filename)
	}
	return &wrapBuffer{buf: &buf}, nil
}

type wrapBuffer struct {
	buf *bytes.Buffer
}

func (w *wrapBuffer) Read(p []byte) (n int, err error) {
	return w.buf.Read(p)
}

func (w *wrapBuffer) Close() (err error) {
	return nil
}

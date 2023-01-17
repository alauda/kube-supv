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
	m        *Manifest
	renderer *template.Template
}

type templateData struct {
	Values  map[string]interface{}
	Name    string
	Version string
	Root    string
}

func NewTemplateInstaller(m *Manifest, destRoot string) Installer {
	return &templateInstaller{
		fileInstaller: fileInstaller{
			srcRoot:  filepath.FromSlash(m.srcRoot),
			destRoot: filepath.FromSlash(destRoot),
		},
		m:        m,
		renderer: renderer.NewRenderer(),
	}
}

func (i *templateInstaller) Install(f *File) ([]InstallFile, error) {
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
	if err := parsedTempalte.Execute(&buf, &templateData{
		Name:    i.m.Name,
		Version: i.m.Version,
		Root:    i.destRoot,
		Values:  i.m.Values,
	}); err != nil {
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

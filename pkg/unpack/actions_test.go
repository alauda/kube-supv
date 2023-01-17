package unpack

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/pkg/errors"
)

var (
	contents = map[string]string{
		"manifest.yaml": `
name: test
version: v1.0.0
files:
  - type: dir
    dest: /test/dir0
  - type: file
    src: test.txt
    dest: /test/dir0/test.txt
    mode: 320
  - type: template
    src: test.yaml
    dest: /test/dir0/test.yaml
values:
  key1: value1
  key2:
    key3: value3
hooks:
  beforeInstall:
    script: beforeInstall.sh
  afterInstall:
    script: afterInstall.sh
`,
		"beforeInstall.sh": `#!/bin/bash
ls
`,
		"afterInstall.sh": `#!/bin/bash
pwd
`,
		"test.txt": `test.txt`,
		"test.yaml": `
name: {{ .Name }}
version: {{ .Version }}
root: {{ .Root }}
key1: {{ .Values.key1 }}
key3: {{ .Values.key3 }}
base: {{ filepath.Base .Root }}
files: |
  {{ shell "ls" .Root }}
`,
	}
	templateFormat = `
name: test
version: v1.0.0
root: %s
key1: value1
key3: %s
base: %s
files: |
  test

`
)

func prepare() (string, error) {
	srcDir, err := os.MkdirTemp("", "kubesupv-test-src-")
	if err != nil {
		return "", errors.Wrap(err, `make temporary source directory`)
	}

	for path, content := range contents {
		path = filepath.Join(srcDir, path)
		f, err := utils.OpenFileToWrite(path, fs.FileMode(0600))
		if err != nil {
			os.RemoveAll(srcDir)
			return "", errors.Wrapf(err, `open file %s`, path)
		}
		if _, err := f.WriteString(content); err != nil {
			os.RemoveAll(srcDir)
			return "", errors.Wrapf(err, `write file %s`, path)
		}
		if err := f.Close(); err != nil {
			os.RemoveAll(srcDir)
			return "", errors.Wrapf(err, `close file %s`, path)
		}
	}

	return srcDir, nil
}

func TestInstallOrUpgrade(t *testing.T) {
	srcDir, err := prepare()
	if err != nil {
		t.Errorf(`prepare error: %v`, err)
		t.FailNow()
	}
	defer os.RemoveAll(srcDir)

	rootDir, err := os.MkdirTemp("", "kubesupv-test-root-")
	if err != nil {
		t.Errorf(`make temporary root error: %v`, err)
		t.FailNow()
	}
	defer os.RemoveAll(rootDir)

	recordDir, err := os.MkdirTemp("", "kubesupv-test-record-")
	if err != nil {
		t.Errorf(`make temporary record directory error: %v`, err)
		t.FailNow()
	}
	defer os.RemoveAll(recordDir)

	val3 := "abc"

	values := map[string]interface{}{
		"key3": val3,
	}

	if err := InstallOrUpgrade(srcDir, rootDir, recordDir, "", values); err != nil {
		t.Errorf(`InstallOrUpgrade error: %v`, err)
		t.FailNow()
	}

	manifest, err := LoadManifest(srcDir)
	if err != nil {
		t.Errorf(`load manifest error: %v`, err)
		t.FailNow()
	}

	packages, err := ListRecords(recordDir)
	if err != nil {
		t.Errorf(`ListRecords error: %v`, err)
		t.Fail()
	}
	if len(packages) != 1 {
		t.Errorf(`excpect one installed package, but got %d`, len(packages))
		t.Fail()
	}

	if packages[0].Name != manifest.Name {
		t.Errorf(`excpect installed package name is "%s", but got "%s"`, manifest.Name, packages[0].Name)
		t.Fail()
	}

	if packages[0].Version != manifest.Version {
		t.Errorf(`excpect installed package version is "%s", but got "%s"`, manifest.Version, packages[0].Version)
		t.Fail()
	}

	for _, file := range manifest.Files {
		path := filepath.Join(rootDir, file.Dest)
		switch file.Type {
		case NormalFile, Template:
			info, err := os.Stat(path)
			if err != nil {
				t.Errorf(`stat file "%s" error: %v`, path, err)
				t.Fail()
			}
			if file.Mode != 0 && info.Mode() != file.Mode {
				t.Errorf(`expect file mode is %d, got %d`, file.Mode, info.Mode())
				t.Fail()
			}
			content, err := os.ReadFile(path)
			if err != nil {
				t.Errorf(`read file "%s" error: %v`, path, err)
				t.Fail()
			}

			var expected string
			if file.Type == NormalFile {
				expected = contents[file.Src]
			} else {
				expected = fmt.Sprintf(templateFormat, rootDir, val3, filepath.Base(rootDir))
			}

			if string(content) != expected {
				t.Errorf(`expect the content of "%s" is: %s, but got %s`, path, expected, string(content))
				t.Fail()
			}

		case Directory:
			exist, err := utils.IsDirExist(path)
			if err != nil {
				t.Errorf(`IsDirExist "%s" error: %v`, path, err)
				t.Fail()
			}
			if !exist {
				t.Errorf(`expect "%s" exist, but not`, path)
				t.Fail()
			}
		}
	}

	if err := Uninstall(recordDir, manifest.Name); err != nil {
		t.Errorf(`Delete "%s" error: %v`, manifest.Name, err)
		t.Fail()
	}

	packages, err = ListRecords(recordDir)
	if err != nil {
		t.Errorf(`ListRecords after Delete error: %v`, err)
		t.Fail()
	}
	if len(packages) != 0 {
		t.Errorf(`excpect no installed package, but got %d`, len(packages))
		t.Fail()
	}
}

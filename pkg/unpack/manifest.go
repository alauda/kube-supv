package unpack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alauda/kube-supv/pkg/errarr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const DefaultManifest = "manifest.yaml"

type Manifest struct {
	Name    string                 `yaml:"name"`
	Version string                 `yaml:"version"`
	Files   []File                 `yaml:"files"`
	Values  map[string]interface{} `yaml:"values"`
	Hooks   map[HookType]Hook      `yaml:"hooks"`
	srcRoot string
}

type File struct {
	Type         FileType     `yaml:"type"`
	Src          string       `yaml:"src"`
	Dest         string       `yaml:"dest"`
	Uid          int          `yaml:"uid"`
	Gid          int          `yaml:"gid"`
	Mode         os.FileMode  `yaml:"mode"`
	DeletePolicy DeletePolicy `ymal:"deletePolicy"`
}

type FileType string

const (
	NormalFile FileType = "file"
	Directory  FileType = "dir"
	Template   FileType = "template"
)

type DeletePolicy string

const (
	DeletePolicyKeep    DeletePolicy = "keep"
	DeletePolicyDelete  DeletePolicy = "delete"
	DefaultDeletePolicy DeletePolicy = DeletePolicyDelete
)

func LoadManifest(srcRoot string) (*Manifest, error) {
	srcRoot = filepath.FromSlash(srcRoot)
	manifestPath := filepath.Join(srcRoot, DefaultManifest)
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, `open "%s"`, manifestPath)
	}
	manifest := Manifest{}
	if err := yaml.NewDecoder(file).Decode(&manifest); err != nil {
		return nil, errors.Wrapf(err, `yaml decode "%s"`, manifestPath)
	}
	if manifest.Name == "" {
		return nil, fmt.Errorf(`the name of manifest "%s" is empty`, manifestPath)
	}
	if manifest.Version == "" {
		return nil, fmt.Errorf(`the versioon of manifest "%s" is empty`, manifestPath)
	}

	for i, n := 0, len(manifest.Files); i < n; i++ {
		if manifest.Files[i].DeletePolicy == "" {
			manifest.Files[i].DeletePolicy = DefaultDeletePolicy
		}
		if manifest.Files[i].Dest == "" {
			return nil, fmt.Errorf(`the dest of "%s" in manifest "%s" is empty`, manifest.Files[i].Src, manifestPath)
		}
	}
	for hookType, hook := range manifest.Hooks {
		if hook.Script == "" {
			return nil, fmt.Errorf(`the "%s" hook's script is empty in manifest "%s"`, hookType, manifestPath)
		}
	}

	if len(manifest.Hooks) > 0 {
		for _, hookType := range []HookType{BeforeDelete, AfterDelete} {
			hook, exist := manifest.Hooks[hookType]
			if !exist {
				continue
			}
			if FindFileBySrc(manifest.Files, hook.Script) == nil {
				return nil, fmt.Errorf(`the "%s" hook's script "%s" is not in the files of manifest "%s" is empty`, hookType, hook.Script, manifestPath)

			}
		}
	}

	manifest.srcRoot = srcRoot

	return &manifest, nil
}

func (m *Manifest) Install(destRoot, recordDir string, values map[string]interface{}) (err error) {
	if err := m.runHook(BeforeInstall); err != nil {
		return err
	}
	installers := NewInstallers(m.srcRoot, destRoot, values)
	if err := m.validateFileType(installers); err != nil {
		return err
	}

	record, err := m.installFiles(installers, destRoot, recordDir, values)
	defer func() {
		if err != nil {
			record.Phase = InstallFailed
			record.Message = err.Error()
		}
		if err2 := record.Save(); err2 != nil {
			err = errarr.NewErrors().Append(err, err2)
		}
	}()
	if err != nil {
		return
	}
	if err = m.runHook(AfterInstall); err != nil {
		return
	}
	return
}

func (m *Manifest) Upgrade(destRoot, recordDir string, values map[string]interface{}, oldInstallRecord *InstallRecord) (err error) {
	if err := m.runHook(BeforeUpgrade); err != nil {
		return err
	}
	installers := NewInstallers(m.srcRoot, destRoot, values)
	if err := m.validateFileType(installers); err != nil {
		return err
	}

	record, err := m.installFiles(installers, destRoot, recordDir, values)
	defer func() {
		if err != nil {
			record.Phase = UpgradeFailed
			record.Message = err.Error()
		}
		if err2 := record.Save(); err2 != nil {
			err = errarr.NewErrors().Append(err, err2)
		}
	}()
	if err != nil {
		return
	}
	for i := len(oldInstallRecord.InstallFiles) - 1; i >= 0; i-- {
		oldFile := oldInstallRecord.InstallFiles[i]
		removed := FindInstallFileByDest(record.InstallFiles, oldFile.Dest) == nil
		if removed && oldFile.DeletePolicy != DeletePolicyKeep {
			if err = oldFile.Remove(); err != nil {
				return err
			}
		}
	}

	if err = m.runHook(AfterUpgrade); err != nil {
		return
	}
	return
}

func (m *Manifest) validateFileType(installers map[FileType]Installer) error {
	for _, f := range m.Files {
		_, exist := installers[f.Type]
		if !exist {
			return fmt.Errorf(`unsupported type "%s" for "%s"`, f.Type, f.Src)
		}
	}
	return nil
}

func (m *Manifest) installFiles(installers map[FileType]Installer, destRoot, recordDir string, values map[string]interface{}) (*InstallRecord, error) {
	record := NewInstallRecord(m, recordDir)
	for _, f := range m.Files {
		installFile, err := installers[f.Type].Install(&f)
		if err != nil {
			return nil, err
		}
		record.Append(installFile)
	}
	if len(m.Hooks) > 0 {
		for _, hookType := range []HookType{BeforeDelete, AfterDelete} {
			hook, exist := m.Hooks[hookType]
			if !exist {
				continue
			}
			file := FindFileBySrc(m.Files, hook.Script)
			if file == nil {
				return nil, fmt.Errorf(`the "%s" hook's script "%s" is not in the files of manifest`, hookType, hook.Script)
			}
			record.Hooks[hookType] = Hook{
				Script: file.Dest,
			}
		}
	}

	return record, nil
}

func (m *Manifest) runHook(hookType HookType) error {
	if m.Hooks != nil {
		if hook, exist := m.Hooks[hookType]; exist {
			return hook.Run(m.srcRoot)
		}
	}
	return nil
}

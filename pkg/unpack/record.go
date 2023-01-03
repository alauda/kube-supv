package unpack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	recordFileMode os.FileMode = 0600
)

type InstallRecord struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	InstallFiles []InstallFile     `yaml:"installFiles"`
	Phase        RecordPhase       `yaml:"phase"`
	Message      string            `yaml:"message"`
	Hooks        map[HookType]Hook `yaml:"hooks"`
	recordDir    string
}

type InstallFile struct {
	Dest         string       `yaml:"dest"`
	Type         FileType     `yaml:"type"`
	Uid          int          `yaml:"uid"`
	Gid          int          `yaml:"gid"`
	Mode         os.FileMode  `yaml:"mode"`
	Hash         string       `yaml:"hash"`
	DeletePolicy DeletePolicy `ymal:"deletePolicy"`
}

func (f *InstallFile) Remove() error {
	if f.DeletePolicy == "" {
		f.DeletePolicy = DefaultDeletePolicy
	}
	if f.DeletePolicy == DeletePolicyKeep {
		return nil
	}
	switch f.Type {
	case NormalFile, Template:
		if err := os.Remove(f.Dest); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return errors.Wrapf(err, `remove "%s"`, f.Dest)
		}
	case Directory:
		if err := os.RemoveAll(f.Dest); err != nil {
			return errors.Wrapf(err, `remove dir "%s"`, f.Dest)
		}
	}
	return nil
}

type RecordPhase string

const (
	Success       RecordPhase = "Success"
	InstallFailed RecordPhase = "InstallFailed"
	UpgradeFailed RecordPhase = "UpgradeFailed"
	DeleteFailed  RecordPhase = "DeleteFailed"
)

func NewInstallRecord(manifest *Manifest, recordDir string) *InstallRecord {
	return &InstallRecord{
		Name:      manifest.Name,
		Version:   manifest.Version,
		recordDir: recordDir,
		InstallFiles: []InstallFile{
			{
				Dest:         recordPath(recordDir, manifest.Name),
				Type:         NormalFile,
				Uid:          os.Getuid(),
				Gid:          os.Getegid(),
				Mode:         recordFileMode,
				Hash:         "",
				DeletePolicy: DefaultDeletePolicy,
			},
		},
		Hooks: map[HookType]Hook{},
	}
}

func IsInstalled(recordDir, name string) (bool, error) {
	recordPath := recordPath(recordDir, name)
	_, err := os.Stat(recordPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func LoadInstallRecord(recordDir, name string) (*InstallRecord, error) {
	recordPath := recordPath(recordDir, name)
	file, err := os.Open(recordPath)
	if err != nil {
		return nil, errors.Wrapf(err, `open "%s"`, recordPath)
	}

	record := InstallRecord{}
	if err := json.NewDecoder(file).Decode(&record); err != nil {
		return nil, errors.Wrapf(err, `json decode "%s"`, recordPath)
	}
	if record.Name != name {
		return nil, fmt.Errorf(`name of "%s" is "%s", but need "%s"`, recordPath, record.Name, name)
	}
	record.recordDir = recordDir
	return &record, nil
}

func recordPath(recordDir, name string) string {
	recordDir = filepath.FromSlash(recordDir)
	return filepath.Join(recordDir, fmt.Sprintf("%s.json", name))
}

func (r *InstallRecord) Append(installFile *InstallFile) {
	if installFile != nil {
		r.InstallFiles = append(r.InstallFiles, *installFile)
	}
}

func (r *InstallRecord) Save() error {
	recordDir := filepath.FromSlash(r.recordDir)
	recordPath := filepath.Join(recordDir, fmt.Sprintf("%s.json", r.Name))

	if err := MakeParentDir(recordPath); err != nil {
		return errors.Wrapf(err, `make dir for "%s"`, recordPath)
	}

	data, err := json.Marshal(r)
	if err != nil {
		return errors.Wrapf(err, `marshal install record of "%s" to json`, r.Name)
	}

	if err := os.WriteFile(recordPath, data, recordFileMode); err != nil {
		return errors.Wrapf(err, `write "%s"`, recordPath)
	}
	return nil
}

func (r *InstallRecord) Uninstall() error {
	if err := r.runHook(BeforeDelete); err != nil {
		return err
	}
	for i := len(r.InstallFiles) - 1; i >= 0; i-- {
		if err := r.InstallFiles[i].Remove(); err != nil {
			return err
		}
	}
	if err := r.runHook(AfterDelete); err != nil {
		return err
	}
	return nil
}

func (r *InstallRecord) runHook(hookType HookType) error {
	if r.Hooks != nil {
		if hook, exist := r.Hooks[hookType]; exist {
			return hook.Run("")
		}
	}
	return nil
}

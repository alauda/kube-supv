package unpack

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	recordFileMode    os.FileMode = 0600
	DefaultRecordDir              = "/var/lib/kubesupv"
	recordFileExtName             = ".yaml"
)

type InstallRecord struct {
	Name          string                 `yaml:"name"`
	Version       string                 `yaml:"version"`
	Image         string                 `yaml:"image"`
	Files         []InstallFile          `yaml:"files"`
	Phase         RecordPhase            `yaml:"phase"`
	Message       string                 `yaml:"message"`
	Hooks         map[HookType]Hook      `yaml:"hooks"`
	ImageValues   map[string]interface{} `yaml:"imageValues"`
	PackageValues map[string]interface{} `yaml:"packageValues"`
	NodeValues    map[string]interface{} `yaml:"nodeValues"`
	Values        map[string]interface{} `yaml:"values"`
	InstallRoot   string                 `yaml:"installRoot"`
	Histories     []InstallHistory       `yaml:"histories"`
	recordDir     string
}

type InstallHistory struct {
	Version string
	Phase   RecordPhase `yaml:"phase"`
	Message string      `yaml:"message"`
	Time    string      `yaml:"time"`
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
		f.DeletePolicy = DefaultDeletePolicy(f.Type)
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
	InstallSuccess RecordPhase = "Success"
	InstallFailed  RecordPhase = "InstallFailed"
	UpgradeFailed  RecordPhase = "UpgradeFailed"
	DeleteFailed   RecordPhase = "DeleteFailed"
)

func NewInstallRecord(manifest *Manifest, destRoot, recordDir, image string) *InstallRecord {
	return &InstallRecord{
		Name:      manifest.Name,
		Version:   manifest.Version,
		Image:     image,
		recordDir: recordDir,
		Files: []InstallFile{
			{
				Dest:         recordPath(recordDir, manifest.Name),
				Type:         NormalFile,
				Uid:          os.Getuid(),
				Gid:          os.Getegid(),
				Mode:         recordFileMode,
				Hash:         "",
				DeletePolicy: DeletePolicyDelete,
			},
		},
		Hooks:       map[HookType]Hook{},
		InstallRoot: destRoot,
		Values:      manifest.Values,
	}
}

func ListRecords(recordDir string) ([]InstallRecord, error) {
	recordDir = filepath.FromSlash(recordDir)
	dirEntries, err := os.ReadDir(recordDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, `read directory "%s"`, recordDir)
	}
	var r []InstallRecord
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if filepath.Ext(fileName) == recordFileExtName {
			name := fileName[:len(fileName)-len(recordFileExtName)]
			record, exist, err := LoadInstallRecord(recordDir, name)
			if err != nil || !exist {
				continue
			}
			r = append(r, *record)
		}
	}
	return r, nil
}

func LoadInstallRecord(recordDir, name string) (*InstallRecord, bool, error) {
	recordPath := recordPath(recordDir, name)
	file, err := os.Open(recordPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, errors.Wrapf(err, `open "%s"`, recordPath)
	}

	record := InstallRecord{}
	if err := record.Decode(file); err != nil {
		return nil, true, errors.Wrapf(err, `decode "%s"`, recordPath)
	}
	if record.Name != name {
		return nil, true, fmt.Errorf(`name of "%s" is "%s", but need "%s"`, recordPath, record.Name, name)
	}
	record.recordDir = recordDir
	return &record, true, nil
}

func recordPath(recordDir, name string) string {
	recordDir = filepath.FromSlash(recordDir)
	return filepath.Join(recordDir, fmt.Sprintf("%s%s", name, recordFileExtName))
}

func (r *InstallRecord) Append(installFiles ...InstallFile) {
	if len(installFiles) > 0 {
		r.Files = append(r.Files, installFiles...)
	}
}

func (r *InstallRecord) Save() error {
	recordDir := filepath.FromSlash(r.recordDir)
	recordPath := recordPath(recordDir, r.Name)

	if err := utils.MakeParentDir(recordPath); err != nil {
		return errors.Wrapf(err, `make dir for "%s"`, recordPath)
	}

	f, err := utils.OpenFileToWrite(recordPath, recordFileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := r.Encode(f); err != nil {
		return err
	}

	return nil
}

func (r *InstallRecord) Encode(w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)

	if err := encoder.Encode(r); err != nil {
		return errors.Wrapf(err, `marshal install record of "%s"`, r.Name)
	}
	return nil
}

func (r *InstallRecord) Decode(in io.Reader) error {
	decoder := yaml.NewDecoder(in)

	if err := decoder.Decode(&r); err != nil {
		return errors.Wrapf(err, `unmarshal install record`)
	}
	return nil
}

func (r *InstallRecord) Uninstall() error {
	if err := r.runHook(BeforeUninstall); err != nil {
		return err
	}
	for i := len(r.Files) - 1; i >= 0; i-- {
		if err := r.Files[i].Remove(); err != nil {
			return err
		}
	}
	return nil
}

func (r *InstallRecord) runHook(hookType HookType) error {
	if r.Hooks != nil {
		if hook, exist := r.Hooks[hookType]; exist {
			return hook.Run(r.InstallRoot, "")
		}
	}
	return nil
}

package unpack

import (
	"fmt"

	"github.com/pkg/errors"
)

func InstallOrUpgrade(srcRoot, destRoot, recordDir, image string, values map[string]interface{}) error {
	manifest, err := LoadManifest(srcRoot)
	if err != nil {
		return err
	}

	oldRecord, _, _ := LoadInstallRecord(recordDir, manifest.Name)

	upgrade := false
	if oldRecord != nil {
		upgrade = oldRecord.Version != manifest.Version
	}

	if upgrade {
		return manifest.Upgrade(destRoot, recordDir, image, values, oldRecord)
	}

	return manifest.Install(destRoot, recordDir, image, values, oldRecord)
}

func Uninstall(recordDir, name string) error {
	installRecord, exist, err := LoadInstallRecord(recordDir, name)
	if err != nil {
		return errors.Wrapf(err, `load install record %s`, name)
	}
	if !exist {
		return fmt.Errorf(`package "%s" does not exist`, name)
	}
	if err := installRecord.Uninstall(); err != nil {
		return errors.Wrapf(err, `uninstall %s`, name)
	}
	return nil
}

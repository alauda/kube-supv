package unpack

func Install(srcRoot, destRoot, recordDir string, values map[string]interface{}) error {
	manifest, err := LoadManifest(srcRoot)
	if err != nil {
		return err
	}
	return manifest.Install(destRoot, recordDir, values)
}

func Upgrade(srcRoot, destRoot, recordDir string, values map[string]interface{}) error {
	manifest, err := LoadManifest(srcRoot)
	if err != nil {
		return err
	}
	oldInstallRecord, err := LoadInstallRecord(recordDir, manifest.Name)
	if err != nil {
		return err
	}
	return manifest.Upgrade(destRoot, recordDir, values, oldInstallRecord)
}

func InstallOrUpgrade(srcRoot, destRoot, recordDir string, values map[string]interface{}) error {
	manifest, err := LoadManifest(srcRoot)
	if err != nil {
		return err
	}
	installed, err := IsInstalled(recordDir, manifest.Name)
	if err != nil {
		return err
	}
	if installed {
		oldInstallRecord, err := LoadInstallRecord(recordDir, manifest.Name)
		if err != nil {
			return err
		}
		return manifest.Upgrade(destRoot, recordDir, values, oldInstallRecord)
	}

	return manifest.Install(destRoot, recordDir, values)
}

func Delete(recordDir, name string) error {
	installRecord, err := LoadInstallRecord(recordDir, name)
	if err != nil {
		return err
	}
	return installRecord.Uninstall()
}

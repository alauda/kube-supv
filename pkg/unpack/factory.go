package unpack

var factories = map[FileType]InstallerFactory{}

type InstallerFactory func(srcRoot, destRoot string, values map[string]interface{}) Installer

type Installer interface {
	Install(*File) (*InstallFile, error)
}

func AddInstallerFactory(fileType FileType, factory InstallerFactory) {
	factories[fileType] = factory
}

func NewInstallers(srcRoot, destRoot string, values map[string]interface{}) map[FileType]Installer {
	installers := map[FileType]Installer{}
	for fileType, factory := range factories {
		installers[fileType] = factory(srcRoot, destRoot, values)
	}
	return installers
}

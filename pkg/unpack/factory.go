package unpack

var factories = map[FileType]InstallerFactory{}

type InstallerFactory func(m *Manifest, destRoot string) Installer

type Installer interface {
	Install(*File) ([]InstallFile, error)
}

func AddInstallerFactory(fileType FileType, factory InstallerFactory) {
	factories[fileType] = factory
}

func NewInstallers(m *Manifest, destRoot string) map[FileType]Installer {
	installers := map[FileType]Installer{}
	for fileType, factory := range factories {
		installers[fileType] = factory(m, destRoot)
	}
	return installers
}

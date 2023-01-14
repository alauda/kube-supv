package unpack

func FindFileBySrc(files []File, src string) *File {
	for i, n := 0, len(files); i < n; i++ {
		if files[i].Src == src {
			return &files[i]
		}
	}
	return nil
}

func FindInstallFileByDest(installFiles []InstallFile, dest string) *InstallFile {
	for i, n := 0, len(installFiles); i < n; i++ {
		if installFiles[i].Dest == dest {
			return &installFiles[i]
		}
	}
	return nil
}

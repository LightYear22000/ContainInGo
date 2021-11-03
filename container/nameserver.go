package container

import(
	"os"
	"ContainInGo/utils"
)

func copyNameserverConfig(containerID string) error {
	resolvFilePaths := []string{
		"/var/run/systemd/resolve/resolv.conf",
		"/etc/gockerresolv.conf",
		"/etc/resolv.conf",
	}
	for _, resolvFilePath := range resolvFilePaths {
		if _, err := os.Stat(resolvFilePath); os.IsNotExist(err) {
			continue
		} else {
			return utils.CopyFile(resolvFilePath,
				GetContainerFSHome(containerID)+"/mnt/etc/resolv.conf")
		}
	}
	return nil
}
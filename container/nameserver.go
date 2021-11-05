package container

import (
	"ContainInGo/utils"
	"log"
	"os"

	"golang.org/x/sys/unix"
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

func unmountNetworkNamespace(containerID string) {
	netNsPath := utils.GetCigNetNsPath() + "/" + containerID
	if err := unix.Unmount(netNsPath, 0); err != nil {
		log.Fatalf("Uable to mount network namespace: %v at %s", err, netNsPath)
	}
}

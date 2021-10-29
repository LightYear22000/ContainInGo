package container

import (
	"log"
	"math/rand"
	"fmt"
	"ContainInGo/image"
	"ContainInGo/utils"
	"ContainInGo/network"
	"os"
)

/* Generate container id */
func generateContainerID() string {
	randBytes := make([]byte, 6)
	rand.Read(randBytes)
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x",
		randBytes[0], randBytes[1], randBytes[2],
		randBytes[3], randBytes[4], randBytes[5])
}

func createDirsIfDontExist(dirs []string) error {
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0755); err != nil {
				log.Printf("Error creating directory: %v\n", err)
				return err
			}
		}
	}
	return nil
}

func createContainerDirectories(containerID string) {
	contHome := utils.GetCigContainersPath() + "/" + containerID
	contDirs := []string{contHome + "/fs", contHome + "/fs/mnt", contHome + "/fs/upperdir", contHome + "/fs/workdir"}
	if err := createDirsIfDontExist(contDirs); err != nil {
		log.Fatalf("Unable to create required directories: %v\n", err)
	}
}

func InitContainer(mem int, swap int, pids int, cpus float64, src string, args []string) {
	containerID := generateContainerID()
	log.Printf("New container ID: %s\n", containerID)
	imageShaHex := image.DownloadImageIfRequired(src)
	fmt.Printf(src + " hash : %v\n", imageShaHex)
	createContainerDirectories(containerID)
	mountOverlayFileSystem(containerID, imageShaHex)
	if err := network.SetupVirtualEthOnHost(containerID); err != nil {
		log.Fatalf("Unable to setup Veth0 on host: %v", err)
	}
}

package container

import (
	"log"
	"math/rand"
	"fmt"
	"ContainInGo/image"
)

/* Generate container id */
func generateContainerID() string {
	randBytes := make([]byte, 6)
	rand.Read(randBytes)
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x",
		randBytes[0], randBytes[1], randBytes[2],
		randBytes[3], randBytes[4], randBytes[5])
}

func InitContainer(mem int, swap int, pids int, cpus float64, src string, args []string) {
	containerID := generateContainerID()
	log.Printf("New container ID: %s\n", containerID)
	imageShaHex := image.DownloadImageIfRequired(src)
	fmt.Printf(src + " hash : %v\n", imageShaHex)
}

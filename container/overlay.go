package container

import (
	"ContainInGo/image"
	"ContainInGo/utils"
	"log"
	"strings"
	"golang.org/x/sys/unix"
)

func getContainerFSHome(contanerID string) string {
	return utils.GetCigContainersPath() + "/" + contanerID + "/fs"
}

func mountOverlayFileSystem(containerID string, imageShaHex string) {
	var srcLayers []string
	pathManifest := image.GetManifestPathForImage(imageShaHex)
	mani := utils.Manifest{}
	utils.ParseManifest(pathManifest, &mani)
	if len(mani) == 0 || len(mani[0].Layers) == 0 {
		log.Fatal("Could not find any layers.")
	}
	if len(mani) > 1 {
		log.Fatal("I don't know how to handle more than one manifest.")
	}

	imageBasePath := image.GetBasePathForImage(imageShaHex)
	for _, layer := range mani[0].Layers {
		srcLayers = append([]string{imageBasePath + "/" + layer[:12] + "/fs"}, srcLayers...)
		//srcLayers = append(srcLayers, imageBasePath + "/" + layer[:12] + "/fs")
	}
	contFSHome := getContainerFSHome(containerID)
	mntOptions := "lowerdir=" + strings.Join(srcLayers, ":") + ",upperdir=" + contFSHome + "/upperdir,workdir=" + contFSHome + "/workdir"
	if err := unix.Mount("none", contFSHome+"/mnt", "overlay", 0, mntOptions); err != nil {
		log.Fatalf("Mount failed: %v\n", err)
	}
}
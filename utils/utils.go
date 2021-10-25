package utils

import (
	"log"
	"os"
)

const cigHomePath = "/var/lib/cig"
const cigTempPath = cigHomePath + "/tmp"
const cigImagesPath = cigHomePath + "/images"
const cigContainersPath = "/var/run/cig/containers"
const cigNetNsPath = "/var/run/cig/net-ns"

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func InitCigDirs() (err error) {
	dirs := []string{cigHomePath, cigTempPath, cigImagesPath, cigContainersPath}
	return createDirsIfDontExist(dirs)
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

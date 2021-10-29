package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
)

const cigHomePath = "/var/lib/cig"
const cigTempPath = cigHomePath + "/tmp"
const cigImagesPath = cigHomePath + "/images"
const cigContainersPath = "/var/run/cig/containers"
const cigNetNsPath = "/var/run/cig/net-ns"

// return cigImagesPath if it exists
func GetCigImagesPath() string {
	return cigImagesPath
}

// return cigTempPath if it exists
func GetCigTempPath() string {
	return cigTempPath
}

// return cigHomepath if it exists
func GetCigHomePath() string {
	return cigHomePath
}

// return cigContainersPath if it exists
func GetCigContainersPath() string {
	return cigContainersPath
}

// return cignetnspath if exits
func GetCigNetNsPath() string {
	return cigNetNsPath
}

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

func ParseManifest(manifestPath string, mani *Manifest) error {
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, mani); err != nil {
		return err
	}

	return nil
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func RemoveLinkIfExists(path string) {
	if _, err := os.Lstat(path); err == nil {
		os.Remove(path)
	}
}

func DeleteFiles(path string) {
	doOrDieWithMsg(os.RemoveAll(path),
		"Unable to file: "+path)
}

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %v\n", err)
	}
}

func doOrDieWithMsg(err error, msg string) {
	if err != nil {
		log.Fatalf("Fatal error: %s: %v\n", msg, err)
	}
}

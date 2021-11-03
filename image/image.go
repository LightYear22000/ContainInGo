package image

import (
	"ContainInGo/utils"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func GetBasePathForImage(imageShaHex string) string {
	return utils.GetCigImagesPath() + "/" + imageShaHex
}

func GetManifestPathForImage(imageShaHex string) string {
	return GetBasePathForImage(imageShaHex) + "/manifest.json"
}

func GetConfigPathForImage(imageShaHex string) string {
	return GetBasePathForImage(imageShaHex) + "/" + imageShaHex + ".json"
}

/*
	Parse Image name and tag name from source
	Example : alphine:latest
*/
func getImageNameAndTag(src string) (string, string) {
	s := strings.Split(src, ":")
	var img, tag string
	if len(s) > 1 {
		img, tag = s[0], s[1]
	} else {
		img = s[0]
		tag = "latest"
	}
	return img, tag
}

/*
* Check if images.json already exits.
* If not, create an empty one.
* Read from imageDBpath and parse the image metadata.

 */

func parseImagesMetadata(idb *utils.ImagesDB) {
	imagesDBPath := utils.GetCigImagesPath() + "/" + "images.json"
	if _, err := os.Stat(imagesDBPath); os.IsNotExist(err) {
		/* If it doesn't exist create an empty DB */
		ioutil.WriteFile(imagesDBPath, []byte("{}"), 0644)
	}
	data, err := ioutil.ReadFile(imagesDBPath)
	if err != nil {
		log.Fatalf("Could not read images DB: %v\n", err)
	}
	if err := json.Unmarshal(data, idb); err != nil {
		log.Fatalf("Unable to parse images DB: %v\n", err)
	}
}

/*
* Check if image already exists by hash, return metadata.
 */

func imageExistsByHash(imageShaHex string) (string, string) {
	idb := utils.ImagesDB{}
	parseImagesMetadata(&idb)
	for imgName, avlImages := range idb {
		for imgTag, imgHash := range avlImages {
			if imgHash == imageShaHex {
				return imgName, imgTag
			}
		}
	}
	return "", ""
}

/*
* Check if image already exists by tag, return metadata.
 */

func imageExistByTag(imgName string, tagName string) (bool, string) {
	idb := utils.ImagesDB{}
	parseImagesMetadata(&idb)
	for k, v := range idb {
		if k == imgName {
			for k, v := range v {
				if k == tagName {
					return true, v
				}
			}
		}
	}
	return false, ""
}

func marshalImageMetadata(idb utils.ImagesDB) {
	fileBytes, err := json.Marshal(idb)
	if err != nil {
		log.Fatalf("Unable to marshall images data: %v\n", err)
	}
	imagesDBPath := utils.GetCigImagesPath() + "/" + "images.json"
	if err := ioutil.WriteFile(imagesDBPath, fileBytes, 0644); err != nil {
		log.Fatalf("Unable to save images DB: %v\n", err)
	}
}

/*
* Store image metadata in images.json
* ubuntu -> unique_hash
* ubuntu:latest -> ubuntu
 */

func storeImageMetadata(image string, tag string, imageShaHex string) {
	idb := utils.ImagesDB{}
	ientry := utils.ImageEntries{}
	parseImagesMetadata(&idb)
	if idb[image] != nil {
		ientry = idb[image]
	}
	ientry[tag] = imageShaHex
	idb[image] = ientry
	marshalImageMetadata(idb)
}

/*
* Download image if required and write it's metadata.
 */

func downloadImage(img v1.Image, imageShaHex string, src string) {
	path := utils.GetCigTempPath() + "/" + imageShaHex
	os.Mkdir(path, 0755)
	path += "/package.tar"
	/* Save the image as a tar file */
	if err := crane.SaveLegacy(img, src, path); err != nil {
		log.Fatalf("saving tarball %s: %v", path, err)
	}
	log.Printf("Successfully downloaded %s\n", src)
}

func DownloadImageIfRequired(src string) string {
	imgName, tagName := getImageNameAndTag(src)
	if downloadNotRequired, imageShaHex := imageExistByTag(imgName, tagName); !downloadNotRequired {
		/* Setup the image we want to pull */
		log.Printf("Downloading metadata for %s:%s, please wait...", imgName, tagName)
		img, err := crane.Pull(strings.Join([]string{imgName, tagName}, ":"))
		if err != nil {
			log.Fatal(err)
		}

		manifest, _ := img.Manifest()
		imageShaHex = manifest.Config.Digest.Hex[:12]
		log.Printf("imageHash: %v\n", imageShaHex)
		log.Println("Checking if image exists under another name...")
		/* Identify cases where ubuntu:latest could be the same as ubuntu:20.04*/
		altImgName, altImgTag := imageExistsByHash(imageShaHex)
		if len(altImgName) > 0 && len(altImgTag) > 0 {
			log.Printf("The image you requested %s:%s is the same as %s:%s\n",
				imgName, tagName, altImgName, altImgTag)
			storeImageMetadata(imgName, tagName, imageShaHex)
			return imageShaHex
		} else {
			log.Println("Image doesn't exist. Downloading...")
			downloadImage(img, imageShaHex, src)
			untarFile(imageShaHex)
			processLayerTarballs(imageShaHex, manifest.Config.Digest.Hex)
			storeImageMetadata(imgName, tagName, imageShaHex)
			/*
				Delete folder containing tarball of image
			*/
			tmpPath := utils.GetCigTempPath() + "/" + imageShaHex
			utils.DeleteFiles(tmpPath)
			return imageShaHex
		}
	} else {
		log.Println("Image already exists. Not downloading.")
		return imageShaHex
	}
}

func ParseContainerConfig(imageShaHex string) utils.ImageConfig {
	imagesConfigPath := GetConfigPathForImage(imageShaHex)
	data, err := ioutil.ReadFile(imagesConfigPath)
	if err != nil {
		utils.LogErr(err)
		log.Fatalf("Could not read image config file")
	}
	// log.Println(data)
	imgConfig := utils.ImageConfig{}
	if err := json.Unmarshal(data, &imgConfig); err != nil {
		log.Fatalf("Unable to parse image config data!")
	}
	return imgConfig
}

package image

import (
	"ContainInGo/utils"
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
)

/*
	Untar tar file pointed by pathTar at the pathDir
*/

func untarFile(imageShaHex string) {
	pathDir := utils.GetCigTempPath() + "/" + imageShaHex
	pathTar := pathDir + "/package.tar"
	if err := untar(pathTar, pathDir); err != nil {
		log.Fatalf("Error untaring file: %v\n", err)
	}
}

/*
	Copy tarball folder structure with all the permissions according to the fileinfo.
*/

func untar(tarball, target string) error {
	hardLinks := make(map[string]string)
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()

		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue

		case tar.TypeLink:
			/* Store details of hard links, which we process finally */
			linkPath := filepath.Join(target, header.Linkname)
			linkPath2 := filepath.Join(target, header.Name)
			hardLinks[linkPath2] = linkPath
			continue

		case tar.TypeSymlink:
			linkPath := filepath.Join(target, header.Name)
			if err := os.Symlink(header.Linkname, linkPath); err != nil {
				if os.IsExist(err) {
					continue
				}
				return err
			}
			continue

		case tar.TypeReg:
			/* Ensure any missing directories are created */
			if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
				os.MkdirAll(filepath.Dir(path), 0755)
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if os.IsExist(err) {
				continue
			}
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tarReader)
			file.Close()
			if err != nil {
				return err
			}

		default:
			log.Printf("Warning: File type %d unhandled by untar function!\n", header.Typeflag)
		}
	}

	/* To create hard links the targets must exist, so we do this finally */
	for k, v := range hardLinks {
		if err := os.Link(v, k); err != nil {
			return err
		}
	}
	return nil
}

/*
	Copy tarball folder structure with all the permissions according to the fileinfo.
*/

func untargz(tarball, target string) error {
	hardLinks := make(map[string]string)
	log.Println(tarball, target)
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	gr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gr.Close()
	tarReader := tar.NewReader(gr)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		info := header.FileInfo()
		path := filepath.Join(target, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue

		case tar.TypeLink:
			/* Store details of hard links, which we process finally */
			linkPath := filepath.Join(target, header.Linkname)
			linkPath2 := filepath.Join(target, header.Name)
			hardLinks[linkPath2] = linkPath
			continue

		case tar.TypeSymlink:
			linkPath := filepath.Join(target, header.Name)
			if err := os.Symlink(header.Linkname, linkPath); err != nil {
				if os.IsExist(err) {
					continue
				}
				return err
			}
			continue

		case tar.TypeReg:
			/* Ensure any missing directories are created */
			if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
				os.MkdirAll(filepath.Dir(path), 0755)
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if os.IsExist(err) {
				continue
			}
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tarReader)
			file.Close()
			if err != nil {
				return err
			}

		default:
			log.Printf("Warning: File type %d unhandled by untar function!\n", header.Typeflag)
		}
	}

	/* To create hard links the targets must exist, so we do this finally */
	for k, v := range hardLinks {
		if err := os.Link(v, k); err != nil {
			return err
		}
	}
	return nil
}

func processLayerTarballs(imageShaHex string, fullImageHex string) {
	tmpPathDir := utils.GetCigTempPath() + "/" + imageShaHex
	pathManifest := tmpPathDir + "/manifest.json"
	pathConfig := tmpPathDir + "/" + fullImageHex + ".json"

	mani := utils.Manifest{}
	utils.ParseManifest(pathManifest, &mani)
	log.Println(mani)
	if len(mani) == 0 || len(mani[0].Layers) == 0 {
		log.Fatal("Could not find any layers.")
	}
	if len(mani) > 1 {
		log.Fatal("I don't know how to handle more than one manifest.")
	}

	imagesDir := utils.GetCigImagesPath() + "/" + imageShaHex
	_ = os.Mkdir(imagesDir, 0755)
	/* untar the layer files. These become the basis of our container root fs */
	for _, layer := range mani[0].Layers {
		imageLayerDir := imagesDir + "/" + layer[:12] + "/fs"
		log.Printf("Uncompressing layer to: %s \n", imageLayerDir)
		_ = os.MkdirAll(imageLayerDir, 0755)
		srcLayer := tmpPathDir + "/" + layer
		if err := untargz(srcLayer, imageLayerDir); err != nil {
			log.Fatalf("Unable to untar layer file: %s: %v\n", srcLayer, err)
		}
	}
	/* Copy the manifest file for reference later */
	utils.CopyFile(pathManifest, getManifestPathForImage(imageShaHex))
	utils.CopyFile(pathConfig, getConfigPathForImage(imageShaHex))
}



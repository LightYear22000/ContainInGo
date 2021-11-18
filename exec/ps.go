package exec

import (
	"ContainInGo/image"
	"ContainInGo/utils"
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"log"
)

/*
		Get the list of running container IDs.

		Implementation logic:
		- Cig creates multiple folders in the /sys/fs/cgroup hierarchy
		- For example, for setting cpu limits, Cig uses /sys/fs/cgroup/cpu/cig
	- Inside that folder are folders one each for currently running containers
	- Those folder names are the container IDs we create.
	- getContainerInfoForId() does more work. It gathers more information about running
		containers. See struct runningContainerInfo for details.
	- Inside each of those folders is a "cgroup.procs" file that has the list
		of PIDs of processes inside of that container. From the PID, we can
		get the mounted path from which the process was started. From that
		mounted path, we can get the image of the containers since containers
		are mounted via the overlay file system.
*/

func getRunningContainers() ([]utils.RunningContainerInfo, error) {
	var containers []utils.RunningContainerInfo
	basePath := "/sys/fs/cgroup/cpu/cig"

	entries, err := ioutil.ReadDir(basePath)
	if os.IsNotExist(err) {
		return containers, nil
	} else {
		if err != nil {
			return nil, err
		} else {
			for _, entry := range entries {
				if entry.IsDir() {
					container, _ := getRunningContainerInfoForId(entry.Name())
					if container.Pid > 0 {
						containers = append(containers, container)
					}
				}
			}
			return containers, nil
		}
	}
}

func getRunningContainerInfoForId(containerID string) (utils.RunningContainerInfo, error) {
	container := utils.RunningContainerInfo{}
	var procs []string
	basePath := "/sys/fs/cgroup/cpu/cig"

	file, err := os.Open(basePath + "/" + containerID + "/cgroup.procs")
	if err != nil {
		fmt.Println("Unable to read cgroup.procs")
		return container, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		procs = append(procs, scanner.Text())
	}
	if len(procs) > 0 {
		pid, err := strconv.Atoi(procs[len(procs)-1])
		if err != nil {
			fmt.Println("Unable to read PID")
			return container, err
		}
		cmd, err := os.Readlink("/proc/" + strconv.Itoa(pid) + "/exe")
		containerMntPath := utils.GetCigContainersPath() + "/" + containerID + "/fs/mnt"
		realContainerMntPath, err := filepath.EvalSymlinks(containerMntPath)
		if err != nil {
			fmt.Println("Unable to resolve path")
			return container, err
		}

		if err != nil {
			fmt.Println("Unable to read command link.")
			return container, err
		}
		image, _ := getDistribution(containerID)
		container = utils.RunningContainerInfo{
			ContainerId: containerID,
			Image:       image,
			Command:     cmd[len(realContainerMntPath):],
			Pid:         pid,
		}
	}
	return container, nil
}

func getDistribution(containerID string) (string, error) {
	var lines []string
	file, err := os.Open("/proc/mounts")
	if err != nil {
		fmt.Println("Unable to read /proc/mounts")
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for _, line := range lines {
		if strings.Contains(line, containerID) {
			parts := strings.Split(line, " ")
			for _, part := range parts {
				if strings.Contains(part, "lowerdir=") {
					options := strings.Split(part, ",")
					for _, option := range options {
						if strings.Contains(option, "lowerdir=") {
							imagesPath := utils.GetCigImagesPath()
							leaderString := "lowerdir=" + imagesPath + "/"
							trailerString := option[len(leaderString):]
							imageID := trailerString[:12]
							image, tag := image.GetImageAndTagForHash(imageID)
							return fmt.Sprintf("%s:%s", image, tag), nil
						}
					}
				}
			}
		}
	}
	return "", nil
}

func PrintRunningContainers() {
	containers, err := getRunningContainers()
	if err != nil {
		os.Exit(1)
	}

	fmt.Println("CONTAINER ID\tIMAGE\t\tCOMMAND")
	for _, container := range containers {
		fmt.Printf("%s\t%s\t%s\n", container.ContainerId, container.Image, container.Command)
	}
}

func DeleteImageByHash(imageShaHex string) {
	// Ensure that no running container is using the image we're setting
	// out to delete. There is a race condition possible here, but we use
	// the ostrich algorithm
	imgName, imgTag := image.GetImageAndTagForHash(imageShaHex)
	if len(imgName) == 0 {
		log.Fatalf("No such image")
	}
	containers, err := getRunningContainers()
	if err != nil {
		log.Fatalf("Unable to get running containers list: %v\n", err)
	}
	for _, container := range containers {
		if container.Image == imgName + ":" + imgTag {
			log.Fatalf("Cannot delete image becuase it is in use by: %s",
						container.ContainerId)
		}
	}

	utils.LogErrWithMsg(os.RemoveAll(utils.GetCigImagesPath() + "/" + imageShaHex),
		"Unable to remove image directory")
	image.RemoveImageMetadata(imageShaHex)
}

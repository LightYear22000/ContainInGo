package exec

import (
	"ContainInGo/container"
	"ContainInGo/image"
	"ContainInGo/utils"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

func getPidForRunningContainer(containerID string) int {
	containers, err := getRunningContainers()
	if err != nil {
		log.Fatalf("Unable to get running containers: %v\n", err)
	}
	for _, container := range containers {
		if container.ContainerId == containerID {
			return container.Pid
		}
	}
	return 0
}

func ExecInContainer(containerId string) {
	pid := getPidForRunningContainer(containerId)
	if pid == 0 {
		log.Fatalf("No such container!")
	}
	baseNsPath := "/proc/" + strconv.Itoa(pid) + "/ns"
	ipcFd, ipcErr := os.Open(baseNsPath + "/ipc")
	mntFd, mntErr := os.Open(baseNsPath + "/mnt")
	netFd, netErr := os.Open(baseNsPath + "/net")
	pidFd, pidErr := os.Open(baseNsPath + "/pid")
	utsFd, utsErr := os.Open(baseNsPath + "/uts")

	if ipcErr != nil || mntErr != nil || netErr != nil ||
		pidErr != nil || utsErr != nil {
		log.Fatalf("Unable to open namespace files!")
	}

	unix.Setns(int(ipcFd.Fd()), unix.CLONE_NEWIPC)
	unix.Setns(int(mntFd.Fd()), unix.CLONE_NEWNS)
	unix.Setns(int(netFd.Fd()), unix.CLONE_NEWNET)
	unix.Setns(int(pidFd.Fd()), unix.CLONE_NEWPID)
	unix.Setns(int(utsFd.Fd()), unix.CLONE_NEWUTS)

	containerConfig, err := getRunningContainerInfoForId(containerId)
	if err != nil {
		log.Fatalf("Unable to get container configuration")
	}
	imageNameAndTag := strings.Split(containerConfig.Image, ":")
	exists, imageShaHex := image.ImageExistByTag(imageNameAndTag[0], imageNameAndTag[1])
	if !exists {
		log.Fatalf("Unable to get image details")
	}
	imgConfig := image.ParseContainerConfig(imageShaHex)
	containerMntPath := utils.GetCigContainersPath() + "/" + containerId + "/fs/mnt"
	container.CreateCGroups(containerId, false)
	utils.LogErrWithMsg(unix.Chroot(containerMntPath), "Unable to chroot")
	os.Chdir("/")
	cmd := exec.Command(os.Args[3], os.Args[4:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = imgConfig.Config.Env
	utils.LogErrWithMsg(cmd.Run(), "Unable to exec command in container")
}

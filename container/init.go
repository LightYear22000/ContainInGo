package container

import (
	"ContainInGo/image"
	"ContainInGo/network"
	"ContainInGo/utils"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"

	"golang.org/x/sys/unix"
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

func prepareAndExecuteContainer(mem int, swap int, pids int, cpus float64,
	containerID string, imageShaHex string, cmdArgs []string) {

	/* Setup the network namespace  */
	cmd := &exec.Cmd{
		Path:   "/proc/self/exe",
		Args:   []string{"/proc/self/exe", "setup-netns", containerID},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	cmd.Run()

	/* Namespace and setup the virtual interface  */
	cmd = &exec.Cmd{
		Path:   "/proc/self/exe",
		Args:   []string{"/proc/self/exe", "setup-veth", containerID},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	cmd.Run()
	/*
		From namespaces(7)
		       Namespace Flag            Isolates
		       --------- ----   		 --------
		       Cgroup    CLONE_NEWCGROUP Cgroup root directory
		       IPC       CLONE_NEWIPC    System V IPC,
		                                 POSIX message queues
		       Network   CLONE_NEWNET    Network devices,
		                                 stacks, ports, etc.
		       Mount     CLONE_NEWNS     Mount points
		       PID       CLONE_NEWPID    Process IDs
		       Time      CLONE_NEWTIME   Boot and monotonic
		                                 clocks
		       User      CLONE_NEWUSER   User and group IDs
		       UTS       CLONE_NEWUTS    Hostname and NIS
		                                 domain name
	*/
	var opts []string
	if mem > 0 {
		opts = append(opts, "--mem="+strconv.Itoa(mem))
	}
	if swap >= 0 {
		opts = append(opts, "--swap="+strconv.Itoa(swap))
	}
	if pids > 0 {
		opts = append(opts, "--pids="+strconv.Itoa(pids))
	}
	if cpus > 0 {
		opts = append(opts, "--cpus="+strconv.FormatFloat(cpus, 'f', 1, 64))
	}
	opts = append(opts, "--img="+imageShaHex)
	args := append([]string{containerID}, cmdArgs...)
	args = append(opts, args...)
	args = append([]string{"child-mode"}, args...)
	cmd = exec.Command("/proc/self/exe", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &unix.SysProcAttr{
		Cloneflags: unix.CLONE_NEWPID |
			unix.CLONE_NEWNS |
			unix.CLONE_NEWUTS |
			unix.CLONE_NEWIPC,
	}
	utils.LogErr(cmd.Run())
}

func ExecContainerCommand(mem int, swap int, pids int, cpus float64,
	containerID string, imageShaHex string, args []string) {
	mntPath := GetContainerFSHome(containerID) + "/mnt"
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	imgConfig := image.ParseContainerConfig(imageShaHex)
	utils.LogErrWithMsg(unix.Sethostname([]byte(containerID)), "Unable to set hostname")
	utils.LogErrWithMsg(network.JoinContainerNetworkNamespace(containerID), "Unable to join container network namespace")
	CreateCGroups(containerID, true)
	ConfigureCGroups(containerID, mem, swap, pids, cpus)
	utils.LogErrWithMsg(copyNameserverConfig(containerID), "Unable to copy resolve.conf")
	utils.LogErrWithMsg(unix.Chroot(mntPath), "Unable to chroot")
	utils.LogErrWithMsg(os.Chdir("/"), "Unable to change directory")
	createDirsIfDontExist([]string{"/proc", "/sys"})
	utils.LogErrWithMsg(unix.Mount("proc", "/proc", "proc", 0, ""), "Unable to mount proc")
	utils.LogErrWithMsg(unix.Mount("tmpfs", "/tmp", "tmpfs", 0, ""), "Unable to mount tmpfs")
	utils.LogErrWithMsg(unix.Mount("tmpfs", "/dev", "tmpfs", 0, ""), "Unable to mount tmpfs on /dev")
	createDirsIfDontExist([]string{"/dev/pts"})
	utils.LogErrWithMsg(unix.Mount("devpts", "/dev/pts", "devpts", 0, ""), "Unable to mount devpts")
	utils.LogErrWithMsg(unix.Mount("sysfs", "/sys", "sysfs", 0, ""), "Unable to mount sysfs")
	network.SetupLocalInterface()
	cmd.Env = imgConfig.Config.Env
	cmd.Run()
	utils.LogErr(unix.Unmount("/dev/pts", 0))
	utils.LogErr(unix.Unmount("/dev", 0))
	utils.LogErr(unix.Unmount("/sys", 0))
	utils.LogErr(unix.Unmount("/proc", 0))
	utils.LogErr(unix.Unmount("/tmp", 0))
}

func InitContainer(mem int, swap int, pids int, cpus float64, src string, args []string) {
	containerID := generateContainerID()
	log.Printf("New container ID: %s\n", containerID)
	imageShaHex := image.DownloadImageIfRequired(src)
	fmt.Printf(src+" hash : %v\n", imageShaHex)
	createContainerDirectories(containerID)
	mountOverlayFileSystem(containerID, imageShaHex)
	if err := network.SetupVirtualEthOnHost(containerID); err != nil {
		log.Fatalf("Unable to setup Veth0 on host: %v", err)
	}
	prepareAndExecuteContainer(mem, swap, pids, cpus, containerID, imageShaHex, args)
	log.Printf("Container done.\n")
	unmountNetworkNamespace(containerID)
	unmountContainerFs(containerID)
	removeCGroups(containerID)
	os.RemoveAll(utils.GetCigContainersPath() + "/" + containerID)
}

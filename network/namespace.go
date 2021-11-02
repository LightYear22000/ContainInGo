package network

import (
	"ContainInGo/utils"
	"golang.org/x/sys/unix"
	"log"
	"github.com/vishvananda/netlink"
	"net"
)

/*
	* Set up a new network namespace for the  current process and mounts it.
*/
func SetupNewNetworkNamespace(containerID string) {
	_ = utils.CreateDirsIfDontExist([]string{utils.GetCigContainersPath()})
	nsMount := utils.GetCigNetNsPath() + "/" + containerID
	if _, err := unix.Open(nsMount, unix.O_RDONLY|unix.O_CREAT|unix.O_EXCL, 0644); err != nil {
		log.Fatalf("Unable to open bind mount file: :%v\n", err)
	}

	fd, err := unix.Open("/proc/self/ns/net", unix.O_RDONLY, 0)
	defer unix.Close(fd)
	if err != nil {
		log.Fatalf("Unable to open: %v\n", err)
	}

	if err := unix.Unshare(unix.CLONE_NEWNET); err != nil {
		log.Fatalf("Unshare system call failed: %v\n", err)
	}
	if err := unix.Mount("/proc/self/ns/net", nsMount, "bind", unix.MS_BIND, ""); err != nil {
		log.Fatalf("Mount system call failed: %v\n", err)
	}
	if err := unix.Setns(fd, unix.CLONE_NEWNET); err != nil {
		log.Fatalf("Setns system call failed: %v\n", err)
	}
}

/*
	*  We connect veth0 part of the pair to our bridge, cig0 bridge on the host. 
	*  Later, we will use veth1 part of the pair inside the container. 
	*  This pair is like a pipe and is the secret to network communication from within 
	*  containers which have their own network namespace.
*/

func SetupContainerNetworkInterfaceStep1(containerID string) {
	nsMount := utils.GetCigNetNsPath() + "/" + containerID

	fd, err := unix.Open(nsMount, unix.O_RDONLY, 0)
	defer unix.Close(fd)
	if err != nil {
		log.Fatalf("Unable to open: %v\n", err)
	}
	/* Set veth1 of the new container to the new network namespace */
	veth1 := "veth1_" + containerID[:6]
	veth1Link, err := netlink.LinkByName(veth1)
	if err != nil {
		log.Fatalf("Unable to fetch veth1: %v\n", err)
	}
	if err := netlink.LinkSetNsFd(veth1Link, fd); err != nil {
		log.Fatalf("Unable to set network namespace for veth1: %v\n", err)
	}
}

func SetupContainerNetworkInterfaceStep2(containerID string) {
	nsMount := utils.GetCigNetNsPath() + "/" + containerID
	fd, err := unix.Open(nsMount, unix.O_RDONLY, 0)
	defer unix.Close(fd)
	if err != nil {
		log.Fatalf("Unable to open: %v\n", err)
	}
	if err := unix.Setns(fd, unix.CLONE_NEWNET); err != nil {
		log.Fatalf("Setns system call failed: %v\n", err)
	}

	veth1 := "veth1_" + containerID[:6]
	veth1Link, err := netlink.LinkByName(veth1)
	if err != nil {
		log.Fatalf("Unable to fetch veth1: %v\n", err)
	}
	addr, _ := netlink.ParseAddr(CreateIPAddress() + "/16")
	if err := netlink.AddrAdd(veth1Link, addr); err != nil {
		log.Fatalf("Error assigning IP to veth1: %v\n", err)
	}

	/* Bring up the interface */
	utils.DoOrDieWithMsg(netlink.LinkSetUp(veth1Link), "Unable to bring up veth1")

	/* Add a default route */
	route := netlink.Route{
		Scope:     netlink.SCOPE_UNIVERSE,
		LinkIndex: veth1Link.Attrs().Index,
		Gw:        net.ParseIP("172.29.0.1"),
		Dst:       nil,
	}
	utils.DoOrDieWithMsg(netlink.RouteAdd(&route), "Unable to add default route")
}
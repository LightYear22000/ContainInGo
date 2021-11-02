package network

import (
	"github.com/vishvananda/netlink"
	"math/rand"
	"net"
	"fmt"
)

func createMACAddress() net.HardwareAddr {
	hw := make(net.HardwareAddr, 6)
	hw[0] = 0x02
	hw[1] = 0x42
	rand.Read(hw[2:])
	return hw
}

func CreateIPAddress() string {
	byte1 := rand.Intn(254)
	byte2 := rand.Intn(254)
	return fmt.Sprintf("172.29.%d.%d", byte1, byte2)
}

func SetupVirtualEthOnHost(containerID string) error {
	veth0 := "veth0_" + containerID[:6]
	veth1 := "veth1_" + containerID[:6]
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = veth0
	veth0Struct := &netlink.Veth{
		LinkAttrs:        linkAttrs,
		PeerName:         veth1,
		PeerHardwareAddr: createMACAddress(),
	}
	if err := netlink.LinkAdd(veth0Struct); err != nil {
		return err
	}
	netlink.LinkSetUp(veth0Struct)
	gockerBridge, _ := netlink.LinkByName("gocker0")
	netlink.LinkSetMaster(veth0Struct, gockerBridge)

	return nil
}
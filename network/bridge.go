package network

import (
	"log"

	"github.com/vishvananda/netlink"
)

/*
	Go through the list of interfaces and return true if the cig0 bridge is up
*/

func IsBridgeUp() (bool, error) {
	if links, err := netlink.LinkList(); err != nil {
		log.Printf("Unable to get list of links.\n")
		return false, err
	} else {
		for _, link := range links {
			if link.Type() == "bridge" && link.Attrs().Name == "cig0" {
				return true, nil
			}
		}
		return false, err
	}
}

/*
	This function sets up the "cig0" bridge, which is our main bridge
	interface. To keep things simple, we assign the hopefully unassigned
	and obscure private IP 172.29.0.1 to it, which is from the range of
	IPs which we will also use for our containers.
*/

func SetupBridge() error {
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = "cig0"
	gockerBridge := &netlink.Bridge{LinkAttrs: linkAttrs}
	if err := netlink.LinkAdd(gockerBridge); err != nil {
		return err
	}
	addr, _ := netlink.ParseAddr("172.29.0.1/16")
	netlink.AddrAdd(gockerBridge, addr)
	netlink.LinkSetUp(gockerBridge)
	return nil
}

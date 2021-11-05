package main

import (
	net "ContainInGo/network"
	"ContainInGo/utils"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"ContainInGo/container"
	flag "github.com/spf13/pflag"
)

func usage() {
	fmt.Println("Welcome to ContainInGo!")
	fmt.Println("Supported commands:")
	fmt.Println("cig run [--mem] [--swap] [--pids] [--cpus] <image> <command>")
	fmt.Println("cig exec <container-id> <command>")
	fmt.Println("cig images")
	fmt.Println("cig rmi <image-id>")
	fmt.Println("cig ps")
}

func main() {
	options := []string{"run", "child-mode", "setup-netns", "setup-veth", "ps", "exec", "images", "rmi"}

	/* Check if arguments are valid */
	if len(os.Args) < 2 || !utils.StringInSlice(os.Args[1], options) {
		usage()
		os.Exit(1)
	}

	/* Seed the random number generator */
	rand.Seed(time.Now().UnixNano())

	/* We chroot and write to privileged directories. We need to be root */
	if os.Geteuid() != 0 {
		log.Fatal("You need root privileges to run this program. Please run again with root privileges")
	}

	/* Create the directories we require */
	if err := utils.InitCigDirs(); err != nil {
		log.Fatalf("Unable to create directories required: %v", err)
	}

	log.Printf("Cmd args: %v\n", os.Args)
	switch os.Args[1] {
	/*
		Case run:
			* takes care of setting up the CIG bridge and downloads the image if required
	*/
	case "run":

		fs := flag.FlagSet{}
		fs.ParseErrorsWhitelist.UnknownFlags = true

		mem := fs.Int("mem", -1, "Max RAM to allow in MB")
		swap := fs.Int("swap", -1, "Max swap to allow in MB")
		pids := fs.Int("pids", -1, "Number of max processes to allow")
		cpus := fs.Float64("cpus", -1, "Number of CPU cores to restrict to")
		if err := fs.Parse(os.Args[2:]); err != nil {
			fmt.Println("Error parsing: ", err)
		}
		if len(fs.Args()) < 2 {
			log.Fatalf("Please pass image name and command to run")
		}
		/* Create and setup the CIG network bridge we need */
		if isUp, _ := net.IsBridgeUp(); !isUp {
			log.Println("Bringing up the cig0 bridge...")
			if err := net.SetupBridge(); err != nil {
				log.Fatalf("Unable to create cig0 bridge: %v", err)
			}
		}
		log.Println("Bridge set up succesfully!")
		container.InitContainer(*mem, *swap, *pids, *cpus, fs.Args()[0], fs.Args()[1:])

	/*
		Setup Network namespace for container.
	*/
	case "setup-netns":
		net.SetupNewNetworkNamespace(os.Args[2])

	/*
		Setup virtual lan interface for the container and assign the ip-address, subnet mask
		and default gateway.
	*/
	case "setup-veth":
		net.SetupContainerNetworkInterfaceStep1(os.Args[2])
		net.SetupContainerNetworkInterfaceStep2(os.Args[2])
	
	case "child-mode":
		fs := flag.FlagSet{}
		fs.ParseErrorsWhitelist.UnknownFlags = true

		mem := fs.Int("mem", -1, "Max RAM to allow in  MB")
		swap := fs.Int("swap", -1, "Max swap to allow in  MB")
		pids := fs.Int("pids", -1, "Number of max processes to allow")
		cpus := fs.Float64("cpus", -1, "Number of CPU cores to restrict to")
		image := fs.String("img", "", "Container image")
		if err := fs.Parse(os.Args[2:]); err != nil {
			fmt.Println("Error parsing: ", err)
		}
		if len(fs.Args()) < 2 {
			log.Fatalf("Please pass image name and command to run")
		}
		container.ExecContainerCommand(*mem, *swap, *pids, *cpus, fs.Args()[0], *image, fs.Args()[1:])
		
	default:
		usage()

	}
}

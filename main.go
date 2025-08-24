package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getMACAddresses() []string {
	var macs []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return macs
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) > 0 {
			macs = append(macs, iface.HardwareAddr.String())
		}
	}
	return macs
}

func runCmd(cmd string, args ...string) string {
	c := exec.Command(cmd, args...)
	var out bytes.Buffer
	c.Stdout = &out
	err := c.Run()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func getMotherboardSerial() string {
	switch runtime.GOOS {
	case "windows":
		return runCmd("wmic", "baseboard", "get", "serialnumber")
	case "linux":
		return runCmd("cat", "/sys/class/dmi/id/board_serial")
	default:
		return ""
	}
}

func getCPUSerial() string {
	switch runtime.GOOS {
	case "windows":
		return runCmd("wmic", "cpu", "get", "ProcessorId")
	case "linux":
		// Try to extract 'Serial' or 'ProcessorId' from /proc/cpuinfo
		output := runCmd("cat", "/proc/cpuinfo")
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "Serial") || strings.Contains(line, "ProcessorId") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
		return "Not available"
	default:
		return ""
	}
}

func getDriveSerial() string {
	switch runtime.GOOS {
	case "windows":
		return runCmd("wmic", "diskdrive", "get", "SerialNumber")
	case "linux":
		// Try udevadm, fallback to lsblk
		output := runCmd("udevadm", "info", "--query=all", "--name=/dev/sda")
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "ID_SERIAL=") {
				parts := strings.Split(line, "=")
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
		// Fallback: try lsblk
		output = runCmd("lsblk", "-o", "NAME,SERIAL")
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "sda") {
				fields := strings.Fields(line)
				if len(fields) == 2 {
					return fields[1]
				}
			}
		}
		return "Not available"
	default:
		return ""
	}
}

func getUUID() string {
	switch runtime.GOOS {
	case "windows":
		return runCmd("wmic", "csproduct", "get", "UUID")
	case "linux":
		return runCmd("cat", "/sys/class/dmi/id/product_uuid")
	default:
		return ""
	}
}

func getManufacturerModel() string {
	switch runtime.GOOS {
	case "windows":
		man := runCmd("wmic", "computersystem", "get", "manufacturer")
		mod := runCmd("wmic", "computersystem", "get", "model")
		return "Manufacturer: " + man + "\nModel: " + mod
	case "linux":
		man := runCmd("cat", "/sys/class/dmi/id/sys_vendor")
		mod := runCmd("cat", "/sys/class/dmi/id/product_name")
		return "Manufacturer: " + man + "\nModel: " + mod
	default:
		return ""
	}
}

func getOSInfo() string {
	if runtime.GOOS == "windows" {
		return runCmd("cmd", "/c", "ver")
	}

	output := runCmd("lsb_release", "-a")
	if output != "" {
		return output
	}
	output = runCmd("cat", "/etc/os-release")
	if output != "" {
		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(line[13:], "\"")
			}
		}
	}
	return "Unknown Linux OS"
}

func main() {
	hostname, _ := os.Hostname()
	fmt.Println("=== Device Fingerprint Proof of Concept ===")

	fmt.Println("MAC Address(es):")
	for _, mac := range getMACAddresses() {
		fmt.Println("  ", mac)
	}

	fmt.Println("\nMotherboard/BIOS Serial Number:")
	fmt.Println("  ", getMotherboardSerial())

	fmt.Println("\nCPU Serial/ID:")
	fmt.Println("  ", getCPUSerial())

	fmt.Println("\nDrive Serial Number:")
	fmt.Println("  ", getDriveSerial())

	fmt.Println("\nHostname:")
	fmt.Println("  ", hostname)

	fmt.Println("\nOS Info:")
	fmt.Println("  ", getOSInfo())

	fmt.Println("\nUUID/Hardware ID:")
	fmt.Println("  ", getUUID())

	fmt.Println("\nSystem Manufacturer/Model:")
	fmt.Println("  ", getManufacturerModel())
}

package main
import (
	"os"
	"path/filepath"
	"encoding/json"
	"fmt"
	"strings"
)

type TFState struct {
	Resources []struct {
		Type      string `json:"type"`
		Instances []struct {
			Attributes map[string]interface{} `json:"attributes"`
		} `json:"instances"`
	} `json:"resources"`
}

type VMInfo struct {
	Name   string `json:"Name"`
	VMID   int    `json:"VMID"`
	IP     string `json:"IP"`
	Memory int    `json:"Memory"`
	Cores  int    `json:"Cores"`
}

func listUserVMs(userID string) ([]VMInfo, error) {
	baseDir := filepath.Join("terraform", userID)

	var vms []VMInfo
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// ディレクトリはそのまま降りる
		if info.IsDir() {
			return nil
		}

		// tfstate だけ処理
		if info.Name() != "terraform.tfstate" {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var state TFState
		if err := json.Unmarshal(b, &state); err != nil {
			return nil
		}

		for _, r := range state.Resources {
			if r.Type != "proxmox_vm_qemu" {
				continue
			}

			for _, inst := range r.Instances {
				attr := inst.Attributes

				vm := VMInfo{
					Name:   fmt.Sprint(attr["name"]),
					VMID:   int(attr["vmid"].(float64)),
					IP:     parseIP(fmt.Sprint(attr["ipconfig0"])),
					Memory: int(attr["memory"].(float64)),
					Cores:  int(attr["cores"].(float64)),
				}

				vms = append(vms, vm)
			}
		}

		return nil
	})

	return vms, err
}

func parseIP(ipconfig string) string {
	for _, part := range strings.Split(ipconfig, ",") {
		if strings.HasPrefix(part, "ip=") {
			ip := strings.TrimPrefix(part, "ip=")
			if ip == "dhcp" {
				return "DHCP"
			}
			return strings.Split(ip, "/")[0]
		}
	}
	return ""
}
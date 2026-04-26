package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var NODE_NAME string
var PORT string

func main() {
	godotenv.Load()
	NODE_NAME = os.Getenv("HOST_NAME")
	PORT = os.Getenv("PORT")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/create-vm", createVMHandler)

	fmt.Println("Server started at http://localhost:" + PORT)
	log.Fatal(http.ListenAndServe(":" + PORT, nil))
}

func createVMHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cpu, _ := strconv.Atoi(r.FormValue("cpu"))
	memory, _ := strconv.Atoi(r.FormValue("memory"))
	hdd, _ := strconv.Atoi(r.FormValue("hdd"))
	servername := r.FormValue("servername")

	// 実行用ディレクトリ作成
	workdir := filepath.Join("terraform", fmt.Sprintf("run_%d", time.Now().Unix()))
	os.MkdirAll(workdir, 0755)

	tfContent := fmt.Sprintf(`
resource "proxmox_virtual_environment_vm" "%s" {
  name      = "%s"
  node_name = "%s"

  cpu {
    cores   = %d
    sockets = 1
    type    = "kvm64"
  }

  memory {
    dedicated = %d
  }

  disk {
    datastore_id = "local-lvm"
    interface    = "scsi0"
    size         = %d
  }

  network_device {
    bridge  = "vmbr0"
    model   = "virtio"
    vlan_id = 20
  }

  operating_system {
    type = "l26"
  }

  agent {
    enabled = true
  }
}
`, servername, servername, NODE_NAME, cpu, memory, hdd)

	tfFile := filepath.Join(workdir, "vm.tf")
	os.WriteFile(tfFile, []byte(tfContent), 0644)

	// provider.tf をコピー
	copyFile("terraform/provider.tf", filepath.Join(workdir, "provider.tf"))

	// Terraform実行
	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = workdir
	initOut, _ := initCmd.CombinedOutput()

	applyCmd := exec.Command("terraform", "apply", "-auto-approve")
	applyCmd.Dir = workdir
	applyOut, _ := applyCmd.CombinedOutput()

	fmt.Fprintf(w, `
	<html>
	<head><meta charset="UTF-8"><title>Terraform実行ログ</title></head>
	<body>
	<h2>Terraform 実行結果</h2>
	<h3>terraform init</h3>
	<pre>%s</pre>
	<h3>terraform apply</h3>
	<pre>%s</pre>
	<a href="/">戻る</a>
	</body>
	</html>
	`, initOut, applyOut)
}

func copyFile(src, dst string) {
	input, err := os.ReadFile(src)
	if err != nil {
		return
	}
	os.WriteFile(dst, input, 0644)
}
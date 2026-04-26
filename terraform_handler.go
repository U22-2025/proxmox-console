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
	"encoding/json"
	"strings"
)

func runTerraformJob(jobID string, r *http.Request) {
	jobAny, _ := jobs.Load(jobID)
	job := jobAny.(*Job)

	cpu, _ := strconv.Atoi(r.FormValue("cpu"))
	memory, _ := strconv.Atoi(r.FormValue("memory"))
	hdd, _ := strconv.Atoi(r.FormValue("hdd"))
	servername := r.FormValue("servername")
	username := r.FormValue("username")
	password := r.FormValue("password")

	hash, err := hashPasswordForLinux(password)
	if err != nil {
		log.Fatal(err)
	}

	// 実行用ディレクトリ作成
	workdir := filepath.Join("terraform", fmt.Sprintf("run_%d", time.Now().Unix()))
	os.MkdirAll(workdir, 0755)

	tfvars := fmt.Sprintf(`
	servername    = "%s"
	cpu           = %d
	memory        = %d
	hdd           = %d
	username      = "%s"
	password_hash = "%s"
	`,
		servername, cpu, memory, hdd, username, hash,
	)

	os.WriteFile(filepath.Join(workdir, "runtime.tfvars"), []byte(tfvars), 0600)
	// ファイルをコピー
	copyFile("terraform/provider.tf", filepath.Join(workdir, "provider.tf"))
	copyFile("terraform/variables.tf", filepath.Join(workdir, "variables.tf"))
	copyFile("terraform/proxmox.auto.tfvars", filepath.Join(workdir, "proxmox.auto.tfvars"))
	copyFile("terraform/snippets.tf", filepath.Join(workdir, "snippets.tf"))
	copyFile("terraform/vm.tf", filepath.Join(workdir, "vm.tf"))
	copyFile("terraform/cloud-config.yaml", filepath.Join(workdir, "cloud-config.yaml"))

	// Terraform実行
	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = workdir
	initOut, err := initCmd.CombinedOutput()
	job.Log += string(initOut)
	if err != nil {
		job.Status = "error"
		return
	}

	applyCmd := exec.Command("terraform", "apply", "-auto-approve", "-var-file=runtime.tfvars")
	applyCmd.Dir = workdir
	applyOut, err := applyCmd.CombinedOutput()
	if err != nil {
		job.Status = "error"
		return
	}

	job.Log = string(initOut) + "\n\n" + string(applyOut)

	if err != nil {
		job.Status = "error"
		return
	}

	job.IP = getVMIP(workdir)
	job.Status = "done"
}

func getVMIP(dir string) string {
	cmd := exec.Command("terraform", "output", "-raw", "vm_ip")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		job.Status = "error"
		return ""
	}
	
	return strings.TrimSpace(string(out))
}

func createVMHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	jobID := fmt.Sprintf("%d", time.Now().UnixNano())

	jobs.Store(jobID, &Job{Status: "running"})

	go runTerraformJob(jobID, r)

	http.Redirect(w, r, "/status.html?id="+jobID, http.StatusSeeOther)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	jobAny, ok := jobs.Load(id)
	if !ok {
		w.WriteHeader(404)
		return
	}

	job := jobAny.(*Job)
	json.NewEncoder(w).Encode(job)
}
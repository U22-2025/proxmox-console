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

func runTerraformJob(jobID string, req VMRequest) {
	// 実行用ディレクトリ作成
	workdir := filepath.Join("terraform", fmt.Sprintf("run_%d", time.Now().Unix()))
	os.MkdirAll(workdir, 0755)

	jobAny, _ := jobs.Load(jobID)
	job := jobAny.(*Job)

	job.LogPath = filepath.Join(workdir, "terraform.log")
    logFile, _ := os.Create(job.LogPath)
    defer logFile.Close()
	job.Status = "running(init)"

	hash, err := hashPasswordForLinux(req.Password)
	if err != nil {
		log.Fatal(err)
	}

	tfvars := fmt.Sprintf(`
	servername    = "%s"
	cpu           = %d
	memory        = %d
	hdd           = %d
	username      = "%s"
	password_hash = "%s"
	`,
		req.Servername, req.CPU, req.Memory, req.HDD, req.Username, hash,
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
	if err := runCmdWithLog(initCmd, logFile); err != nil {
        job.Status = "error"
        return
    }

	job.Status = "running(apply)"
	applyCmd := exec.Command("terraform", "apply", "-auto-approve", "-var-file=runtime.tfvars")
	applyCmd.Dir = workdir
	if err := runCmdWithLog(applyCmd, logFile); err != nil {
        job.Status = "error"
        return
    }

	job.IP = getVMIP(workdir, job)
	job.Status = "done"
}

func getVMIP(dir string, job *Job) string {
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

	req := VMRequest{
		CPU:        atoiSafe(r.FormValue("cpu")),
		Memory:     atoiSafe(r.FormValue("memory")),
		HDD:        atoiSafe(r.FormValue("hdd")),
		Servername: r.FormValue("servername"),
		Username:   r.FormValue("username"),
		Password:   r.FormValue("password"),
	}

	go runTerraformJob(jobID, req)
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
	logBytes, _ := os.ReadFile(job.LogPath)

    resp := map[string]string{
		"status": job.Status,
		"ip":     job.IP,
		"log":    string(logBytes),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func atoiSafe(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
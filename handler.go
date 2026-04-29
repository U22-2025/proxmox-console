package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
	"encoding/json"
	"context"
	tfexec "github.com/hashicorp/terraform-exec/tfexec"
)

func runTerraformJob(jobID string, req *VMRequest) {
	// 実行用ディレクトリ作成
	workdir := filepath.Join("terraform", fmt.Sprintf("run_%d", time.Now().Unix()))
	os.MkdirAll(workdir, 0755)

	jobAny, _ := jobs.Load(jobID)
	job := jobAny.(*Job)

	job.Workdir = workdir
	job.LogPath = filepath.Join(workdir, "terraform.log")
	job.Status = "running(init)"
	jobs.Store(jobID, job)

	logFile, _ := os.Create(job.LogPath)
	defer logFile.Close()

	hash, err := hashPasswordForLinux(req.Password)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		job.Status = "error"
		return
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
	tf, err := tfexec.NewTerraform(workdir, "terraform")
	if err != nil {
		job.Status = "error"
		fmt.Println("Error creating Terraform executor:", err)
		return
	}
	tf.SetStdout(logFile)
	tf.SetStderr(logFile)

	ctx := context.Background()

	// init
	if err := tf.Init(ctx, tfexec.Upgrade(true)); err != nil {
		job.Status = "error"
		return
	}

	job.Status = "running(apply)"
	jobs.Store(jobID, job)

	// apply
	if err := tf.Apply(ctx,
		tfexec.VarFile("runtime.tfvars"),
	); err != nil {
		job.Status = "error"
		fmt.Println("Error applying Terraform configuration:", err)
		return
	}

	job.IP = getVMIP(job)
	job.Status = "done"
	jobs.Store(jobID, job)
}

func getVMIP(job *Job) string {
	tf, err := tfexec.NewTerraform(job.Workdir, "terraform")
	if err != nil {
		job.Status = "error"
		return ""
	}

	ctx := context.Background()

	// terraform output -json と同じ
	out, err := tf.Output(ctx)
	if err != nil {
		job.Status = "error"
		return ""
	}

	// vm_ip という output 名を直接取得
	v, ok := out["vm_ip"]
	if !ok {
		return ""
	}

	// Value は interface{} なので JSON 経由で安全に []string に
	b, _ := json.Marshal(v.Value)

	var ips []string
	if err := json.Unmarshal(b, &ips); err != nil {
		return ""
	}

	if len(ips) > 0 {
		return ips[0]
	}
	return ""
}

func createVMHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	jobID := fmt.Sprintf("%d", time.Now().UnixNano())

	req := VMRequest{
		CPU:        atoiSafe(r.FormValue("cpu")),
		Memory:     atoiSafe(r.FormValue("memory")),
		HDD:        atoiSafe(r.FormValue("hdd")),
		Servername: r.FormValue("servername"),
		Username:   r.FormValue("username"),
		Password:   r.FormValue("password"),
	}

	jobs.Store(jobID, &Job{Status: "running", Servername: req.Servername})

	go runTerraformJob(jobID, &req)

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"job_id": jobID})
		return
	}
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

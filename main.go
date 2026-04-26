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
	"crypto/rand"
	"encoding/base64"

	"github.com/amoghe/go-crypt"
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
	// provider.tf をコピー
	copyFile("terraform/provider.tf", filepath.Join(workdir, "provider.tf"))
	copyFile("terraform/variables.tf", filepath.Join(workdir, "variables.tf"))
	copyFile("terraform/terraform.tfvars", filepath.Join(workdir, "proxmox.auto.tfvars"))

	// Terraform実行
	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = workdir
	initOut, _ := initCmd.CombinedOutput()

	applyCmd := exec.Command("terraform", "apply", "-auto-approve", "-var-file=runtime.tfvars")
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

func hashPasswordForLinux(password string) (string, error) {
	// ランダムsalt生成（16byte）
	saltBytes := make([]byte, 16)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return "", err
	}

	salt := base64.RawStdEncoding.EncodeToString(saltBytes)

	// $6$ = SHA-512 crypt
	hash, err := crypt.Crypt(password, "$6$"+salt)
	if err != nil {
		return "", err
	}

	return hash, nil
}
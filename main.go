package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"os/exec"
	"io"

	"github.com/amoghe/go-crypt"
	"github.com/joho/godotenv"
)

var PORT string

func main() {
	godotenv.Load()
	PORT = os.Getenv("PORT")
	loadConfig()

	fs := http.FileServer(http.Dir("./static"))
	http.HandleFunc("/", requireLogin(func(w http.ResponseWriter, r *http.Request) {

		// ルートは resource.html を表示
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./static/resource.html")
			return
		}

		// それ以外は静的ファイルとして配信（一覧は出ない）
		fs.ServeHTTP(w, r)
	}))
	http.HandleFunc("/create-vm", requireLogin(createVMHandler))
	http.HandleFunc("/status", requireLogin(statusHandler))

	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(":" + PORT, nil))
}

func copyFile(src, dst string) {
	input, err := os.ReadFile(src)
	if err != nil {
		fmt.Println("Error reading source file:", err)
		return
	}
	os.WriteFile(dst, input, 0644)
}

func runCmdWithLog(cmd *exec.Cmd, logFile *os.File) ([]byte, error) {
	var buf bytes.Buffer

	cmd.Stdout = io.MultiWriter(logFile, &buf)
	cmd.Stderr = io.MultiWriter(logFile, &buf)

	err := cmd.Run()
	if err != nil {
		log.Printf("Command failed: %v", err)
	}

	return buf.Bytes(), err
}

func hashPasswordForLinux(password string) (string, error) {
	// ランダムsalt生成（16byte）
	saltBytes := make([]byte, 16)
	_, err := rand.Read(saltBytes)
	if err != nil {
		fmt.Println("Error generating salt:", err)
		return "", err
	}

	salt := base64.RawStdEncoding.EncodeToString(saltBytes)

	// $6$ = SHA-512 crypt
	hash, err := crypt.Crypt(password, "$6$"+salt)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return "", err
	}

	return hash, nil
}
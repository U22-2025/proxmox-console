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
	http.HandleFunc("/status", statusHandler)

	fmt.Println("Server started at http://172.32.0.70:" + PORT)
	log.Fatal(http.ListenAndServe(":" + PORT, nil))
}

func copyFile(src, dst string) {
	input, err := os.ReadFile(src)
	if err != nil {
		return
	}
	os.WriteFile(dst, input, 0644)
}

func runCmdWithLog(cmd *exec.Cmd, logFile *os.File) ([]byte, error) {
	var buf bytes.Buffer

	cmd.Stdout = io.MultiWriter(logFile, &buf)
	cmd.Stderr = io.MultiWriter(logFile, &buf)

	err := cmd.Run()

	return buf.Bytes(), err
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
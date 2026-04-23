package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// 静的ファイル配信
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// VM作成リクエスト受付
	http.HandleFunc("/create-vm", createVMHandler)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createVMHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// フォーム値取得
	cpu := r.FormValue("cpu")
	memory := r.FormValue("memory")
	hdd := r.FormValue("hdd")
	servername := r.FormValue("servername")
	username := r.FormValue("username")
	password := r.FormValue("password")

	// YAML生成
	yaml := fmt.Sprintf(`vm:
  name: %s
  resources:
    cpu: %s
    memory: %s
    hdd: %s
  user:
    name: %s
    password: %s
`, servername, cpu, memory, hdd, username, password)

	// ファイル名（重複しないようタイムスタンプ）
	filename := fmt.Sprintf("vm_%s_%d.yaml", servername, time.Now().Unix())

	err := os.WriteFile(filename, []byte(yaml), 0644)
	if err != nil {
		http.Error(w, "YAMLファイル生成失敗", http.StatusInternalServerError)
		return
	}

	// 完了画面
	fmt.Fprintf(w, `
		<html>
		<head><meta charset="UTF-8"><title>完了</title></head>
		<body>
			<h2>仮想マシン申請を受け付けました</h2>
			<p>YAMLファイルを生成しました: %s</p>
			<a href="/">最初のページへ戻る</a>
		</body>
		</html>
	`, filename)
}
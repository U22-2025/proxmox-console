package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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

		// ルートは dashboard.html を表示
		if r.URL.Path == "/" {
			w.Header().Set("Cache-Control", "no-store")
			http.ServeFile(w, r, "./static/dashboard.html")
			return
		}

		// それ以外は静的ファイルとして配信（一覧は出ない）
		fs.ServeHTTP(w, r)
	}))
	http.HandleFunc("/create-vm", requireLogin(createVMHandler))
	http.HandleFunc("/status", requireLogin(statusHandler))
	http.HandleFunc("/api/jobs", requireLogin(listJobsHandler))
	http.HandleFunc("/logout", logoutHandler)

	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: logout_token を取得
	req1, err := http.NewRequest("GET", AppConfig.Kratos.PublicURL+"/self-service/logout/browser", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, c := range r.Cookies() {
		req1.AddCookie(c)
	}
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp1.Body.Close()

	var kratosResp struct {
		LogoutToken string `json:"logout_token"`
	}
	if err := json.NewDecoder(resp1.Body).Decode(&kratosResp); err != nil || kratosResp.LogoutToken == "" {
		http.Error(w, "failed to get logout token", http.StatusInternalServerError)
		return
	}

	// Step 2: Kratosのセッションをサーバー側で削除
	logoutURL := AppConfig.Kratos.PublicURL + "/self-service/logout?token=" + kratosResp.LogoutToken
	req2, _ := http.NewRequest("GET", logoutURL, nil)
	for _, c := range r.Cookies() {
		req2.AddCookie(c)
	}
	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if resp2, err := noRedirectClient.Do(req2); err == nil {
		for _, cookie := range resp2.Cookies() {
			http.SetCookie(w, cookie)
		}
		resp2.Body.Close()
	}

	// Step 3: セッションCookieをブラウザから確実に削除
	// サーバー側削除のタイミング差を防ぐため Cookie を直接 expire する。
	// Cookie がなければ requireLogin → /sessions/whoami が必ず 401 を返す。
	http.SetCookie(w, &http.Cookie{
		Name:   "ory_kratos_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func listJobsHandler(w http.ResponseWriter, r *http.Request) {
	type jobResp struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		IP          string `json:"ip"`
		Servername  string `json:"servername"`
	}

	var result []jobResp
	jobs.Range(func(key, value interface{}) bool {
		j := value.(*Job)
		result = append(result, jobResp{
			ID:         key.(string),
			Status:     j.Status,
			IP:         j.IP,
			Servername: j.Servername,
		})
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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

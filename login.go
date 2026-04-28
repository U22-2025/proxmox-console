package main
import(
	"net/http"
)

func requireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		req, _ := http.NewRequest("GET", "http://kratos:4433/sessions/whoami", nil)

		// ブラウザのCookieをそのままKratosへ渡す
		for _, c := range r.Cookies() {
			req.AddCookie(c)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			http.Redirect(w, r, "http://172.32.0.70:3000/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
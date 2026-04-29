package main
import(
	"net/http"
)

func requireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		url := AppConfig.Kratos.PublicURL + "/sessions/whoami"
		req, _ := http.NewRequest("GET", url, nil)

		for _, c := range r.Cookies() {
			req.AddCookie(c)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// Kratos 到達不可はログイン扱いにしない
			http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// 200 OK はログイン済み
		if resp.StatusCode == http.StatusOK {
			next(w, r)
			return
		}

		// 401 Unauthorized は未ログイン
		if resp.StatusCode == http.StatusUnauthorized {
			// 元URLを return_to に入れる（超重要）
			returnTo := url.QueryEscape("http://" + r.Host + r.URL.RequestURI())

			loginURL := AppConfig.Kratos.PublicURL +
				"/self-service/login/browser?return_to=" + returnTo

			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}

		http.Error(w, "unexpected auth response", http.StatusInternalServerError)
	}
}
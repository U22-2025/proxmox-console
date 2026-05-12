package main
import(
	"net/http"
	"net/url"
	"strings"
)

func requireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		whoamiURL := AppConfig.Kratos.PublicURL + "/sessions/whoami"
		req, _ := http.NewRequest("GET", whoamiURL, nil)

		for _, c := range r.Cookies() {
			req.AddCookie(c)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			next(w, r)
			return
		}

		if resp.StatusCode == http.StatusUnauthorized {
			// /api/ パスは fetch() から呼ばれるため 302 ではなく 401 を返す。
			// クライアント側で 401 を検知して window.location.href で遷移させる。
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			returnTo := url.QueryEscape("http://" + r.Host + r.URL.RequestURI())
			loginURL := AppConfig.Kratos.PublicURL +
				"/self-service/login/browser?return_to=" + returnTo
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}

		http.Error(w, "unexpected auth response", http.StatusInternalServerError)
	}
}
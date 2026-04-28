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
		if err != nil || resp.StatusCode != 200 {
			http.Redirect(w, r, AppConfig.Kratos.LoginURL, http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
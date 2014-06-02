package ngAuthApi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mrvdot/appengine/accounts"
	"github.com/mrvdot/golang-utils"
)

var (
	sessionCookie = "ngAuthUserSession"
)

func init() {
	r := mux.NewRouter()

	r.HandleFunc("/load", loadUser)
	r.HandleFunc("/register", registerUser)

	http.Handle("/", utils.CorsHandler(accounts.AuthenticatedHandler(r)))

	initAccounts()
}

func initAccounts() {
	accounts.Headers = map[string]string{
		"account": "X-ngauth-account",
		"key":     "X-ngauth-key",
		"session": "X-ngauth-session",
	}
	accounts.InitRouter("_auth")
}

func loadUser(rw http.ResponseWriter, req *http.Request) {
	ctx, err := accounts.GetContext(req)
	if err != nil {
		ctx.Errorf("Error getting context for account: %v", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	session, err := req.Cookie(sessionCookie)
	if err != nil {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte(err.Error()))
		return
	}
	user := getBySession(ctx, session.Value)
	if user == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	sendResult(rw, user)
}
func registerUser(rw http.ResponseWriter, req *http.Request) {
	ctx, err := accounts.GetContext(req)
	if err != nil {
		ctx.Errorf("Error getting context for account: %v", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	user := newUser(ctx, req)
	var domain string
	if reqUrl, err := url.Parse(req.Header.Get("Origin")); err != nil {
		domain = reqUrl.Host
		// If domain includes port, slice it off
		if strings.Contains(domain, ":") {
			domainParts := strings.Split(domain, ":")
			domain = domainParts[0]
		}
	}
	cookie := &http.Cookie{
		Name:   sessionCookie,
		Value:  user.Session,
		Domain: domain,
	}
	http.SetCookie(rw, cookie)
	sendResult(rw, user)
}

func sendResult(rw http.ResponseWriter, result interface{}) {
	resp := &utils.ApiResponse{
		Result: result,
	}
	enc := json.NewEncoder(rw)
	enc.Encode(resp)
}

package logout

import (
    "net/http"
    "encoding/json"
    
    "github.com/gorilla/mux"
    logrus "github.com/sirupsen/logrus"
    "TwitterClone/wrappers"
)


type response struct {
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}
var Log *logrus.Logger
func main() {
    Log.SetLevel(logrus.ErrorLevel)
}

func isLoggedIn(r *http.Request) bool {
    cookie, _ := r.Cookie("username")
    return cookie != nil
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    cookie, res := logoutEndpoint(r)
    if (cookie != nil) {
        http.SetCookie(w, cookie)
    }
    encodeResponse(w, res)
}

func logoutEndpoint(r *http.Request) (*http.Cookie, response) {
    var res response
    var cookie *http.Cookie
    if isLoggedIn(r) {
        cookie, _ = r.Cookie("username")
        cookie.MaxAge = -1 // Delete cookie.
        res.Status = "OK"
    } else {
        res.Status = "error"
        res.Error = "User not logged in."
    }
    return cookie, res
}

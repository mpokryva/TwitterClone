package main

import (
    "context"
    "log"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/mongo"
    "github.com/mongodb/mongo-go-driver/bson"
)


type response struct {
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}

type userDetails struct {
    Username *string `json:"username"`
    Password *string `json:"password"`
}

func main() {
    r := mux.NewRouter()
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    r.HandleFunc("/login", loginHandler).Methods("POST")
    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":8003", nil))
}

func authUser(details userDetails) bool {
    client, err := mongo.NewClient("mongodb://localhost:27017")
    if err != nil {
        log.Println("Mongodb error")
        return false
    }
    db := client.Database("twitter")
    col := db.Collection("users")
    doc := bson.NewDocument(bson.EC.String("username", *details.Username),
            bson.EC.String("password", *details.Password),
            bson.EC.Boolean("verified", true))
    cursor, err := col.Find(
        context.Background(),
        doc)
    return cursor.Next(context.Background())
}

func decodeRequest(r *http.Request) (userDetails, error) {
    decoder := json.NewDecoder(r.Body)
    var details userDetails
    err := decoder.Decode(&details)
    return details, err
}
func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    var res response
    details, err := decodeRequest(r)
    if (err != nil) {
        res.Status = "error"
        res.Error = "Error decoding json."
    } else if validateDetails(details) {
        res = loginEndpoint(details)
    } else {
        res.Status = "error"
        res.Error = "Invalid request."
    }
    if (res.Status == "OK") {
        var cookie http.Cookie
        cookie.Name = "username"
        cookie.Value = *details.Username
        http.SetCookie(w, &cookie)
    }
    encodeResponse(w, res)
}

func loginEndpoint(details userDetails) (response) {
    var res response
    // Check username and password against database.
    shouldLogin := authUser(details)
    if (shouldLogin) {
        res.Status = "OK"
    } else {
        res.Status = "error"
        res.Error = "User not found or incorrect password."
    }
    return res
}

func validateDetails(details userDetails) bool {
    return (details.Username != nil && details.Password != nil)
}

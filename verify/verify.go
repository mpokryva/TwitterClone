package main

import (
    "context"
    "github.com/onrik/logrus/filename"
    log "github.com/sirupsen/logrus"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson"
    "TwitterClone/wrappers"
)

type verification struct {
  Key *string `json:"key"`
  Email *string `json:"email"`
}

type res struct {
  Status string `json:"status"`
  Error string `json:"error,omitempty"`
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/verify", verifyUser).Methods("POST")
    http.Handle("/", r)
    log.AddHook(filename.NewHook())
    log.SetLevel(log.ErrorLevel)
    log.Fatal(http.ListenAndServe(":8004", nil))
}


func verifyUser(w http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var verif verification
    var r res
    err := decoder.Decode(&verif)
    if err != nil {
        panic(err)
    }
    valid := validateParams(verif)
    if valid {
      if(user_exists(verif)){
          r.Status = "OK"
          // log.Println("Line 46")
      }else {
        log.Error("Input not valid!")
        r.Status = "error"
        r.Error = "Could not complete verification"
      }
  }else{
    r.Status = "error"
    r.Error = "Not enough input"
  }
  json.NewEncoder(w).Encode(r)
}

func validateParams(verif verification) bool {
    valid := true
    if (verif.Email == nil) {
        valid = false
    } else if (verif.Key == nil) {
        valid = false
    }
    // if (valid) {
    //     log.Println("Key: ", *verif.Key)
    //     log.Println("Email: ", *verif.Email)
    // }
    return valid
}

func user_exists(verif verification) bool {
    client, err := wrappers.NewClient()
    if err != nil {
        log.Error("Mongodb error")
        return false
    }
    db := client.Database("twitter")
    col := db.Collection("users")
    filter := bson.NewDocument(bson.EC.String("email", *verif.Email),
            bson.EC.String("key", "<" + *verif.Key + ">"),
            bson.EC.Boolean("verified", false))
    update := bson.NewDocument(
        bson.EC.SubDocumentFromElements("$set",
        bson.EC.Boolean("verified", true)))
    result, err := col.UpdateOne(
        context.Background(),
        filter, update)
    //log.Println("Here:" + result.ModifiedCount == int64(1))
    // log.Println("Here")
    return result.ModifiedCount == int64(1)
}

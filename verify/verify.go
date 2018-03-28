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
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    r.HandleFunc("/verify", verifyUser).Methods("POST")
    http.Handle("/", r)
    http.ListenAndServe(":8004", nil)
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
      log.Println(verif)
      if(user_exists(verif)){
          r.Status = "OK"
          log.Println("Line 46")
      }else {
        log.Println("Not valid!")
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
    if (valid) {
        log.Println("Key: ", *verif.Key)
        log.Println("Email: ", *verif.Email)
    }
    return valid
}

func user_exists(verif verification) bool {
    client, err := mongo.NewClient("mongodb://localhost:27017")
    if err != nil {
        log.Println("Mongodb error")
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
    log.Println("Here")
    return result.ModifiedCount == int64(1)
}

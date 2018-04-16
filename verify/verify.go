package verify

import (
    "context"
    "os"
    logrus "github.com/sirupsen/logrus"
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
var log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/verify", VerifyHandler).Methods("POST")
    http.Handle("/", r)
    // Log to a file
    var f *os.File
    var err error
    log, f, err = wrappers.FileLogger("verify.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        Log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    Log.SetLevel(logrus.ErrorLevel)
    Log.Fatal(http.ListenAndServe(":8004", nil))
}


func VerifyHandler(w http.ResponseWriter, req *http.Request) {
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
      }else {
        Log.Error("Input not valid!")
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
        Log.Debug("Key: ", *verif.Key)
        Log.Debug("Email: ", *verif.Email)
    }
    return valid
}

func user_exists(verif verification) bool {
    client, err := wrappers.NewClient()
    if err != nil {
        Log.Error("Mongodb error")
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
    result, err := col.UpdateMany(
        context.Background(),
        filter, update)
    if err != nil {
        Log.Error(err)
    }
    return result.ModifiedCount == 1
}

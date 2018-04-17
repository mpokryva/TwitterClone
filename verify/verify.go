package verify

import (
    "context"
    "time"
    logrus "github.com/sirupsen/logrus"
    "net/http"
    "encoding/json"
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
var Log *logrus.Logger
func main() {
    Log.SetLevel(logrus.ErrorLevel)
}


func VerifyHandler(w http.ResponseWriter, req *http.Request) {
  start := time.Now()
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

  elapsed := time.Since(start)
  Log.Info("Verify elapsed: " + elapsed.String())
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
    dbStart := time.Now()
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
      elapsed := time.Since(dbStart)
      Log.WithFields(logrus.Fields{"msg":"Check if user exists time elapsed", "timeElapsed":elapsed.String()}).Error(err)
    }
      elapsed := time.Since(dbStart)
      Log.WithFields(logrus.Fields{"msg":"Check if user exists time elapsed", "timeElapsed":elapsed.String()}).Info()
    return result.ModifiedCount == 1
}

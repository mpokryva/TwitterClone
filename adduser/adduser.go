package main

import (
    "context"
    "log"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/mongo"
    "github.com/mongodb/mongo-go-driver/bson"
    "crypto/md5"
    "encoding/hex"
    "net/smtp"
    "math/rand"
    "strconv"
)

type user struct {
    Username *string `json: "username"`
    Password *string `json: "password"`
    Email *string `json: "email"`
}

type res struct {
  Status string `json: "status"`
  Error string `json: "error"`
}

func main() {
    r := mux.NewRouter()
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    r.HandleFunc("/adduser", addUser).Methods("POST")
    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}

func insertUser(us *user, key string) bool{
    client, err := mongo.NewClient("mongodb://localhost:27017")
    if err != nil {
        log.Println("Panicking")
        panic(err)
    }
    db := client.Database("twitter")
    col := db.Collection("users")
    log.Println(*us)
    doc := bson.NewDocument(bson.EC.String("username", *(us.Username)))
    doc.Append(bson.EC.String("email", *(us.Email)))
    // bytePassword := []byte(*(us.Password))
    // hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
    // if err != nil{
    //   panic(err)
    // }
    doc.Append(bson.EC.String("password", *(us.Password)))
    doc.Append(bson.EC.String("verify", "0"))
    doc.Append(bson.EC.String("key", "<"+key+">"))
    _,err2 := col.InsertOne(context.Background(),doc)
    if err2 != nil {
        return false
    } else {
        return true
    }
}

func addUser(w http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var us user
    var r res
    err := decoder.Decode(&us)
    if err != nil {
        panic(err)
    }
    log.Println(us)
    valid := validateUser(us)
    if valid {
      //create the hashed verification key
      num := rand.Intn(1000)
      numstring := strconv.Itoa(num)
      hasher := md5.New()
      hasher.Write([]byte(*(us.Username)))
      hasher.Write([]byte(numstring))
      key := hex.EncodeToString(hasher.Sum(nil))
      // Add the user.
      log.Println(us)
      if(email(us, key)){
        insert := insertUser(&us, key)
        if(insert){
          r.Status = "OK"
          r.Error = ""
        }else {
          log.Println("Not valid!")
          r.Status = "error"
          r.Error = "Invalid/not enough input."
        }
      }else{
        log.Println("Couldn't email")
        r.Status = "error"
        r.Error = "Email could not be sent"
      }
    return json.NewEncoder(w).Encode(r)
  }
}
func email(us user, key string) bool{
  msg := "From: twiti.verify@gmail.com \n To: " + *(us.Username) + "\n" +
    "Subject: Account Verification\n\n"+
    "Thank you for joining Twiti!\n This is your validation key: <" + key + "> \n Please click the link to quickly veify your account."

  err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("","twiti.verify@gmail.com","cloud356", "smtp.gmail.com"),"twiti.verify@gmail.com",[]string{*(us.Email)}, []byte(msg) )
  if err != nil {
		log.Printf("smtp error: %s", err)
		return false
	}
  return true
}
func validateUser(us user) bool {
    valid := true
    if (us.Username == nil) {
        valid = false
    } else if (us.Password == nil) {
        valid = false
    }else if (us.Email == nil) {
        valid = false
    }
    return valid
}

package main

import (
    "context"
    "log"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/mongodb/mongo-go-driver/bson"
    "crypto/md5"
    "encoding/hex"
    "net/smtp"
    "math/rand"
    "strconv"
    "golang.org/x/crypto/bcrypt"
    "TwitterClone/user"
    "TwitterClone/wrappers"
)

type request struct {
    Username *string `json:"username"`
    Password *string `json:"password"`
    Email *string `json:"email"`
}

type res struct {
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}


func main() {
    r := mux.NewRouter()
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    r.HandleFunc("/adduser", addUser).Methods("POST")
    http.Handle("/", r)
    http.ListenAndServe(":8002", nil)
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func insertUser(user user.User, key string) bool {
    client, err := wrappers.NewClient()
    if err != nil {
        log.Println("Panicking")
        panic(err)
    }
    db := client.Database("twitter")
    col := db.Collection("users")
    existingDoc :=bson.NewDocument(bson.EC.String("email", user.Email))
    err1, errorName := col.Count(context.Background(), existingDoc);
    log.Println(existingDoc)
    if err1 > 0{
        log.Println(errorName)
        return false
    }
    doc := bson.NewDocument(bson.EC.String("username", user.Username))
    err4, errorEmail := col.Count(context.Background(),doc);
    if err4 > 0{
        log.Println("username error: %s",errorEmail)
        return false
    }
    bytePassword := []byte(user.Password)
    hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
    if err != nil {
        return false
    }
    user.Password = (string)(hashedPassword)
    user.Key = "<"+key+">"
    user.Verified = false
    log.Println(user)
    _, err = col.InsertOne(context.Background(), &user)
    return err == nil
}

func addUser(w http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var us request
    var r res
    err := decoder.Decode(&us)
    if err != nil {
        panic(err)
    }
    log.Println(us)
    valid := validateUser(us)
    if valid {
        var user user.User
        user.Email = *us.Email
        user.Username = *us.Username
        user.Password = *us.Password
        log.Println(user)
        //create the hashed verification key
        num := rand.Int()
        numstring := strconv.Itoa(num)
        log.Println(num, numstring)
        hasher := md5.New()
        hasher.Write([]byte(user.Username))
        hasher.Write([]byte(numstring))
        key := hex.EncodeToString(hasher.Sum(nil))
        // Add the user.
        insert := insertUser(user, key)
        em := email(user, key)
        log.Println(insert, em)
        if(insert == true && em == true){
            r.Status = "OK"
            //resM,_ := json.Marshal(r)
            //r.Status = "OK"
        }else {
            log.Println("Not valid!")
            r.Status = "error"
            r.Error = "Username/email is already in use"
        }
    }else{
        r.Status = "error"
        r.Error = "Not enough input"
    }
    encodeResponse(w, r)
}

func email(us user.User, key string) bool{
    link := "http://nsamba.cse356.compas.cs.stonybrook.edu/verify?email="+us.Email+"&key="+key
    msg := []byte("To: "+us.Email+"\r\n" +
    "Subject: Validation Email\r\n" +
    "\r\n" +
    "Thank you for joining Twiti!\n This is your validation key: <" + key + "> \n Please click the link to quickly veify your account: "+ link+"\r\n")

    err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("","twiti.verify@gmail.com","cloud356", "smtp.gmail.com"),"twiti.verify@gmail.com",[]string{us.Email}, []byte(msg) )
    if err != nil {
        log.Printf("smtp error: %s", err)
        return false
    }
    return true
}

func validateUser(us request) bool {
    valid := true
    if (us.Username == nil) {
        valid = false
    } else if (us.Password == nil) {
        valid = false
    }else if (us.Email == nil) {
        valid = false
    }
    log.Println(valid)
    return valid
}

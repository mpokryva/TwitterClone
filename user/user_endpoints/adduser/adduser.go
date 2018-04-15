package adduser

import (
    "context"
    "os"
    "errors"
    "github.com/sirupsen/logrus"
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

type response struct {
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}
var log *logrus.Logger
func main() {
    r := mux.NewRouter()
    r.HandleFunc("/adduser", AddUserHandler).Methods("POST")
    http.Handle("/", r)
    // Log to a file
    var f *os.File
    var err error
    log, f, err = wrappers.FileLogger("adduser.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    log.SetLevel(logrus.ErrorLevel)
    http.ListenAndServe(":8002", nil)
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func insertUser(user user.User, key string) error {
    client, err := wrappers.NewClient()
    if err != nil {
        log.Error(err)
        return err
    }
    db := client.Database("twitter")
    col := db.Collection("users")
    filter := bson.NewDocument(bson.EC.String("email", user.Email))
    count, err := col.Count(context.Background(), filter);
    if count > 0 {
        err = errors.New("The email " + user.Email + " is already in use.")
        log.Error(err)
        return err
    } else if err != nil {
        log.Error(err)
        return err
    }
    filter = bson.NewDocument(bson.EC.String("username", user.Username))
    count, err = col.Count(context.Background(), filter);
    if count > 0 {
        err = errors.New("The username " + user.Username + " is already in use.")
        log.Error(err)
        return err
    } else if err != nil {
        log.Error(err)
        return err
    }
    bytePassword := []byte(user.Password)
    hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
    if err != nil {
        log.Error(err)
        return err
    }
    user.Password = (string)(hashedPassword)
    user.Key = "<"+key+">"
    user.Verified = false
    _, err = col.InsertOne(context.Background(), &user)
    if err != nil {
        log.Error(err)
    }
    return err
}

func sendError(w http.ResponseWriter, err error) {
    var res response
    res.Status = "error"
    res.Error = err.Error()
    encodeResponse(w, res)
}

func AddUserHandler(w http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var us request
    err := decoder.Decode(&us)
    if err != nil {
        sendError(w, err)
        return
    }
    err = validateUser(us)
    if err != nil {
        sendError(w, err)
        return
    }
    var user user.User
    user.Email = *us.Email
    user.Username = *us.Username
    user.Password = *us.Password
    log.Debug(user)
    // Create the hashed verification key.
    num := rand.Int()
    numstring := strconv.Itoa(num)
    log.Println(num, numstring)
    hasher := md5.New()
    hasher.Write([]byte(user.Username))
    hasher.Write([]byte(numstring))
    key := hex.EncodeToString(hasher.Sum(nil))
    // Add the user.
    err = insertUser(user, key)
    if err != nil {
        sendError(w, err)
        return
    }
    // Email user once inserted into db.
    err = email(user, key)
    if err != nil {
        sendError(w, err)
        return
    }
    var res response
    res.Status = "OK"
    encodeResponse(w, res)
}

func email(us user.User, key string) error {
    link := "http://nsamba.cse356.compas.cs.stonybrook.edu/verify?email="+us.Email+"&key="+key
    msg := []byte("To: "+us.Email+"\r\n" +
    "Subject: Validation Email\r\n" +
    "\r\n" +
    "Thank you for joining Twiti!\n This is your validation key: <" + key + "> \n Please click the link to quickly veify your account: "+ link+"\r\n")
    addr := "192.168.1.24:25"
    err := smtp.SendMail(addr, nil,
    "<mongo-config>",
       []string{us.Email}, msg)
    if err != nil {
        log.Error(err)
    }
    return err
}

func validateUser(us request) error {
    var err error
    if (us.Username == nil) {
        err = errors.New("No username in adduser request.")
    } else if (us.Password == nil) {
        err = errors.New("No password in adduser request.")
    }else if (us.Email == nil) {
        err = errors.New("No email in adduser request.")
    }
    if err != nil {
        log.Error(err)
    }
    return err
}

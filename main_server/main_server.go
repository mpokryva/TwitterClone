package main

import (
    "log"
    "net/http"
    "github.com/sirupsen/logrus"
    "TwitterClone/wrappers"
    "TwitterClone/additem"
    /*
    "TwitterClone/user/user_endpoints"
    "TwitterClone/user/user_endpoints/followInfo"
    "TwitterClone/user/user_endpoints/adduser"
    "TwitterClone/search"
    "TwitterClone/media/addmedia"
    "TwitterClone/media/media_endpoints"
    "TwitterClone/follow"
    "TwitterClone/item/item_endpoints"
    "TwitterClone/verify"
    "TwitterClone/login"
    "TwitterClone/logout"
    */
)

var Log *logrus.Logger

func main() {
    // Log to a file
    var f *os.File
    var err error
    Log, f, err = wrappers.FileLogger("central.log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    log.SetLevel(logrus.ErrorLevel)
    additem.Log = Log
    router := NewRouter()
    http.Handle("/", router)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

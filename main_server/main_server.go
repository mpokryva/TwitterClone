package main

import (
    "log"
    "net/http"
    "github.com/sirupsen/Logrus"
    "TwitterClone/wrappers"
    "TwitterClone/additem"
    "TwitterClone/user/user_endpoints"
    "TwitterClone/user/user_endpoints/followInfo"
    "TwitterClone/user/user_endpoints/adduser"
    "TwitterClone/search"
    "TwitterClone/media/addmedia"
    "TwitterClone/media/media_endpoints"
    "TwitterClone/follow"
    "TwitterClone/item/item_endpoints"
    "TwitterClone/verify"
    "TwitterClone/Login"
    "TwitterClone/Logout"
)

var Log *Logrus.Logger

func main() {
    // Log to a file
    var f *os.File
    var err error
    Log, f, err = wrappers.FileLogger("central.Log", os.O_CREATE | os.O_RDWR,
        0666)
    if err != nil {
        Log.Fatal("Logging file could not be opened.")
    }
    f.Truncate(0)
    f.Seek(0, 0)
    defer f.Close()
    Log.SetLevel(Logrus.ErrorLevel)
    injectLogger()
    router := NewRouter()
    http.Handle("/", router)
    Log.Fatal(http.ListenAndServe(":8080", nil))
}

func injectLogger() {
    additem.Log = Log
    user_endpoints.Log = Log
    followInfo.Log = Log
    addUser.Log = Log
    search.Log = Log
    addmedia.Log = Log
    media_endpoints.Log = Log
    follow.Log = Log
    item_endpoints.Log = Log
    verify.Log = Log
    Login.Log = Log
    Logout.Log = Log
}

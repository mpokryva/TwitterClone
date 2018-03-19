package main

import (
    //"io"
    "log"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/satori/go.uuid"
)

type item struct {
    Content string
    ChildType *string
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/additem", addItem).Methods("POST")
    http.Handle("/", r)
    http.ListenAndServe(":8080", nil)
}


func addItem(w http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var it item
    err := decoder.Decode(&it)
    if err != nil {
        panic(err)
    }
    log.Println(it)
    valid := validateItem(it)
    if valid {
        // Add the item.
        id := uuid.Must(uuid.NewV4())
        log.Println(id)

    } else {

    }
    // Validate req
    /*for {
        var it item
        if err := decoder.Decode(&it); err == io.EOF {
            break
        } else if err != nil {
            panic(err)
        }
    }*/
}

func validateItem(it item) bool {
    log.Println("Hey")
    valid := true
    if (it.ChildType == nil) {
        valid = true
    } else if (*it.ChildType != "retweet" && *it.ChildType != "reply") {
        // Invalid req
        log.Println("childType not valid")
        valid = false
    }
    return valid
}

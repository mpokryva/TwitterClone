package memcached

import (
    "net/http"
    "encoding/json"
    "github.com/bradfitz/gomemcache/memcache"
    "github.com/gorilla/mux"
    "github.com/sirupsen/logrus"
)

type response struct {
    Value []byte `json:"value,omitempty"`
    Error string `json:"error,omitempty"`
}

type setRequest struct {
    Key string
    Value string
}

var Log *logrus.Logger

func GetSingleHandler(w http.ResponseWriter, r *http.Request) {
    Log.SetLevel(logrus.DebugLevel)
    mc := memcache.New("memcached-1:11211")
    vars := mux.Vars(r)
    key := vars["key"]
    Log.Debug(key)
    item, err := mc.Get(key)
    if err != nil {
        Log.Error(err)
        sendError(w, err)
    } else {
        Log.Debug(item.Value)
        var res response
        res.Value = item.Value
        encodeResponse(w, res)
    }
}

func SetHandler(w http.ResponseWriter, r *http.Request) {
    mc := memcache.New("memcached-1:11211")
    item, err := decodeRequest(r)
    if err != nil {
        Log.Error(err)
        sendError(w, err)
        return
    }
    err = mc.Set(&item)
    if err != nil {
        Log.Error(err)
        sendError(w, err)
    }
}

func decodeRequest(r *http.Request) (memcache.Item, error) {
    decoder := json.NewDecoder(r.Body)
    var setReq setRequest
    var item memcache.Item
    err := decoder.Decode(&setReq)
    if err != nil {
        return item, err
    }
    item.Key = setReq.Key
    item.Value = []byte(setReq.Value)
    return item, nil
}

func sendError(w http.ResponseWriter, err error) {
    var res response
    res.Error = err.Error()
    w.WriteHeader(http.StatusInternalServerError)
    encodeResponse(w, res)
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

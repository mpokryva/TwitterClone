package memcached

import (
    "net/http"
    "encoding/json"
    "github.com/bradfitz/gomemcache/memcache"
    "github.com/gorilla/mux"
    "github.com/sirupsen/logrus"
)

type GetResponse struct {
    Value []byte `json:"value,omitempty"`
    Error string `json:"error,omitempty"`
}

type SetResponse struct {
    Error string `json:"error,omitempty"`
}


type SetRequest struct {
    Key string
    Value []byte
}

var Log *logrus.Logger
var mc *memcache.Client
func MemcachedClient() *memcache.Client {
    if mc == nil {
        mc = memcache.New("memcached-1:11211")
    }
    return mc
}

func init() {
    mc = memcache.New("memcached-1:11211")
}

func GetSingleHandler(w http.ResponseWriter, r *http.Request) {
    Log.SetLevel(logrus.InfoLevel)
    vars := mux.Vars(r)
    key := vars["key"]
    Log.Debug(key)
    item, err := mc.Get(key)
    if err != nil {
        Log.Error(err)
        sendGetError(w, err, http.StatusNotFound)
    } else {
        Log.Debug(item.Value)
        Log.Debug(string(item.Value))
        var res GetResponse
        res.Value = item.Value
        encodeResponse(w, res)
    }
}

func SetHandler(w http.ResponseWriter, r *http.Request) {
    item, err := decodeRequest(r)
    if err != nil {
        Log.Error(err)
        sendSetError(w, err, http.StatusInternalServerError)
        return
    }
    err = mc.Set(&item)
    if err != nil {
        Log.Error(err)
        // Not necessarily a cache miss. Could be an actual error
        // but choosing not to spend time checking.
        sendSetError(w, err, http.StatusNotFound)
    }
}

func decodeRequest(r *http.Request) (memcache.Item, error) {
    decoder := json.NewDecoder(r.Body)
    var setReq SetRequest
    var item memcache.Item
    err := decoder.Decode(&setReq)
    if err != nil {
        return item, err
    }
    item.Key = setReq.Key
    item.Value = setReq.Value
    return item, nil
}

func sendSetError(w http.ResponseWriter, err error, statusCode int) {
    var res SetResponse
    res.Error = err.Error()
    sendError(w, res, statusCode)
}

func sendError(w http.ResponseWriter, response interface{}, statusCode int) {
    w.WriteHeader(statusCode)
    encodeResponse(w, response)
}

func sendGetError(w http.ResponseWriter, err error, statusCode int) {
    var res GetResponse
    res.Error = err.Error()
    sendError(w, res, statusCode)
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

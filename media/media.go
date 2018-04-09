package media

import(
    "github.com/mongodb/mongo-go-driver/bson/objectid"
)

type Media struct {
    ID objectid.ObjectID `bson:"_id"`
    Content []byte `bson:"content"`
    Username string `bson:"username"`
    ItemIDs []objectid.ObjectID `bson:"item_ids,omitempty"`
}

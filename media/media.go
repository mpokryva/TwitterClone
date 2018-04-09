package media

import(
    "github.com/mongodb/mongo-go-driver/bson/objectid"
)

type Media struct {
    ID objectid.ObjectID `bson:"_id"`
    Content []byte `bson:"content,omitempty"`
}

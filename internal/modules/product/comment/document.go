package comment

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Comment is a customer's review of a purchased product, stored in MongoDB
// (not Postgres) because photo/video attachments make it a variable-shape
// document. Username and ProductID are resolved from Postgres at write time
// and stored as a snapshot, not a live reference.
type Comment struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Username   string        `bson:"username" json:"username"`
	ProductID  uint64        `bson:"product_id" json:"productId"`
	Text       string        `bson:"text" json:"text"`
	PhotoLinks []string      `bson:"photo_links" json:"photoLinks"`
	VideoLinks []string      `bson:"video_links" json:"videoLinks"`
	CreatedAt  time.Time     `bson:"created_at" json:"createdAt"`
}

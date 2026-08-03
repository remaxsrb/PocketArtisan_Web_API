package comment

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionName = "product_comments"

type Repository interface {
	Create(ctx context.Context, c *Comment) error
	ListByProduct(ctx context.Context, productID uint64, skip, limit int64) ([]*Comment, error)
	CountByProduct(ctx context.Context, productID uint64) (int64, error)
}

type MongoRepository struct {
	col *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) Repository {
	return &MongoRepository{col: db.Collection(collectionName)}
}

func (r *MongoRepository) Create(ctx context.Context, c *Comment) error {
	c.CreatedAt = time.Now().UTC()
	res, err := r.col.InsertOne(ctx, c)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		c.ID = id
	}
	return nil
}

func (r *MongoRepository) ListByProduct(ctx context.Context, productID uint64, skip, limit int64) ([]*Comment, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cur, err := r.col.Find(ctx, bson.M{"product_id": productID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	comments := make([]*Comment, 0)
	if err := cur.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *MongoRepository) CountByProduct(ctx context.Context, productID uint64) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{"product_id": productID})
}

// EnsureIndexes creates the index product-comment listing relies on (filter
// by product, sorted newest-first). Call once at startup; CreateOne is a
// no-op if an equivalent index already exists.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "product_id", Value: 1}, {Key: "created_at", Value: -1}},
	})
	return err
}

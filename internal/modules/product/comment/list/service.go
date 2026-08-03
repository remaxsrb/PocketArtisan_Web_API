package list

import (
	"PocketArtisan/internal/modules/product/comment"
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Service struct {
	comments comment.Repository
}

func NewService(mongoDB *mongo.Database) *Service {
	return &Service{comments: comment.NewMongoRepository(mongoDB)}
}

func (uc *Service) Execute(ctx context.Context, req Request) (Response, error) {
	raw, err := uc.comments.ListByProduct(ctx, req.ProductID, req.Skip, req.Limit)
	if err != nil {
		return Response{}, err
	}

	total, err := uc.comments.CountByProduct(ctx, req.ProductID)
	if err != nil {
		return Response{}, err
	}

	list := make([]*comment.Response, 0, len(raw))
	for _, c := range raw {
		list = append(list, comment.ToResponse(c))
	}
	return Response{Comments: list, Total: total}, nil
}

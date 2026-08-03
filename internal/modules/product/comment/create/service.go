package create

import (
	"PocketArtisan/internal/modules/product/comment"
	"context"
	"errors"

	prodmod "PocketArtisan/internal/modules/product"
	usersmod "PocketArtisan/internal/modules/users"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrNotPurchased    = errors.New("you can only comment on products from a completed purchase")
)

type Service struct {
	products prodmod.Repository
	users    usersmod.Repository
	comments comment.Repository
}

func NewService(db *gorm.DB, mongoDB *mongo.Database) *Service {
	return &Service{
		products: prodmod.NewGormRepository(db),
		users:    usersmod.NewGormRepository(db),
		comments: comment.NewMongoRepository(mongoDB),
	}
}

func (uc *Service) Execute(ctx context.Context, req Request) (*comment.Response, error) {
	customerID := ctx.Value("user_id").(uint64)

	if _, err := uc.products.FindByID(ctx, req.ProductID); err != nil {
		return nil, ErrProductNotFound
	}

	purchased, err := uc.products.HasPurchased(ctx, customerID, req.ProductID)
	if err != nil {
		return nil, err
	}
	if !purchased {
		return nil, ErrNotPurchased
	}

	user, err := uc.users.FindUserByID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	c := &comment.Comment{
		Username:   user.Username,
		ProductID:  req.ProductID,
		Text:       req.Text,
		PhotoLinks: req.PhotoLinks,
		VideoLinks: req.VideoLinks,
	}
	if err := uc.comments.Create(ctx, c); err != nil {
		return nil, err
	}

	return comment.ToResponse(c), nil
}

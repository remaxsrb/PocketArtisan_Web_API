package get_all

import (
	"PocketArtisan/internal/modules/users"
	"context"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UseCase struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUseCase(db *gorm.DB, cache *redis.Client) *UseCase {
	return &UseCase{db: db, cache: cache}
}

func (uc *UseCase) Execute(ctx context.Context, req GetAllRequest) (GetAllResponse, error) {

	const maxLimit = 100
	const defaultLimit = 20

	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	user_list := make([]*users.User, 0, req.Limit)

	query := uc.db.WithContext(ctx).Model(&user_list)

	if req.CursorAt != nil && req.ID != nil {
		if req.Direction == Prev {
			query = query.
				Where("(created_at, id) > (?, ?)", *req.CursorAt, *req.ID).
				Order("created_at ASC, id ASC")
		} else {
			query = query.
				Where("(created_at, id) < (?, ?)", *req.CursorAt, *req.ID).
				Order("created_at DESC, id DESC")
		}
	} else {
		query = query.Order("created_at DESC, id DESC")
	}

	if err := query.Limit(req.Limit).Find(&user_list).Error; err != nil {
		return GetAllResponse{}, err
	}

	// reverse results for "prev" so UI always gets consistent order
	if req.Direction == "prev" {
		for i, j := 0, len(user_list)-1; i < j; i, j = i+1, j-1 {
			user_list[i], user_list[j] = user_list[j], user_list[i]
		}
	}

	resp := GetAllResponse{
		Users: user_list,
	}

	if len(user_list) > 0 {
		first := user_list[0]
		last := user_list[len(user_list)-1]

		resp.PrevAt = &first.CreatedAt
		resp.PrevID = &first.ID

		resp.NextAt = &last.CreatedAt
		resp.NextID = &last.ID
	}

	return resp, nil

}

package ad

import (
	"context"
)

type AdInterface interface {
	Create(ctx context.Context, ad *Ad) (*Ad, error)
	List(ctx context.Context, params ListAdsParams) ([]listAdsResponseItem, error)
}

type Ad struct {
	ID          uint32  `json:"id"`
	UserID      uint32  `json:"user_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	SquadID     uint32  `json:"squad_id"`
}

type ListAdsParams struct {
	Limit       int
	Offset      int
	SortBy      string
	Order       string
	MinPrice    float64
	MaxPrice    float64
	UserFilter  bool
	UserID      uint32
	SquadFilter bool
	SquadID     uint32
}

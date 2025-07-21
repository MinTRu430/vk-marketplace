package ad

import "context"

type AdRepository interface {
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
}

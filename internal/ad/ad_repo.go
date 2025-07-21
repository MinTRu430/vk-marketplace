package ad

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdRepo struct {
	db *pgxpool.Pool
}

func NewAdRepo(db *pgxpool.Pool) *AdRepo {
	return &AdRepo{db: db}
}

func (r *AdRepo) Create(ctx context.Context, ad *Ad) (*Ad, error) {
	err := r.db.QueryRow(ctx, `
        INSERT INTO ads (user_id, title, description, image_url, price)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `, ad.UserID, ad.Title, ad.Description, ad.ImageURL, ad.Price).Scan(&ad.ID)
	if err != nil {
		return nil, fmt.Errorf("insert ad: %w", err)
	}
	return ad, nil
}

type ListAdsParams struct {
	Limit      int
	Offset     int
	SortBy     string  // "created_at" или "price"
	Order      string  // "asc" или "desc"
	MinPrice   float64 // 0 если не задан
	MaxPrice   float64 // 0 если не задан
	UserFilter bool    // фильтровать по userID
	UserID     uint32
}

func (r *AdRepo) List(ctx context.Context, params ListAdsParams) ([]listAdsResponseItem, error) {
	args := []interface{}{}
	whereClauses := []string{}

	// Фильтр по цене
	if params.MinPrice > 0 {
		args = append(args, params.MinPrice)
		whereClauses = append(whereClauses, fmt.Sprintf("price >= $%d", len(args)))
	}
	if params.MaxPrice > 0 && params.MaxPrice >= params.MinPrice {
		args = append(args, params.MaxPrice)
		whereClauses = append(whereClauses, fmt.Sprintf("price <= $%d", len(args)))
	}

	// Фильтр по автору (если нужно)
	if params.UserFilter {
		args = append(args, params.UserID)
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", len(args)))
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Формируем запрос с JOIN к таблице пользователей чтобы получить login
	query := fmt.Sprintf(`
		SELECT
			ads.id, ads.title, ads.description, ads.image_url, ads.price, users.login
		FROM ads
		JOIN users ON ads.user_id = users.id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, params.SortBy, params.Order, len(args)+1, len(args)+2)

	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ads []listAdsResponseItem
	for rows.Next() {
		var ad listAdsResponseItem
		err := rows.Scan(&ad.ID, &ad.Title, &ad.Description, &ad.ImageURL, &ad.Price, &ad.AuthorLogin)
		if err != nil {
			return nil, err
		}
		ads = append(ads, ad)
	}
	return ads, nil
}

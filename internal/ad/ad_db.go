package ad

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdDB struct {
	db *pgxpool.Pool
}

func NewAdRepo(db *pgxpool.Pool) *AdDB {
	return &AdDB{db: db}
}

func (r *AdDB) Create(ctx context.Context, ad *Ad) (*Ad, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO ads (user_id, title, description, image_url, price, squad_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, ad.UserID, ad.Title, ad.Description, ad.ImageURL, ad.Price, ad.SquadID).Scan(&ad.ID)

	if err != nil {
		return nil, fmt.Errorf("insert ad: %w", err)
	}

	return ad, nil
}

func (r *AdDB) List(ctx context.Context, params ListAdsParams) ([]listAdsResponseItem, error) {
	validSortBy := map[string]string{
		"created_at": "ads.created_at",
		"price":      "ads.price",
	}
	validOrder := map[string]string{
		"ASC":  "ASC",
		"DESC": "DESC",
	}

	sortBy, ok := validSortBy[params.SortBy]
	if !ok {
		sortBy = "ads.created_at"
	}

	order, ok := validOrder[strings.ToUpper(params.Order)]
	if !ok {
		order = "DESC"
	}

	args := []interface{}{}
	whereClauses := []string{}

	if params.SquadFilter {
		args = append(args, params.SquadID)
		whereClauses = append(whereClauses, fmt.Sprintf("ads.squad_id = $%d", len(args)))
	}

	if params.MinPrice > 0 {
		args = append(args, params.MinPrice)
		whereClauses = append(whereClauses, fmt.Sprintf("price >= $%d", len(args)))
	}
	if params.MaxPrice > 0 && params.MaxPrice >= params.MinPrice {
		args = append(args, params.MaxPrice)
		whereClauses = append(whereClauses, fmt.Sprintf("price <= $%d", len(args)))
	}

	if params.UserFilter {
		args = append(args, params.UserID)
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", len(args)))
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	query := fmt.Sprintf(`
			SELECT
				ads.id, ads.title, ads.description, ads.image_url, ads.price, users.login
			FROM ads
			JOIN users ON ads.user_id = users.id
			%s
			ORDER BY %s %s
			LIMIT $%d OFFSET $%d
		`, whereSQL, sortBy, order, len(args)+1, len(args)+2)

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

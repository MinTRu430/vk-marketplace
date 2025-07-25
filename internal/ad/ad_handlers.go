package ad

import (
	"VK/internal/session"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// // import (
// // 	"VK/internal/session"
// // 	"encoding/json"
// // 	"net/http"
// // 	"strconv"
// // 	"strings"
// // )

type AdHandler struct {
	Ads      AdInterface
	Sessions session.SessionManager
}

type createAdRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
}

func (h *AdHandler) CreateAd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, err := h.Sessions.Check(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req createAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Price <= 0 {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	ad := &Ad{
		UserID:      sess.UserID,
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		SquadID:     sess.SquadID,
	}

	createdAd, err := h.Ads.Create(r.Context(), ad)
	if err != nil {
		http.Error(w, "Failed to create ad", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAd)
}

type listAdsResponseItem struct {
	ID          uint32  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	AuthorLogin string  `json:"author_login"`
}

func (h *AdHandler) ListAds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	page, err := strconv.Atoi(q.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	sortBy := strings.ToLower(q.Get("sort_by"))
	if sortBy != "price" && sortBy != "created_at" {
		sortBy = "created_at"
	}

	order := strings.ToLower(q.Get("order"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	order = strings.ToUpper(order)

	minPrice, err := strconv.ParseFloat(q.Get("min_price"), 64)
	if err != nil {
		minPrice = 0
	}
	maxPrice, err := strconv.ParseFloat(q.Get("max_price"), 64)
	if err != nil {
		maxPrice = 0
	}

	if maxPrice > 0 && maxPrice < minPrice {
		http.Error(w, "max_price must be greater or equal to min_price", http.StatusBadRequest)
		return
	}

	var userID uint32
	var isAuthUser bool
	if strings.ToLower(q.Get("my")) == "true" {
		sess, err := h.Sessions.Check(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID = sess.UserID
		isAuthUser = true
	}

	squadFilter := false
	var squadID uint32
	if strings.ToLower(q.Get("squad")) == "true" {
		sess, err := h.Sessions.Check(r)
		if err == nil {
			squadID = sess.SquadID
			squadFilter = true
		}
	}

	ads, err := h.Ads.List(r.Context(), ListAdsParams{
		Limit:       limit,
		Offset:      offset,
		SortBy:      sortBy,
		Order:       order,
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		UserFilter:  isAuthUser,
		UserID:      userID,
		SquadFilter: squadFilter,
		SquadID:     squadID,
	})
	if err != nil {
		http.Error(w, "Failed to get ads", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ads)
}

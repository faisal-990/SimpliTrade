package controllers_test

// E2E for the social/investor endpoints: real router + auth middleware +
// InvestorHandler + InvestorService. Only the repository is faked. Reuses
// do/dataOf from authflow_test.go (same test package).

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/controllers"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeInvestorRepo struct {
	investors []repository.InvestorSummary
	trades    []models.Trade
	feed      []models.Trade
	follows   map[[2]uuid.UUID]bool // [follower, investor]
}

func (f *fakeInvestorRepo) ListInvestors(_ context.Context, limit, offset int) ([]repository.InvestorSummary, error) {
	return f.investors, nil
}
func (f *fakeInvestorRepo) GetInvestor(_ context.Context, id uuid.UUID) (*repository.InvestorSummary, error) {
	for i := range f.investors {
		if f.investors[i].ID == id {
			return &f.investors[i], nil
		}
	}
	return nil, repository.ErrNotFound
}
func (f *fakeInvestorRepo) ListInvestorTrades(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Trade, error) {
	return f.trades, nil
}
func (f *fakeInvestorRepo) Follow(_ context.Context, follower, investor uuid.UUID) error {
	f.follows[[2]uuid.UUID{follower, investor}] = true
	return nil
}
func (f *fakeInvestorRepo) Unfollow(_ context.Context, follower, investor uuid.UUID) error {
	delete(f.follows, [2]uuid.UUID{follower, investor})
	return nil
}
func (f *fakeInvestorRepo) ListFollowing(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (f *fakeInvestorRepo) FeedTrades(_ context.Context, _ uuid.UUID, _ int) ([]models.Trade, error) {
	return f.feed, nil
}

func newSocialRouter(repo repository.InvestorRepo) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := controllers.NewInvestorHandler(service.NewInvestorService(repo))
	mw := middlewares.AuthMiddleware(tm)

	r := gin.New()
	g := r.Group("/api/investor")
	g.Use(mw)
	{
		g.GET("/", h.HandleGetInvestor)
		g.GET("/:id", h.HandleGetInvestorById)
		g.GET("/:id/trades", h.HandleGetInvestorTrades)
		g.POST("/:id/follow", h.HandleFollowInvestor)
		g.DELETE("/:id/follow", h.HandleUnfollowInvestor)
	}
	r.GET("/api/feed", mw, h.HandleGetFeed)

	token, _, _ := tm.GenerateAccessToken(uuid.New().String(), uuid.New().String(), "user")
	return r, token
}

func TestSocial_Leaderboard(t *testing.T) {
	repo := &fakeInvestorRepo{
		investors: []repository.InvestorSummary{
			{ID: uuid.New(), Name: "Warren Buffett", Strategy: "quality_value", ROI: 0.12, Rank: 1},
			{ID: uuid.New(), Name: "Jesse Livermore", Strategy: "trend_momentum", ROI: -0.05, Rank: 2},
		},
		follows: map[[2]uuid.UUID]bool{},
	}
	r, token := newSocialRouter(repo)

	status, body := do(t, r, http.MethodGet, "/api/investor/", token, nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	items, ok := body["data"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected 2 investors, got %v", body["data"])
	}
	first := items[0].(map[string]any)
	if first["name"] != "Warren Buffett" || first["rank"].(float64) != 1 {
		t.Errorf("unexpected leader: %v", first)
	}
}

func TestSocial_ProfileAndNotFound(t *testing.T) {
	id := uuid.New()
	repo := &fakeInvestorRepo{
		investors: []repository.InvestorSummary{{ID: id, Name: "Cathie Wood", Strategy: "disruptive_growth", ROI: 0.3, Rank: 1}},
		follows:   map[[2]uuid.UUID]bool{},
	}
	r, token := newSocialRouter(repo)

	status, body := do(t, r, http.MethodGet, "/api/investor/"+id.String(), token, nil)
	if status != http.StatusOK {
		t.Fatalf("profile status = %d, want 200", status)
	}
	if dataOf(t, body)["name"] != "Cathie Wood" {
		t.Errorf("profile name = %v", dataOf(t, body)["name"])
	}

	status, _ = do(t, r, http.MethodGet, "/api/investor/"+uuid.New().String(), token, nil)
	if status != http.StatusNotFound {
		t.Fatalf("unknown investor status = %d, want 404", status)
	}
}

func TestSocial_FollowUnfollow(t *testing.T) {
	investorID := uuid.New()
	repo := &fakeInvestorRepo{
		investors: []repository.InvestorSummary{{ID: investorID, Name: "Graham"}},
		follows:   map[[2]uuid.UUID]bool{},
	}
	r, token := newSocialRouter(repo)

	status, body := do(t, r, http.MethodPost, "/api/investor/"+investorID.String()+"/follow", token, nil)
	if status != http.StatusOK || dataOf(t, body)["following"] != true {
		t.Fatalf("follow: status %d body %v", status, body)
	}
	if len(repo.follows) != 1 {
		t.Errorf("expected one follow recorded, got %d", len(repo.follows))
	}

	status, body = do(t, r, http.MethodDelete, "/api/investor/"+investorID.String()+"/follow", token, nil)
	if status != http.StatusOK || dataOf(t, body)["following"] != false {
		t.Fatalf("unfollow: status %d body %v", status, body)
	}
	if len(repo.follows) != 0 {
		t.Errorf("expected follow removed, got %d", len(repo.follows))
	}
}

func TestSocial_Feed(t *testing.T) {
	investorID := uuid.New()
	repo := &fakeInvestorRepo{
		follows: map[[2]uuid.UUID]bool{},
		feed: []models.Trade{{
			Type: "buy", Quantity: 10, Price: 100, ExecutedAt: time.Now(),
			Stock:   models.Stock{Symbol: "AAPL"},
			Account: models.Account{UserID: investorID, User: models.User{Name: "Warren Buffett"}},
		}},
	}
	r, token := newSocialRouter(repo)

	status, body := do(t, r, http.MethodGet, "/api/feed", token, nil)
	if status != http.StatusOK {
		t.Fatalf("feed status = %d, want 200 (%v)", status, body)
	}
	items, ok := body["data"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected 1 feed item, got %v", body["data"])
	}
	item := items[0].(map[string]any)
	if item["investor_name"] != "Warren Buffett" || item["symbol"] != "AAPL" {
		t.Errorf("unexpected feed item: %v", item)
	}
}

func TestSocial_RequiresAuth(t *testing.T) {
	repo := &fakeInvestorRepo{follows: map[[2]uuid.UUID]bool{}}
	r, _ := newSocialRouter(repo)
	if status, _ := do(t, r, http.MethodGet, "/api/investor/", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", status)
	}
}

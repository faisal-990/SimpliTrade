package trading

import (
	"errors"
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestComputeBuy_FreshPosition(t *testing.T) {
	got, err := ComputeBuy(10000, Position{}, 100, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approx(got.Cost, 1000) {
		t.Errorf("cost = %v, want 1000", got.Cost)
	}
	if !approx(got.NewBalance, 9000) {
		t.Errorf("balance = %v, want 9000", got.NewBalance)
	}
	if !approx(got.NewPosition.Quantity, 10) || !approx(got.NewPosition.AvgPrice, 100) {
		t.Errorf("position = %+v, want qty 10 @ 100", got.NewPosition)
	}
}

func TestComputeBuy_AveragesCostBasis(t *testing.T) {
	// Hold 10 @ $100, buy 10 @ $200 -> 20 @ $150.
	got, err := ComputeBuy(5000, Position{Quantity: 10, AvgPrice: 100}, 200, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approx(got.NewPosition.Quantity, 20) {
		t.Errorf("qty = %v, want 20", got.NewPosition.Quantity)
	}
	if !approx(got.NewPosition.AvgPrice, 150) {
		t.Errorf("avg price = %v, want 150", got.NewPosition.AvgPrice)
	}
}

func TestComputeBuy_InsufficientFunds(t *testing.T) {
	if _, err := ComputeBuy(999, Position{}, 100, 10); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("err = %v, want ErrInsufficientFunds", err)
	}
}

func TestComputeBuy_ExactBalanceSucceeds(t *testing.T) {
	got, err := ComputeBuy(1000, Position{}, 100, 10)
	if err != nil {
		t.Fatalf("spending the exact balance should succeed, got %v", err)
	}
	if !approx(got.NewBalance, 0) {
		t.Errorf("balance = %v, want 0", got.NewBalance)
	}
}

func TestComputeBuy_RejectsNonPositive(t *testing.T) {
	if _, err := ComputeBuy(1000, Position{}, 100, 0); !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("zero qty: err = %v, want ErrInvalidQuantity", err)
	}
	if _, err := ComputeBuy(1000, Position{}, 0, 10); !errors.Is(err, ErrInvalidPrice) {
		t.Errorf("zero price: err = %v, want ErrInvalidPrice", err)
	}
}

func TestComputeSell_Partial(t *testing.T) {
	// Hold 10 @ $100, sell 4 @ $150.
	got, err := ComputeSell(0, Position{Quantity: 10, AvgPrice: 100}, 150, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approx(got.Proceeds, 600) {
		t.Errorf("proceeds = %v, want 600", got.Proceeds)
	}
	if !approx(got.NewBalance, 600) {
		t.Errorf("balance = %v, want 600", got.NewBalance)
	}
	if !approx(got.NewPosition.Quantity, 6) || !approx(got.NewPosition.AvgPrice, 100) {
		t.Errorf("position = %+v, want qty 6 @ 100 (avg unchanged on sell)", got.NewPosition)
	}
}

func TestComputeSell_FullCloseResetsAverage(t *testing.T) {
	got, err := ComputeSell(0, Position{Quantity: 10, AvgPrice: 100}, 120, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approx(got.NewPosition.Quantity, 0) || !approx(got.NewPosition.AvgPrice, 0) {
		t.Errorf("closed position = %+v, want qty 0 @ 0", got.NewPosition)
	}
}

func TestComputeSell_Oversell(t *testing.T) {
	if _, err := ComputeSell(0, Position{Quantity: 5, AvgPrice: 100}, 150, 6); !errors.Is(err, ErrInsufficientShares) {
		t.Fatalf("err = %v, want ErrInsufficientShares", err)
	}
}

func TestComputeSell_RejectsNonPositive(t *testing.T) {
	if _, err := ComputeSell(0, Position{Quantity: 5}, 100, -1); !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("negative qty: err = %v, want ErrInvalidQuantity", err)
	}
}

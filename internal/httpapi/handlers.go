package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"

	"calculator-service/internal/service"
)

type twoNumberRequest struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

type oneNumberResponse struct {
	Value float64 `json:"value"`
}

const maxAbs = 1e15

func (a *App) handleAdd(w http.ResponseWriter, r *http.Request) {
	a.handleTwoNumberOp(w, r, a.Calc.Add)
}
func (a *App) handleSub(w http.ResponseWriter, r *http.Request) {
	a.handleTwoNumberOp(w, r, a.Calc.Sub)
}
func (a *App) handleMul(w http.ResponseWriter, r *http.Request) {
	a.handleTwoNumberOp(w, r, a.Calc.Mul)
}
func (a *App) handleDiv(w http.ResponseWriter, r *http.Request) {
	a.handleTwoNumberOp(w, r, a.Calc.Div)
}

func (a *App) handleTwoNumberOp(
	w http.ResponseWriter,
	r *http.Request,
	fn func(context.Context, float64, float64) (float64, error),
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var req twoNumberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if invalidNumber(req.A) || invalidNumber(req.B) {
		http.Error(w, "invalid number", http.StatusBadRequest)
		return
	}

	res, err := fn(r.Context(), req.A, req.B)
	if err != nil {
		if errors.Is(err, service.ErrDivisionByZero) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if invalidNumber(res) {
		http.Error(w, "invalid result", http.StatusBadRequest)
		return
	}

	writeJSON(w, oneNumberResponse{Value: res})
}

func (a *App) handleRecent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	n := 5
	if q := r.URL.Query().Get("n"); q != "" {
		parsed, err := strconv.Atoi(q)
		if err != nil || parsed < 0 {
			http.Error(w, "invalid n", http.StatusBadRequest)
			return
		}
		n = parsed
	}
	if n > 20 {
		n = 20
	}

	res, err := a.Store.ListRecent(r.Context(), n)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, res)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func invalidNumber(v float64) bool {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return true
	}
	return math.Abs(v) > maxAbs
}

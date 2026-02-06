package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"calculator-service/internal/model"
)

// withMemory expects that withLogging has wrapped the ResponseWriter with *responseRecorder,
// so it can read status/body after next() without creating another recorder.
func (a *App) withMemory(op string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(body))

			var req twoNumberRequest
			if err := json.Unmarshal(body, &req); err != nil {
				next(w, r)
				return
			}

			next(w, r)

			rec, ok := w.(*responseRecorder)
			if !ok {
				return
			}

			status := rec.status
			if status == 0 {
				status = http.StatusOK
			}

			var resPtr *float64
			var errStr string

			if status >= 200 && status < 300 {
				var resp oneNumberResponse
				if json.Unmarshal(rec.body.Bytes(), &resp) == nil {
					resPtr = &resp.Value
				}
			} else {
				errStr = string(bytes.TrimSpace(rec.body.Bytes()))
			}

			entry := model.ResultEntry{
				Time:      time.Now(),
				Op:        model.Operation(op),
				A:         req.A,
				B:         req.B,
				Result:    resPtr,
				Error:     errStr,
				ExprHuman: exprHuman(op, req.A, req.B, resPtr, errStr),
			}

			_ = a.Store.Add(context.Background(), entry)
		}
	}
}

func exprHuman(op string, a1, b1 float64, res *float64, errStr string) string {
	sym := map[string]string{
		"add":      "+",
		"subtract": "-",
		"multiply": "*",
		"divide":   "/",
	}[op]
	if sym == "" {
		sym = "?"
	}

	if errStr != "" {
		return fmt.Sprintf("%.5f %s %.5f = error: %s", a1, sym, b1, errStr)
	}
	if res == nil {
		return fmt.Sprintf("%.5f %s %.5f = (no result)", a1, sym, b1)
	}
	return fmt.Sprintf("%.5f %s %.5f = %.5f", a1, sym, b1, *res)
}

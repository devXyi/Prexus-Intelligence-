package main

// risk.go — Proxies /risk/* routes to the Python data engine service
// Add to main.go routes: mux.HandleFunc("/risk/", riskProxy)

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

var dataEngineURL = os.Getenv("DATA_ENGINE_URL")   // e.g. https://meteorium-engine.onrender.com
var dataEngineKey = os.Getenv("DATA_ENGINE_KEY")   // shared secret

// riskProxy forwards /risk/* requests to the Python FastAPI data engine.
// The Go backend acts as a single entry point — frontend never calls the
// data engine directly, keeping the API key hidden server-side.
func riskProxy(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == http.MethodOptions {
		return
	}

	if dataEngineURL == "" {
		http.Error(w, `{"error":"data engine not configured"}`, http.StatusServiceUnavailable)
		return
	}

	// Verify JWT from frontend before proxying
	userID := authMiddlewareID(r)
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Build upstream URL
	upstream := dataEngineURL + r.URL.Path
	if r.URL.RawQuery != "" {
		upstream += "?" + r.URL.RawQuery
	}

	body, _ := io.ReadAll(r.Body)

	req, err := http.NewRequest(r.Method, upstream, bytes.NewReader(body))
	if err != nil {
		http.Error(w, `{"error":"proxy error"}`, http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", dataEngineKey)
	req.Header.Set("X-User-Id", userID) // pass authenticated user downstream

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"data engine unavailable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// authMiddlewareID extracts user_id from Bearer JWT without full middleware
// (reuses the jwtVerify logic already in auth.go)
func authMiddlewareID(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	claims, err := jwtVerify(token)
	if err != nil {
		return ""
	}
	return claims["sub"].(string)
}

// scoreSingleAsset is a convenience direct call (no proxy) if data engine
// is colocated with the Go backend on the same server.
// Useful for Render internal networking to avoid extra HTTP hop.
func scoreSingleAsset(assetID, countryCode string, lat, lon, valueMm float64, horizonDays int) (map[string]interface{}, error) {
	payload, _ := json.Marshal(map[string]interface{}{
		"asset_id":     assetID,
		"lat":          lat,
		"lon":          lon,
		"country_code": countryCode,
		"value_mm":     valueMm,
		"horizon_days": horizonDays,
		"scenario":     "baseline",
	})

	req, err := http.NewRequest("POST", dataEngineURL+"/risk/asset", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", dataEngineKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

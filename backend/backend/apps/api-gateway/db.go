package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

/* ═══════════════════════════════════════════════════
   Asset model
═══════════════════════════════════════════════════ */

type Asset struct {
	ID       interface{} `json:"id,omitempty"`
	UserID   string      `json:"user_id,omitempty"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Lat      float64     `json:"lat"`
	Lon      float64     `json:"lon"`
	Country  string      `json:"country"`
	CC       string      `json:"cc"`
	ValueMm  float64     `json:"value_mm"`
	PR       float64     `json:"pr"`
	TR       float64     `json:"tr"`
	CR       float64     `json:"cr"`
	Alerts   int         `json:"alerts"`
}

/* ═══════════════════════════════════════════════════
   GET /assets  — list all assets for authenticated user
═══════════════════════════════════════════════════ */

func assetsGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Header.Get("X-User-Id")

	b, status, err := supaReq("GET", "assets",
		"user_id=eq."+userID+"&order=created_at.asc", "")
	if err != nil || status >= 400 {
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("fetch failed (%d): %s", status, string(b)),
		})
		return
	}

	var assets []map[string]interface{}
	json.Unmarshal(b, &assets)
	if assets == nil {
		assets = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(assets)
}

/* ═══════════════════════════════════════════════════
   POST /assets  — create a new asset
═══════════════════════════════════════════════════ */

func assetsPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Header.Get("X-User-Id")

	var a Asset
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid body"})
		return
	}
	if strings.TrimSpace(a.Name) == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "asset name is required"})
		return
	}
	a.UserID = userID

	body, _ := json.Marshal(a)
	b, status, err := supaReq("POST", "assets", "", string(body))
	if err != nil || status >= 400 {
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("create failed (%d): %s", status, string(b)),
		})
		return
	}

	var created []map[string]interface{}
	json.Unmarshal(b, &created)
	if len(created) == 0 {
		json.NewEncoder(w).Encode(map[string]string{"error": "asset creation failed"})
		return
	}
	json.NewEncoder(w).Encode(created[0])
}

/* ═══════════════════════════════════════════════════
   PUT /assets/{id}  — update an asset (must belong to user)
═══════════════════════════════════════════════════ */

func assetsPutHandler(w http.ResponseWriter, r *http.Request, assetID string) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Header.Get("X-User-Id")

	var a Asset
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid body"})
		return
	}
	a.UserID = userID

	body, _ := json.Marshal(a)
	b, status, err := supaReq("PATCH", "assets",
		"id=eq."+assetID+"&user_id=eq."+userID, string(body))
	if err != nil || status >= 400 {
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("update failed (%d): %s", status, string(b)),
		})
		return
	}

	var updated []map[string]interface{}
	json.Unmarshal(b, &updated)
	if len(updated) == 0 {
		json.NewEncoder(w).Encode(map[string]string{"error": "asset not found or unauthorized"})
		return
	}
	json.NewEncoder(w).Encode(updated[0])
}

/* ═══════════════════════════════════════════════════
   DELETE /assets/{id}  — delete an asset (must belong to user)
═══════════════════════════════════════════════════ */

func assetsDeleteHandler(w http.ResponseWriter, r *http.Request, assetID string) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Header.Get("X-User-Id")

	_, status, err := supaReq("DELETE", "assets",
		"id=eq."+assetID+"&user_id=eq."+userID, "")
	if err != nil || status >= 400 {
		json.NewEncoder(w).Encode(map[string]string{"error": "delete failed"})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"ok": "deleted"})
}

/* ═══════════════════════════════════════════════════
   assetsRouter — routes /assets and /assets/{id}
═══════════════════════════════════════════════════ */

func assetsRouter(w http.ResponseWriter, r *http.Request) {
	// Strip "/assets" prefix to get optional /{id}
	path := strings.TrimPrefix(r.URL.Path, "/assets")
	path = strings.Trim(path, "/")

	if path == "" {
		// /assets
		switch r.Method {
		case "GET":
			assetsGetHandler(w, r)
		case "POST":
			assetsPostHandler(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
	} else {
		// /assets/{id}
		switch r.Method {
		case "PUT", "PATCH":
			assetsPutHandler(w, r, path)
		case "DELETE":
			assetsDeleteHandler(w, r, path)
		default:
			http.Error(w, "method not allowed", 405)
		}
	}
}

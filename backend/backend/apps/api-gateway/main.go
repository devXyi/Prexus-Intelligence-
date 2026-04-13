// backend/apps/api-gateway/main.go
// Prexus Intelligence — API Gateway (RBAC Integrated)

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

const VERSION = "2.1.0"

// ── Allowed Claude models ──────────────────────────────────

var allowedModels = map[string]struct{}{
	"claude-opus-4-5":   {},
	"claude-sonnet-4-5": {},
	"claude-haiku-4-5":  {},
}

const defaultModel = "claude-haiku-4-5"

// ── Rate limiter store (per-IP) ────────────────────────────

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	limiters   = make(map[string]*ipLimiter)
	limitersMu sync.Mutex
)

// getLimiter returns (or creates) a per-IP rate limiter.
// Allows 5 requests/second with a burst of 10.
func getLimiter(ip string) *rate.Limiter {
	limitersMu.Lock()
	defer limitersMu.Unlock()

	if il, ok := limiters[ip]; ok {
		il.lastSeen = time.Now()
		return il.limiter
	}

	l := rate.NewLimiter(5, 10)
	limiters[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

// cleanupLimiters removes stale IP entries every 5 minutes.
// Exits cleanly when ctx is cancelled (e.g. on server shutdown).
func cleanupLimiters(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			limitersMu.Lock()
			for ip, il := range limiters {
				if time.Since(il.lastSeen) > 10*time.Minute {
					delete(limiters, ip)
				}
			}
			limitersMu.Unlock()
		}
	}
}

// RateLimitMiddleware applies per-IP rate limiting.
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !getLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded — slow down",
			})
			return
		}
		c.Next()
	}
}

// ── Body size cap ──────────────────────────────────────────

const maxBodyBytes = 1 << 20 // 1 MB

func BodySizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		c.Next()
	}
}

// ── main ───────────────────────────────────────────────────

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── Validate data engine URL at startup ───────────────
	dataEngineURL := getDataEngineURL()
	if dataEngineURL == "" {
		log.Fatal("DATA_ENGINE_URL is not set — cannot start")
	}

	// ── Database ──────────────────────────────────────────
	if err := InitDB(); err != nil {
		log.Fatalf("Database init failed: %v", err)
	}
	defer CloseDB()

	log.Printf("✓ Database connected")
	log.Printf("✓ Data engine: %s", dataEngineURL)

	// ── CORS origins from env ─────────────────────────────
	allowedOrigins := getAllowedOrigins()
	log.Printf("✓ CORS origins: %v", allowedOrigins)

	// ── Start limiter GC (context-aware, shuts down cleanly) ─
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	go cleanupLimiters(ctx)

	// ── Router ────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestID())
	r.Use(BodySizeMiddleware())
	r.Use(requestLogger())

	// Trust only known upstream proxy IPs.
	// Set to your Render/Nginx egress IP in production via TRUSTED_PROXIES env,
	// or nil to disable proxy trust entirely (safest default).
	if tp := os.Getenv("TRUSTED_PROXIES"); tp != "" {
		if err := r.SetTrustedProxies(strings.Split(tp, ",")); err != nil {
			log.Fatalf("Invalid TRUSTED_PROXIES: %v", err)
		}
		log.Printf("✓ Trusted proxies: %s", tp)
	} else {
		_ = r.SetTrustedProxies(nil) // c.ClientIP() returns real IP directly
		log.Printf("✓ Trusted proxies: none (direct connections only)")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// ── Public Routes ─────────────────────────────────────
	r.GET("/health", RateLimitMiddleware(), handleHealth)

	// Rate-limited public auth endpoints
	r.POST("/register", RateLimitMiddleware(), handleRegister)
	r.POST("/login", RateLimitMiddleware(), handleLogin)

	// NOTE: /claude removed from public routes intentionally.
	// All AI access requires authentication — see protected group below.

	// ── Protected Routes (Auth + RBAC) ────────────────────
	auth := r.Group("/", AuthMiddleware())
	{
		// ── Assets ────────────────────────────
		auth.GET("/assets",
			RequirePermission("assets:read"),
			handleGetAssets,
		)

		auth.POST("/assets",
			RequirePermission("assets:create"),
			handleCreateAsset,
		)

		auth.PUT("/assets/:id",
			RequirePermission("assets:update"),
			handleUpdateAsset,
		)

		auth.DELETE("/assets/:id",
			RequirePermission("assets:delete"),
			handleDeleteAsset,
		)

		// ── Risk Engine ───────────────────────
		auth.POST("/risk/asset",
			RequirePermission("risk:run"),
			proxyToDataEngine("/risk/asset"),
		)

		auth.POST("/risk/portfolio",
			RequirePermission("risk:run"),
			proxyToDataEngine("/risk/portfolio"),
		)

		auth.POST("/risk/stress-test",
			RequirePermission("risk:run"),
			proxyToDataEngine("/risk/stress-test"),
		)

		auth.GET("/risk/health",
			RequirePermission("risk:run"),
			proxyToDataEngineGET("/risk/health"),
		)

		// ── AI (rate-limited + auth) ──────────
		auth.POST("/chat",
			RateLimitMiddleware(),
			RequirePermission("risk:run"),
			proxyToDataEngine("/chat"),
		)

		auth.POST("/claude",
			RateLimitMiddleware(),
			RequirePermission("risk:run"),
			handleClaude,
		)

		auth.POST("/analyze",
			RateLimitMiddleware(),
			RequirePermission("risk:run"),
			proxyToDataEngine("/analyze"),
		)

		// ── User ──────────────────────────────
		// No additional permission beyond valid session — intentional.
		// Document here so it isn't missed in future RBAC audits.
		auth.GET("/me", handleGetMe)
		auth.PUT("/me", handleUpdateMe)
	}

	log.Printf("🚀 Prexus API Gateway v%s running on :%s (env=%s)", VERSION, port, env)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ── Claude Handler ─────────────────────────────────────────

func handleClaude(c *gin.Context) {
	// Enforce body size (belt-and-suspenders alongside middleware)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	var req struct {
		Message string `json:"message" binding:"required"`
		Model   string `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Validate model against allowlist
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = defaultModel
	}
	if _, ok := allowedModels[model]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unsupported model",
			"allowed": getAllowedModelList(),
		})
		return
	}

	// Extract caller identity for audit log
	// Key matches AuthMiddleware: c.Set("user_id", ...)
	userID, _ := c.Get("user_id")
	reqID, _ := c.Get("request_id")
	log.Printf("[claude] req=%v user=%v model=%s ip=%s msg_len=%d",
		reqID, userID, model, c.ClientIP(), len(req.Message))

	// Enforce a hard timeout so a hung Claude call never stalls the worker
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	reply, err := AnalyzeProbability(ctx, req.Message, model)
	if err != nil {
		log.Printf("[claude] error req=%v user=%v: %v", reqID, userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI inference failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

// ── Health ─────────────────────────────────────────────────

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"service":   "prexus-api-gateway",
		"version":   VERSION,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ── Logger ─────────────────────────────────────────────────

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		userID, exists := c.Get("user_id")
		if !exists {
			userID = "anonymous"
		}
		reqID, _ := c.Get("request_id")

		log.Printf("[%d] %s %s req=%v user=%v ip=%s latency=%v",
			c.Writer.Status(),
			c.Request.Method,
			c.Request.URL.Path,
			reqID,
			userID,
			c.ClientIP(),
			latency,
		)
	}
}

// ── Helpers ────────────────────────────────────────────────

// getAllowedOrigins reads ALLOWED_ORIGINS from env (comma-separated).
// Falls back to localhost for local dev only.
func getAllowedOrigins() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		log.Println("⚠️  ALLOWED_ORIGINS not set — defaulting to localhost (dev only)")
		return []string{"http://localhost:3000", "http://localhost:5173"}
	}

	origins := []string{}
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}

// getAllowedModelList returns allowedModels keys sorted alphabetically.
// Sorted output means deterministic error responses — no random ordering.
func getAllowedModelList() []string {
	list := make([]string, 0, len(allowedModels))
	for m := range allowedModels {
		list = append(list, m)
	}
	sort.Strings(list)
	return list
}

// ── Request ID ─────────────────────────────────────────────

// RequestID stamps every request with a nanosecond-precision ID.
// Emitted as X-Request-ID header so clients can correlate logs end-to-end.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		c.Set("request_id", id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

// ── Banner (suppressed in production) ─────────────────────

func init() {
	if os.Getenv("ENV") == "production" {
		return
	}
	fmt.Printf(`
╔══════════════════════════════════════════╗
║   PREXUS INTELLIGENCE — API GATEWAY     ║
║   Version %-30s ║
╚══════════════════════════════════════════╝
`, VERSION)
}

// backend/apps/api-gateway/main.go
// Prexus Intelligence — API Gateway (FIXED)
// Added /claude route + optional public AI access for testing.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const VERSION = "2.0.1"

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

	// ── Database ─────────────────────────────────────────────
	if err := InitDB(); err != nil {
		log.Fatalf("Database init failed: %v", err)
	}
	defer CloseDB()

	log.Printf("✓ Database connected")
	log.Printf("✓ Data engine: %s", getDataEngineURL())

	// ── Router ───────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// ── Public routes ─────────────────────────────────────────
	r.GET("/health", handleHealth)
	r.POST("/register", handleRegister)
	r.POST("/login", handleLogin)

	// ✅ PUBLIC AI (for testing — remove later if needed)
	r.POST("/claude", proxyToDataEngine("/chat"))

	// ── Protected routes ──────────────────────────────────────
	auth := r.Group("/", AuthMiddleware())
	{
		// Assets
		auth.GET("/assets",        handleGetAssets)
		auth.POST("/assets",       handleCreateAsset)
		auth.PUT("/assets/:id",    handleUpdateAsset)
		auth.DELETE("/assets/:id", handleDeleteAsset)

		// Risk
		auth.POST("/risk/asset",       proxyToDataEngine("/risk/asset"))
		auth.POST("/risk/portfolio",   proxyToDataEngine("/risk/portfolio"))
		auth.POST("/risk/stress-test", proxyToDataEngine("/risk/stress-test"))
		auth.GET("/risk/health",       proxyToDataEngineGET("/risk/health"))

		// AI (protected)
		auth.POST("/chat",    proxyToDataEngine("/chat"))
		auth.POST("/analyze", proxyToDataEngine("/analyze"))

		// Optional: protected /claude also
		auth.POST("/claude", proxyToDataEngine("/chat"))

		// User
		auth.GET("/me", handleGetMe)
		auth.PUT("/me", handleUpdateMe)
	}

	log.Printf("🚀 Prexus API Gateway v%s running on :%s (env=%s)", VERSION, port, env)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
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

		log.Printf("[%d] %s %s %v",
			c.Writer.Status(),
			c.Request.Method,
			c.Request.URL.Path,
			latency,
		)
	}
}

// ── Banner ─────────────────────────────────────────────────

func init() {
	fmt.Printf(`
╔══════════════════════════════════════════╗
║   PREXUS INTELLIGENCE — API GATEWAY     ║
║   Version %-30s ║
╚══════════════════════════════════════════╝
`, VERSION)
}
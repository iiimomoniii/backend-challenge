package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	appauth "github.com/iiimomoniii/backend-challenge/internal/application/auth"
	appuser "github.com/iiimomoniii/backend-challenge/internal/application/user"
	appworker "github.com/iiimomoniii/backend-challenge/internal/application/worker"
	jwtprovider "github.com/iiimomoniii/backend-challenge/internal/infrastructure/auth/jwt"
	mongouser "github.com/iiimomoniii/backend-challenge/internal/infrastructure/database/mongodb/user"
	bcrypthasher "github.com/iiimomoniii/backend-challenge/internal/infrastructure/hasher/bcrypt"
	httpapi "github.com/iiimomoniii/backend-challenge/internal/interface/http"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/handler"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/middleware"
)

// อ่านมาจาก environment variables พร้อมค่า default รันบนเครื่อง local
type config struct {
	HTTPPort        string
	MongoURI        string
	MongoDatabase   string
	MongoCollection string
	MongoTimeout    time.Duration
	JWTSecret       string
	JWTTTL          time.Duration
	CounterInterval time.Duration
	ShutdownTimeout time.Duration
}

// loadConfig อ่านค่า configจาก env
func loadConfig() config {
	cfg := config{
		HTTPPort:        getEnv("HTTP_PORT", "8085"),
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:   getEnv("MONGO_DATABASE", "backend_challenge"),
		MongoCollection: getEnv("MONGO_USER_COLLECTION", "users"),
		MongoTimeout:    getEnvDuration("MONGO_TIMEOUT", 10*time.Second),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:          getEnvDuration("JWT_TTL", 24*time.Hour),
		CounterInterval: getEnvDuration("USER_COUNTER_INTERVAL", 30*time.Second),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}

	// เตือนถ้ายังใช้ JWT secret ค่า default อยู่ (ไม่ควรใช้ตอน deploy จริง)
	if cfg.JWTSecret == "dev-secret-change-me" {
		log.Println("warning: JWT_SECRET is not set, using an insecure default. Set JWT_SECRET before deploying to production.")
	}

	return cfg
}

// getEnv อ่านค่า env variable ตาม key
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvDuration อ่านค่า environment variable แล้ว parse เป็น time.Duration
func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("warning: invalid duration for %s=%q, using default %s (%v)", key, v, fallback, err)
		return fallback
	}
	return d
}

func main() {
	cfg := loadConfig()

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// เชื่อมต่อ MongoDB
	connectCtx, cancelConnect := context.WithTimeout(rootCtx, cfg.MongoTimeout)
	defer cancelConnect()

	mongoClient, err := mongo.Connect(connectCtx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	if err := mongoClient.Ping(connectCtx, nil); err != nil {
		log.Fatalf("failed to ping mongodb: %v", err)
	}
	log.Printf("connected to mongodb (database=%s)", cfg.MongoDatabase)

	// ปิดการเชื่อมต่อ mongodb ก่อนออกจาก app นี้
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			log.Printf("error disconnecting from mongodb: %v", err)
		}
	}()

	userCollection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)

	// สร้าง unique index บน email เพื่อป้องกันข้อมูลซ้ำในระดับ database
	indexCtx, cancelIndex := context.WithTimeout(rootCtx, cfg.MongoTimeout)
	defer cancelIndex()
	if err := ensureIndexes(indexCtx, userCollection); err != nil {
		log.Fatalf("failed to ensure mongodb indexes: %v", err)
	}

	// Infrastructure layer
	userRepo := mongouser.New(userCollection)
	hasher := bcrypthasher.New()
	tokenProvider := jwtprovider.New(cfg.JWTSecret, cfg.JWTTTL)

	//  Application layer (use cases)
	createUC := appuser.NewCreateUseCase(userRepo, hasher)
	searchUC := appuser.NewSearchUseCase(userRepo)
	searchAllUC := appuser.NewSearchAllUseCase(userRepo)
	updateUC := appuser.NewUpdateUseCase(userRepo)
	deleteUC := appuser.NewDeleteUseCase(userRepo)
	loginUC := appauth.NewLoginUseCase(userRepo, hasher, tokenProvider)

	//  Interface layer (HTTP)
	userHandler := handler.NewUserHandler(createUC, searchUC, searchAllUC, updateUC, deleteUC)
	authHandler := handler.NewAuthHandler(loginUC)
	authMiddleware := middleware.NewAuth(tokenProvider)

	router := httpapi.NewRouter(userHandler, authHandler, authMiddleware)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	//  Background worker
	var wg sync.WaitGroup
	counter := appworker.NewCounter(userRepo, cfg.CounterInterval)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("user counter worker started (interval=%s)", cfg.CounterInterval)
		counter.Run(rootCtx)
		log.Println("user counter worker stopped")
	}()

	//  HTTP server
	serverErrCh := make(chan error, 1)
	go func() {
		log.Printf("http server listening on :%s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	// รอจนกว่าจะได้รับ shutdown signal หรือ server ทำงานผิดพลาด
	select {
	case <-rootCtx.Done():
		log.Println("shutdown signal received")
	case err := <-serverErrCh:
		if err != nil {
			log.Printf("http server error: %v", err)
		}
	}

	// หยุดรับ signal เพิ่ม (เผื่อกด Ctrl+C ซ้ำระหว่าง shutdown) และ cancel rootCtx
	// เพื่อให้ background worker เริ่ม shutdown ไปพร้อมกัน
	stop()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("error during http server shutdown: %v", err)
		_ = server.Close()
	}

	// รอให้ background worker หยุดทำงานให้เรียบร้อยก่อนออกจาก app นี้
	wg.Wait()

	log.Println("shutdown complete")
}

// unique index บน email เพื่อป้องกันข้อมูล user ซ้ำในระดับ database
func ensureIndexes(ctx context.Context, collection *mongo.Collection) error {
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_email"),
	})
	return err
}

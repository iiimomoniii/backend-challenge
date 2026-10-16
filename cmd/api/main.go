package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	appauth "github.com/iiimomoniii/backend-challenge/internal/application/auth"
	appuser "github.com/iiimomoniii/backend-challenge/internal/application/user"
	appworker "github.com/iiimomoniii/backend-challenge/internal/application/worker"
	domainauth "github.com/iiimomoniii/backend-challenge/internal/domain/auth"
	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	jwtprovider "github.com/iiimomoniii/backend-challenge/internal/infrastructure/auth/jwt"
	oauth2introspection "github.com/iiimomoniii/backend-challenge/internal/infrastructure/auth/oauth2"
	mongouser "github.com/iiimomoniii/backend-challenge/internal/infrastructure/database/mongodb/user"
	postgresuser "github.com/iiimomoniii/backend-challenge/internal/infrastructure/database/postgresql/user"
	bcrypthasher "github.com/iiimomoniii/backend-challenge/internal/infrastructure/hasher/bcrypt"
	httpapi "github.com/iiimomoniii/backend-challenge/internal/interface/http"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/handler"
	"github.com/iiimomoniii/backend-challenge/internal/interface/http/middleware"
	_ "github.com/lib/pq"
)

// อ่านมาจาก environment variables พร้อมค่า default รันบนเครื่อง local
type config struct {
	HTTPPort                string
	AuthMethod              string
	OAuth2IntrospectionURL  string
	OAuth2ClientID          string
	OAuth2ClientSecret      string
	OAuth2Timeout           time.Duration
	OAuth2AllowInsecureHTTP bool
	OAuth2LocalMock         bool
	OAuth2MockAccessToken   string
	OAuth2MockSubject       string
	OAuth2MockEmail         string
	DatabaseDriver          string
	PostgresURI             string
	MongoURI                string
	MongoDatabase           string
	MongoCollection         string
	MongoTimeout            time.Duration
	JWTSecret               string
	JWTTTL                  time.Duration
	CounterInterval         time.Duration
	ShutdownTimeout         time.Duration
}

// loadConfig อ่านค่า configจาก env
func loadConfig() config {
	cfg := config{
		HTTPPort: getEnv("HTTP_PORT", "8085"),

		AuthMethod: getEnv("AUTH_METHOD", "oauth2"), // Select the auth method: jwt or oauth2;
		//JWT
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:    getEnvDuration("JWT_TTL", 24*time.Hour),
		//oAuth2
		OAuth2IntrospectionURL:  getEnv("OAUTH2_INTROSPECTION_URL", ""),
		OAuth2ClientID:          getEnv("OAUTH2_CLIENT_ID", ""),
		OAuth2ClientSecret:      getEnv("OAUTH2_CLIENT_SECRET", ""),
		OAuth2Timeout:           getEnvDuration("OAUTH2_TIMEOUT", 5*time.Second),
		OAuth2AllowInsecureHTTP: getEnvBool("OAUTH2_ALLOW_INSECURE_HTTP", false),
		OAuth2LocalMock:         getEnvBool("OAUTH2_LOCAL_MOCK", false),
		OAuth2MockAccessToken:   getEnv("OAUTH2_MOCK_ACCESS_TOKEN", "local-test-token"),
		OAuth2MockSubject:       getEnv("OAUTH2_MOCK_SUBJECT", "local-user-1"),
		OAuth2MockEmail:         getEnv("OAUTH2_MOCK_EMAIL", "demo@example.test"),

		DatabaseDriver: getEnv("DATABASE_DRIVER", "postgresql"), // Select the database: mongodb or postgresql
		//Postgres
		PostgresURI: getEnv("POSTGRES_URI", "postgres://backend_challenge:backend_challenge@localhost:5432/backend_challenge?sslmode=disable"),
		//Mongo
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:   getEnv("MONGO_DATABASE", "backend_challenge"),
		MongoCollection: getEnv("MONGO_USER_COLLECTION", "users"),
		MongoTimeout:    getEnvDuration("MONGO_TIMEOUT", 10*time.Second),

		CounterInterval: getEnvDuration("USER_COUNTER_INTERVAL", 30*time.Second),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}

	// เตือนถ้ายังใช้ JWT secret ค่า default อยู่ (ไม่ควรใช้ตอน deploy จริง)
	if cfg.AuthMethod == "jwt" && cfg.JWTSecret == "dev-secret-change-me" {
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

func getEnvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Printf("warning: invalid boolean for %s=%q, using default %t (%v)", key, v, fallback, err)
		return fallback
	}
	return b
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

	// เลือก database adapter จาก DATABASE_DRIVER
	var userRepo domainuser.Repository
	var mongoClient *mongo.Client
	var postgresDB *sql.DB
	var err error
	connectCtx, cancelConnect := context.WithTimeout(rootCtx, cfg.MongoTimeout)
	defer cancelConnect()

	switch cfg.DatabaseDriver {
	case "mongodb":
		mongoClient, err = mongo.Connect(connectCtx, options.Client().ApplyURI(cfg.MongoURI))
		if err != nil {
			log.Fatalf("failed to connect to mongodb: %v", err)
		}
		if err := mongoClient.Ping(connectCtx, nil); err != nil {
			log.Fatalf("failed to ping mongodb: %v", err)
		}
		log.Printf("connected to mongodb (database=%s)", cfg.MongoDatabase)
		defer func() {
			disconnectCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer cancel()
			if err := mongoClient.Disconnect(disconnectCtx); err != nil {
				log.Printf("error disconnecting from mongodb: %v", err)
			}
		}()

		userCollection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)
		if err := ensureIndexes(connectCtx, userCollection); err != nil {
			log.Fatalf("failed to ensure mongodb indexes: %v", err)
		}
		userRepo = mongouser.New(userCollection)
	case "postgresql":
		postgresDB, err = sql.Open("postgres", cfg.PostgresURI)
		if err != nil {
			log.Fatalf("failed to configure postgresql: %v", err)
		}
		if err := postgresDB.PingContext(connectCtx); err != nil {
			log.Fatalf("failed to connect to postgresql: %v", err)
		}
		log.Println("connected to postgresql")
		defer func() {
			if err := postgresDB.Close(); err != nil {
				log.Printf("error closing postgresql connection: %v", err)
			}
		}()

		postgresRepo := postgresuser.New(postgresDB)
		if err := postgresRepo.EnsureSchema(connectCtx); err != nil {
			log.Fatalf("failed to ensure postgresql schema: %v", err)
		}
		userRepo = postgresRepo
	default:
		log.Fatalf("unsupported DATABASE_DRIVER %q (supported: mongodb, postgresql)", cfg.DatabaseDriver)
	}

	// Infrastructure layer
	hasher := bcrypthasher.New()
	var tokenIssuer domainauth.TokenIssuer
	var tokenVerifier domainauth.TokenVerifier
	var oauth2MockServer *http.Server
	var oauth2MockListener net.Listener
	switch cfg.AuthMethod {
	case "jwt":
		jwtTokens := jwtprovider.New(cfg.JWTSecret, cfg.JWTTTL)
		tokenIssuer = jwtTokens
		tokenVerifier = jwtTokens
		log.Println("authentication method: jwt")
	case "oauth2":
		if cfg.OAuth2LocalMock {
			const mockAddr = "127.0.0.1:8090"
			oauth2MockServer = oauth2introspection.NewLocalMockServer(
				mockAddr,
				cfg.OAuth2ClientID,
				cfg.OAuth2ClientSecret,
				cfg.OAuth2MockAccessToken,
				cfg.OAuth2MockSubject,
				cfg.OAuth2MockEmail,
			)
			oauth2MockListener, err = net.Listen("tcp", mockAddr)
			if err != nil {
				log.Fatalf("failed to start local OAuth2 mock on %s: %v", mockAddr, err)
			}
			log.Printf("local OAuth2 mock listening on http://%s", mockAddr)
			log.Printf("local OAuth2 test access token: %s", cfg.OAuth2MockAccessToken)
		}
		oauth2Tokens, err := oauth2introspection.NewIntrospector(
			cfg.OAuth2IntrospectionURL,
			cfg.OAuth2ClientID,
			cfg.OAuth2ClientSecret,
			cfg.OAuth2AllowInsecureHTTP,
			&http.Client{Timeout: cfg.OAuth2Timeout},
		)
		if err != nil {
			log.Fatalf("failed to configure OAuth2 token introspection: %v", err)
		}
		tokenVerifier = oauth2Tokens
		log.Println("authentication method: oauth2 token introspection (tokens are issued by the authorization server)")
	default:
		log.Fatalf("unsupported AUTH_METHOD %q (supported: jwt, oauth2)", cfg.AuthMethod)
	}

	//  Application layer (use cases)
	createUC := appuser.NewCreateUseCase(userRepo, hasher)
	searchUC := appuser.NewSearchUseCase(userRepo)
	searchAllUC := appuser.NewSearchAllUseCase(userRepo)
	updateUC := appuser.NewUpdateUseCase(userRepo)
	deleteUC := appuser.NewDeleteUseCase(userRepo)
	var authHandler *handler.AuthHandler
	if tokenIssuer != nil {
		loginUC := appauth.NewLoginUseCase(userRepo, hasher, tokenIssuer)
		authHandler = handler.NewAuthHandler(loginUC)
	}

	//  Interface layer (HTTP)
	userHandler := handler.NewUserHandler(createUC, searchUC, searchAllUC, updateUC, deleteUC)
	authMiddleware := middleware.NewAuth(tokenVerifier)

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
	if oauth2MockServer != nil {
		go func() {
			if err := oauth2MockServer.Serve(oauth2MockListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("local OAuth2 mock server error: %v", err)
			}
		}()
	}

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
	if oauth2MockServer != nil {
		if err := oauth2MockServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("error during local OAuth2 mock shutdown: %v", err)
			_ = oauth2MockServer.Close()
		}
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

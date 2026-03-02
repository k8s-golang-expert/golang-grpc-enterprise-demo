package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang-grpc-enterprise-demo/internal/config"
	"golang-grpc-enterprise-demo/internal/handler"
	"golang-grpc-enterprise-demo/internal/middleware"
	"golang-grpc-enterprise-demo/internal/model"
	"golang-grpc-enterprise-demo/internal/repository"
	"golang-grpc-enterprise-demo/internal/service"
	"golang-grpc-enterprise-demo/internal/telemetry"
	"golang-grpc-enterprise-demo/pkg/seed"
	userv1 "golang-grpc-enterprise-demo/proto/gen/user/v1"

	_ "golang-grpc-enterprise-demo/docs" // swagger generated docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title           Golang gRPC Enterprise Demo API
// @version         1.0
// @description     High-performance gRPC + REST Hybrid API with PostgreSQL, JWT, Docker & OpenTelemetry.
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Enter "Bearer {token}" (get token from /api/v1/login)
func main() {
	// ---------- Logger ----------
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// ---------- Config ----------
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// ---------- Telemetry ----------
	shutdown, err := telemetry.InitTracer()
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdown(context.Background())

	// ---------- Database ----------
	dsn := cfg.DSN()
	if dsn == "" || dsn == "host= port= user= password= dbname= sslmode=disable" {
		logger.Fatal("DATABASE_URL is empty — check Railway Variables",
			zap.String("DATABASE_URL_env", os.Getenv("DATABASE_URL")),
			zap.String("DB_HOST_env", os.Getenv("DB_HOST")),
		)
	}
	logger.Info("connecting to database", zap.Bool("using_DATABASE_URL", cfg.DatabaseURL != ""))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		logger.Fatal("failed to migrate", zap.Error(err))
	}
	seed.Run(db, logger)

	// ---------- Dependency Injection ----------
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	jwtMgr := middleware.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryH)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPM)

	// ---------- gRPC Server ----------
	grpcHandler := handler.NewGRPCHandler(userSvc)
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(jwtMgr.UnaryServerInterceptor()),
	)
	userv1.RegisterUserServiceServer(grpcServer, grpcHandler)

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			logger.Fatal("failed to listen gRPC", zap.Error(err))
		}
		logger.Info("gRPC server started", zap.String("port", cfg.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("gRPC serve error", zap.Error(err))
		}
	}()

	// ---------- HTTP / REST Server (Gin) ----------
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(rateLimiter.GinMiddleware())

	// Health check (no auth)
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger UI (no auth)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	api.Use(jwtMgr.GinAuth())

	restHandler := handler.NewRESTHandler(userSvc, jwtMgr)
	restHandler.Register(api)

	go func() {
		addr := ":" + cfg.HTTPPort
		logger.Info("HTTP server started", zap.String("port", cfg.HTTPPort))
		if err := router.Run(addr); err != nil {
			logger.Fatal("HTTP serve error", zap.Error(err))
		}
	}()

	// ---------- Graceful Shutdown ----------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	fmt.Println()
	logger.Info("shutting down", zap.String("signal", sig.String()))
	grpcServer.GracefulStop()
	logger.Info("server stopped")
}

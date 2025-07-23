package app

import (
	"context"
	"fmt"
	"net"

	"github.com/AlexMickh/proj-protos/pkg/api/user"
	"github.com/AlexMickh/proj-user/internal/config"
	"github.com/AlexMickh/proj-user/internal/grpc/server"
	"github.com/AlexMickh/proj-user/internal/service"
	"github.com/AlexMickh/proj-user/internal/storage/minio"
	"github.com/AlexMickh/proj-user/internal/storage/postgres"
	"github.com/AlexMickh/proj-user/internal/storage/redis"
	"github.com/AlexMickh/proj-user/pkg/logger"
	"github.com/AlexMickh/proj-user/pkg/minio_client"
	"github.com/AlexMickh/proj-user/pkg/postgres_client"
	"github.com/AlexMickh/proj-user/pkg/redis_client"
	"github.com/jackc/pgx/v5/pgxpool"
	minio_lib "github.com/minio/minio-go/v7"
	redis_lib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	cfg    *config.Config
	db     *pgxpool.Pool
	s3     *minio_lib.Client
	cash   *redis_lib.Client
	server *grpc.Server
}

func Register(ctx context.Context, cfg *config.Config) *App {
	const op = "app.Register"

	log := logger.FromCtx(ctx).With(zap.String("op", op))

	log.Info("initing postgres")
	db, err := postgres_client.New(
		ctx,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.MinPools,
		cfg.DB.MaxPools,
		cfg.DB.MigrationsPath,
	)
	if err != nil {
		log.Fatal("failed to init postgres", zap.Error(err))
	}

	postgres := postgres.New(db)

	log.Info("initing minio")
	s3, err := minio_client.New(
		ctx,
		cfg.Minio.Endpoint,
		cfg.Minio.User,
		cfg.Minio.Password,
		cfg.Minio.BucketName,
		cfg.Minio.IsUseSsl,
	)
	if err != nil {
		log.Fatal("failed to init minio", zap.Error(err))
	}

	minio := minio.New(s3, cfg.Minio.BucketName)

	log.Info("initing redis")
	cash, err := redis_client.New(
		ctx,
		fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		cfg.Redis.User,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatal("failed to init redis", zap.Error(err))
	}

	redis := redis.New(cash, cfg.Redis.Expiration)

	log.Info("initing service")
	service := service.New(postgres, minio, redis)

	srv := server.New(service)
	server := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor(ctx)))
	user.RegisterUserServer(server, srv)

	return &App{
		cfg:    cfg,
		db:     db,
		s3:     s3,
		cash:   cash,
		server: server,
	}
}

func (a *App) Run(ctx context.Context) {
	const op = "app.Run"

	log := logger.FromCtx(ctx).With(zap.String("op", op))

	log.Info("starting server")

	lis, err := net.Listen("tcp", a.cfg.Server.Addr)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	go func() {
		if err := a.server.Serve(lis); err != nil {
			log.Fatal("failed to serve", zap.Error(err))
		}
	}()

	log.Info("server started", zap.String("addr", a.cfg.Server.Addr))
}

func (a *App) GracefulStop() {
	a.db.Close()
	a.s3.CredContext().Client.CloseIdleConnections()
	a.cash.Close()
	a.server.GracefulStop()
}

package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type key string

var (
	Key       = key("logger")
	RequestID = "request_id"
)

func New(ctx context.Context, outputPaths []string, env string) context.Context {
	var cfg zap.Config

	switch env {
	case "local":
		cfg = zap.Config{
			Encoding:         "console",
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			OutputPaths:      outputPaths,
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey: "msg",
				LevelKey:   "level",
				TimeKey:    "ts",
				EncodeTime: zapcore.ISO8601TimeEncoder,
			},
		}
	case "dev":
		cfg = zap.Config{
			Encoding:         "json",
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			OutputPaths:      outputPaths,
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig:    zap.NewProductionEncoderConfig(),
		}
	case "prod":
		cfg = zap.Config{
			Encoding:         "json",
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			OutputPaths:      outputPaths,
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig:    zap.NewProductionEncoderConfig(),
		}
	default:
		cfg = zap.Config{
			Encoding:         "json",
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			OutputPaths:      outputPaths,
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig:    zap.NewProductionEncoderConfig(),
		}
	}

	log, err := cfg.Build()
	if err != nil {
		panic("can't init logger: " + err.Error())
	}

	return context.WithValue(ctx, Key, log)
}

func FromCtx(ctx context.Context) *zap.Logger {
	return ctx.Value(Key).(*zap.Logger)
}

func Interceptor(ctx context.Context) grpc.UnaryServerInterceptor {
	return func(lCtx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log := FromCtx(ctx)
		lCtx = context.WithValue(lCtx, Key, log)

		md, ok := metadata.FromIncomingContext(lCtx)
		if ok {
			guid, ok := md[RequestID]
			if ok {
				FromCtx(lCtx).Error("No request id")
				ctx = context.WithValue(ctx, RequestID, guid)
			}
		}

		FromCtx(lCtx).Info("request",
			zap.String("method", info.FullMethod),
			zap.Time("request time", time.Now()),
		)

		return handler(lCtx, req)
	}
}

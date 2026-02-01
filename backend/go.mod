module github.com/samuelayo/opensms

go 1.22

require (
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/gofiber/contrib/otelfiber v1.0.10
	github.com/jackc/pgx/v5 v5.5.1
	github.com/redis/go-redis/v9 v9.4.0
	github.com/nats-io/nats.go v1.31.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.5.0
	go.uber.org/zap v1.26.0
	github.com/spf13/viper v1.18.2
	golang.org/x/crypto v0.18.0
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	github.com/prometheus/client_golang v1.18.0
	github.com/minio/minio-go/v7 v7.0.66
	github.com/pressly/goose/v3 v3.17.0
)

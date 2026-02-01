package observability

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP Metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opensms_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "opensms_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Database Metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opensms_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "opensms_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opensms_db_connections_active",
			Help: "Number of active database connections",
		},
	)

	// Cache Metrics
	cacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opensms_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	cacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opensms_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	// Business Metrics
	activeUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "opensms_active_users",
			Help: "Number of active users",
		},
		[]string{"role"},
	)

	studentsEnrolled = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "opensms_students_enrolled",
			Help: "Number of enrolled students",
		},
		[]string{"tenant_id", "grade_level"},
	)

	// Event Bus Metrics
	eventsPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opensms_events_published_total",
			Help: "Total number of events published",
		},
		[]string{"subject"},
	)

	eventsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opensms_events_processed_total",
			Help: "Total number of events processed",
		},
		[]string{"subject", "status"},
	)
)

// InitMetrics initializes Prometheus metrics
func InitMetrics() {
	// Metrics are initialized via promauto
}

// MetricsHandler returns a Fiber handler for Prometheus metrics
func MetricsHandler(c *fiber.Ctx) error {
	return adaptor.HTTPHandler(promhttp.Handler())(c)
}

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, path, status string, duration float64) {
	httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordDBQuery records database query metrics
func RecordDBQuery(operation, table string, duration float64) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordCacheHit records a cache hit
func RecordCacheHit() {
	cacheHitsTotal.Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss() {
	cacheMissesTotal.Inc()
}

// SetActiveUsers sets the number of active users
func SetActiveUsers(role string, count float64) {
	activeUsers.WithLabelValues(role).Set(count)
}

// SetStudentsEnrolled sets the number of enrolled students
func SetStudentsEnrolled(tenantID, gradeLevel string, count float64) {
	studentsEnrolled.WithLabelValues(tenantID, gradeLevel).Set(count)
}

// RecordEventPublished records an event publication
func RecordEventPublished(subject string) {
	eventsPublished.WithLabelValues(subject).Inc()
}

// RecordEventProcessed records an event processing result
func RecordEventProcessed(subject, status string) {
	eventsProcessed.WithLabelValues(subject, status).Inc()
}

package monitoring

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Metrics содержит все метрики приложения
type Metrics struct {
	// HTTP метрики
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// Бизнес метрики
	userRegistrationsTotal *prometheus.CounterVec
	userLoginsTotal        *prometheus.CounterVec
	userLoginsFailed       *prometheus.CounterVec
	passwordResetsTotal    *prometheus.CounterVec

	// Системные метрики
	activeSessions      *prometheus.GaugeVec
	databaseConnections *prometheus.GaugeVec
	redisConnections    *prometheus.GaugeVec

	logger *zap.Logger
}

// NewMetrics создает новые метрики
func NewMetrics(logger *zap.Logger) *Metrics {
	m := &Metrics{
		logger: logger,
	}

	// HTTP метрики
	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	m.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	m.httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"method", "endpoint"},
	)

	// Бизнес метрики
	m.userRegistrationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations",
		},
		[]string{"status"},
	)

	m.userLoginsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total number of user logins",
		},
		[]string{"status"},
	)

	m.userLoginsFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_logins_failed",
			Help: "Total number of failed user logins",
		},
		[]string{"reason"},
	)

	m.passwordResetsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "password_resets_total",
			Help: "Total number of password reset requests",
		},
		[]string{"status"},
	)

	// Системные метрики
	m.activeSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_sessions",
			Help: "Number of active user sessions",
		},
		[]string{},
	)

	m.databaseConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Number of active database connections",
		},
		[]string{"status"},
	)

	m.redisConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_connections",
			Help: "Number of active Redis connections",
		},
		[]string{"status"},
	)

	// Регистрируем метрики
	prometheus.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestsInFlight,
		m.userRegistrationsTotal,
		m.userLoginsTotal,
		m.userLoginsFailed,
		m.passwordResetsTotal,
		m.activeSessions,
		m.databaseConnections,
		m.redisConnections,
	)

	return m
}

// HTTPMiddleware возвращает middleware для сбора HTTP метрик
func (m *Metrics) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Увеличиваем счетчик активных запросов
			m.httpRequestsInFlight.WithLabelValues(r.Method, r.URL.Path).Inc()
			defer m.httpRequestsInFlight.WithLabelValues(r.Method, r.URL.Path).Dec()

			// Создаем ResponseWriter для перехвата статуса
			rw := &responseWriter{ResponseWriter: w, statusCode: 200}
			next.ServeHTTP(rw, r)

			// Записываем метрики
			duration := time.Since(start).Seconds()
			status := http.StatusText(rw.statusCode)

			m.httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
			m.httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)

			m.logger.Debug("HTTP request metrics recorded",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.statusCode),
				zap.Float64("duration", duration),
			)
		})
	}
}

// RecordUserRegistration записывает метрику регистрации пользователя
func (m *Metrics) RecordUserRegistration(success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	m.userRegistrationsTotal.WithLabelValues(status).Inc()
}

// RecordUserLogin записывает метрику входа пользователя
func (m *Metrics) RecordUserLogin(success bool, reason string) {
	status := "success"
	if !success {
		status = "failed"
		m.userLoginsFailed.WithLabelValues(reason).Inc()
	}
	m.userLoginsTotal.WithLabelValues(status).Inc()
}

// RecordPasswordReset записывает метрику сброса пароля
func (m *Metrics) RecordPasswordReset(success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	m.passwordResetsTotal.WithLabelValues(status).Inc()
}

// SetActiveSessions устанавливает количество активных сессий
func (m *Metrics) SetActiveSessions(count int) {
	m.activeSessions.WithLabelValues().Set(float64(count))
}

// SetDatabaseConnections устанавливает количество подключений к БД
func (m *Metrics) SetDatabaseConnections(active, idle int) {
	m.databaseConnections.WithLabelValues("active").Set(float64(active))
	m.databaseConnections.WithLabelValues("idle").Set(float64(idle))
}

// SetRedisConnections устанавливает количество подключений к Redis
func (m *Metrics) SetRedisConnections(active int) {
	m.redisConnections.WithLabelValues("active").Set(float64(active))
}

// Handler возвращает HTTP handler для метрик
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// responseWriter перехватывает статус код ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

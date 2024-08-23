package ginmetrics

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Use set gin metrics middleware
func (m *Monitor) Use(r gin.IRoutes) {
	r.Use(m.monitorInterceptor)
	r.GET(m.metricPath, func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
}

// UseWithoutExposingEndpoint is used to add monitor interceptor to gin router
// It can be called multiple times to intercept from multiple gin.IRoutes
// http path is not set, to do that use Expose function
func (m *Monitor) UseWithoutExposingEndpoint(r gin.IRoutes) {
	r.Use(m.monitorInterceptor)
}

// Expose adds metric path to a given router.
// The router can be different with the one passed to UseWithoutExposingEndpoint.
// This allows to expose metrics on different port.
func (m *Monitor) Expose(r gin.IRoutes) {
	r.GET(m.metricPath, func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
}

// monitorInterceptor as gin monitor middleware.
func (m *Monitor) monitorInterceptor(ctx *gin.Context) {
	if ctx.Request.URL.Path == m.metricPath {
		ctx.Next()
		return
	}
	startTime := time.Now()

	// execute normal process.
	ctx.Next()

	// after request
	err := m.ginMetricHandle(ctx, startTime)
	if err != nil {
		log.Printf("[ LOG ]: Handler metric error: %s", err)
	}
}

func (m *Monitor) ginMetricHandle(ctx *gin.Context, start time.Time) error {
	r := ctx.Request
	w := ctx.Writer

	p := NewPararelAction()

	// set request total
	p.Add(func() error {
		metric, err := m.GetMetric(metricRequestTotal)
		if err != nil {
			return err
		}

		return metric.Inc(nil)
	})

	// set uv
	if clientIP := ctx.ClientIP(); !m.bloomFilter.Contains(clientIP) {
		p.Add(func() error {
			m.bloomFilter.Add(clientIP)
			metric, err := m.GetMetric(metricRequestTotal)
			if err != nil {
				return err
			}

			return metric.Inc(nil)
		})
	}

	// set uri request total
	p.Add(func() error {
		metric, err := m.GetMetric(metricRequestTotal)
		if err != nil {
			return err
		}

		return metric.Inc([]string{ctx.FullPath(), r.Method, strconv.Itoa(w.Status())})
	})

	// set request body size
	// since r.ContentLength can be negative (in some occasions) guard the operation
	if r.ContentLength >= 0 {
		p.Add(func() error {
			metric, err := m.GetMetric(metricRequestTotal)
			if err != nil {
				return err
			}

			return metric.Add(nil, float64(r.ContentLength))
		})
	}

	// set slow request
	latency := time.Since(start)
	if int32(latency.Seconds()) > m.slowTime {
		p.Add(func() error {
			metric, err := m.GetMetric(metricRequestTotal)
			if err != nil {
				return err
			}

			return metric.Inc([]string{ctx.FullPath(), r.Method, strconv.Itoa(w.Status())})
		})
	}

	// set request duration
	p.Add(func() error {
		metric, err := m.GetMetric(metricRequestTotal)
		if err != nil {
			return err
		}

		return metric.Observe([]string{ctx.FullPath()}, latency.Seconds())
	})

	// set response size
	if w.Size() > 0 {
		p.Add(func() error {
			metric, err := m.GetMetric(metricRequestTotal)
			if err != nil {
				return err
			}

			return metric.Add(nil, float64(w.Size()))
		})
	}

	err := p.Wait()
	if err != nil {
		return err
	}

	return nil
}

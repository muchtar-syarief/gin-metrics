package ginmetrics

import (
	"fmt"

	"github.com/muchtar-syarief/gin-metrics/bloom"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

var ErrInvalidMetricName = errors.Errorf("invalid metric name")
var ErrInvalidMetricType = errors.Errorf("invalid metric type")
var ErrMetricNotFound = errors.New("metric not found")

// Monitor is an object that uses to set gin server monitor.
type Monitor struct {
	slowTime    int32
	metricPath  string
	reqDuration []float64
	metrics     map[MetricName]*Metric
	bloomFilter *bloom.BloomFilter
}

// Use this function to create Monitor
func NewDefaultMonitor() *Monitor {
	return &Monitor{
		metricPath:  defaultMetricPath,
		reqDuration: defaultDuration,
		slowTime:    defaultSlowTime,
		bloomFilter: bloom.NewBloomFilter(),
		metrics:     make(map[MetricName]*Metric),
	}
}

// Use this function for customing configuration Monitor
func NewMonitor(path string, duration []float64, slowtime int32) *Monitor {
	return &Monitor{
		metricPath:  path,
		reqDuration: duration,
		slowTime:    slowtime,
		bloomFilter: bloom.NewBloomFilter(),
		metrics:     make(map[MetricName]*Metric),
	}
}

// initGinMetrics used to init gin metrics
func (m *Monitor) initGinMetrics() error {
	var err error

	err = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricRequestTotal,
		Description: "all the server received request num.",
		Labels:      nil,
	})
	if err != nil {
		return err
	}

	err = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricRequestUVTotal,
		Description: "all the server received ip num.",
		Labels:      nil,
	})
	if err != nil {
		return err
	}

	err = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricURIRequestTotal,
		Description: "all the server received request num with every uri.",
		Labels:      []string{"uri", "method", "code"},
	})
	if err != nil {
		return err
	}

	err = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricRequestBody,
		Description: "the server received request body size, unit byte",
		Labels:      nil,
	})
	if err != nil {
		return err
	}

	err = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricResponseBody,
		Description: "the server send response body size, unit byte",
		Labels:      nil,
	})
	if err != nil {
		return err
	}

	err = m.AddMetric(&Metric{
		Type:        Histogram,
		Name:        metricRequestDuration,
		Description: "the time server took to handle the request.",
		Labels:      []string{"uri"},
		Buckets:     m.reqDuration,
	})
	if err != nil {
		return err
	}

	err = m.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricSlowRequest,
		Description: fmt.Sprintf("the server handled slow requests counter, t=%d.", m.slowTime),
		Labels:      []string{"uri", "method", "code"},
	})
	if err != nil {
		return err
	}

	return err
}

func (m *Monitor) RegisterDefaultMetrics() error {
	return m.initGinMetrics()
}

func (m *Monitor) SetPrefix(prefix string) {
	for _, metric := range m.metrics {
		metric.Name = MetricName(prefix) + metric.Name
	}
}

func (m *Monitor) SetSuffix(suffix string) {
	for _, metric := range m.metrics {
		metric.Name += MetricName(suffix)
	}
}

// GetMetric used to get metric object by metric_name.
func (m *Monitor) GetMetric(name MetricName) (*Metric, error) {
	metric, ok := m.metrics[name]
	if ok {
		return metric, nil
	}

	return nil, ErrMetricNotFound
}

// SetMetricPath set metricPath property. metricPath is used for Prometheus
// to get gin server monitoring data.
func (m *Monitor) SetMetricPath(path string) *Monitor {
	m.metricPath = path
	return m
}

// SetSlowTime set slowTime property. slowTime is used to determine whether
// the request is slow. For "gin_slow_request_total" metric.
func (m *Monitor) SetSlowTime(slowTime int32) *Monitor {
	m.slowTime = slowTime
	return m
}

// SetDuration set reqDuration property. reqDuration is used to ginRequestDuration
// metric buckets.
func (m *Monitor) SetDuration(duration []float64) *Monitor {
	m.reqDuration = duration
	return m
}

// AddMetric add custom monitor metric.
func (m *Monitor) AddMetric(metric *Metric) error {
	var err error

	if metric.Name == "" {
		return ErrInvalidMetricName
	}

	_, ok := m.metrics[metric.Name]
	if ok {
		err = errors.Errorf("metric '%s' is existed", metric.Name)
		return err
	}

	f, ok := PromtTypeHandler[metric.Type]
	if !ok {
		return ErrInvalidMetricType
	}

	err = f(metric)
	if err != nil {
		return err
	}

	prometheus.MustRegister(metric.vec)
	m.metrics[metric.Name] = metric
	return err

}

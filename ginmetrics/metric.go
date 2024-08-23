package ginmetrics

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

var ErrInvalidMetricVec = errors.New("invalid metric vector")

type MetricType int

const (
	None MetricType = iota
	Counter
	Gauge
	Histogram
	Summary
)

type MetricName string

var (
	metricRequestTotal    MetricName = "gin_request_total"
	metricRequestUVTotal  MetricName = "gin_request_uv_total"
	metricURIRequestTotal MetricName = "gin_uri_request_total"
	metricRequestBody     MetricName = "gin_request_body_total"
	metricResponseBody    MetricName = "gin_response_body_total"
	metricRequestDuration MetricName = "gin_request_duration"
	metricSlowRequest     MetricName = "gin_slow_request_total"
)

// Metric defines a metric object. Users can use it to save
// metric data. Every metric should be globally unique by name.
type Metric struct {
	Type        MetricType
	Name        MetricName
	Description string
	Labels      []string
	Buckets     []float64
	Objectives  map[float64]float64

	vec prometheus.Collector
}

// SetGaugeValue set data for Gauge type Metric.
func (m *Metric) SetGaugeValue(labelValues []string, value float64) error {
	var err error

	if m.Type == None {
		err = errors.Errorf("metric '%s' not existed.", m.Name)
		return err
	}

	if m.Type != Gauge {
		err = errors.Errorf("metric '%s' not Gauge type", m.Name)
		return err
	}

	vec, ok := m.vec.(*prometheus.GaugeVec)
	if !ok {
		return ErrInvalidMetricVec
	}

	vec.WithLabelValues(labelValues...).Set(value)

	return err
}

// Inc increases value for Counter/Gauge type metric, increments
// the counter by 1
func (m *Metric) Inc(labelValues []string) error {
	var err error

	if m.Type == None {
		err = errors.Errorf("metric '%s' not existed.", m.Name)
		return err
	}

	if m.Type != Gauge && m.Type != Counter {
		err = errors.Errorf("metric '%s' not Gauge or Counter type", m.Name)
		return err
	}
	switch m.Type {
	case Counter:
		vec, ok := m.vec.(*prometheus.CounterVec)
		if !ok {
			return ErrInvalidMetricVec
		}

		vec.WithLabelValues(labelValues...).Inc()

	case Gauge:
		vec, ok := m.vec.(*prometheus.GaugeVec)
		if !ok {
			return ErrInvalidMetricVec
		}

		vec.WithLabelValues(labelValues...).Inc()
	}

	return err
}

// Add adds the given value to the Metric object. Only
// for Counter/Gauge type metric.
func (m *Metric) Add(labelValues []string, value float64) error {
	var err error

	if m.Type == None {
		err = errors.Errorf("metric '%s' not existed.", m.Name)
		return err
	}

	if m.Type != Gauge && m.Type != Counter {
		err = errors.Errorf("metric '%s' not Gauge or Counter type", m.Name)
		return err
	}
	switch m.Type {
	case Counter:
		vec, ok := m.vec.(*prometheus.CounterVec)
		if !ok {
			return ErrInvalidMetricVec
		}

		vec.WithLabelValues(labelValues...).Add(value)
	case Gauge:
		vec, ok := m.vec.(*prometheus.GaugeVec)
		if !ok {
			return ErrInvalidMetricVec
		}

		vec.WithLabelValues(labelValues...).Add(value)
	}

	return err
}

// Observe is used by Histogram and Summary type metric to
// add observations.
func (m *Metric) Observe(labelValues []string, value float64) error {
	var err error

	if m.Type == 0 {
		err = errors.Errorf("metric '%s' not existed.", m.Name)
		return err
	}
	if m.Type != Histogram && m.Type != Summary {
		err = errors.Errorf("metric '%s' not Histogram or Summary type", m.Name)
		return err
	}
	switch m.Type {
	case Histogram:
		vec, ok := m.vec.(*prometheus.HistogramVec)
		if !ok {
			return ErrInvalidMetricVec
		}

		vec.WithLabelValues(labelValues...).Observe(value)

	case Summary:
		vec, ok := m.vec.(*prometheus.SummaryVec)
		if !ok {
			return ErrInvalidMetricVec
		}

		vec.WithLabelValues(labelValues...).Observe(value)
	}

	return err
}

# gin-metrics

gin-gonic/gin metrics exporter for Prometheus.

## Special Thanks
This is project based from [here]("https://github.com/penglongli/gin-metrics")

## Introduction

`gin-metrics` 为 gin HTTP 服务定义了一些监控指标，开箱即用。下边是默认监控指标的详细描述：

| Metric                  | Type      | Description                    |
| ----------------------- | --------- | ------------------------------ |
| gin_request_total       | Counter   | 服务接收到的请求总数           |
| gin_request_uv          | Counter   | 服务接收到的 IP 总数           |
| gin_uri_request_total   | Counter   | 每个 URI 接收到的服务请求数    |
| gin_request_body_total  | Counter   | 服务接收到的请求量，单位: 字节 |
| gin_response_body_total | Counter   | 服务返回的请求量，单位: 字节   |
| gin_request_duration    | Histogram | 服务处理请求使用的时间         |
| gin_slow_request_total  | Counter   | 服务接收到的慢请求计数         |

## Installation

```bash
$ go get github.com/muchtar-syarief/gin-metrics
```

## Usage

使用如下代码运行，访问：`http://localhost:8080/metrics` 即可看到暴露出来的监控指标

```go
package main

import (
	"github.com/gin-gonic/gin"

	"github.com/muchtar-syarief/gin-metrics/ginmetrics"
)

func main() {
	r := gin.Default()

	// get global Monitor object
	m := ginmetrics.NewDefaultMonitor()

	// use this function to register metric has define by repo
	err := m.RegisterDefaultMetrics()
	if err != nil {
		panic(err)
	}

	// if you want change the configuration you can use this function
	m.SetMetricPath("/metrics"). // +optional set metric path, default /debug/metrics
		SetSlowTime(10). // +optional set slow time, default 5s
		SetDuration([]float64{0.1, 0.3, 1.2, 5, 10}) // +optional set request duration, default {0.1, 0.3, 1.2, 5, 10} used to p95, p99
	
	m.Use(r)

	r.GET("/product/:id", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("id"),
		})
	})
	_ = r.Run()
}

```

## Custom Metric

`gin-metric` 提供了自定义监控指标的使用方式

### Gauge

使用 `Gauge` 类型监控指标，可以通过 3 种方法来修改监控值：`SetGaugeValue`、`Inc`、`Add`

首先，需要定义一个 `Gauge` 类型的监控指标：

```go
gaugeMetric := &ginmetrics.Metric{
    Type:        ginmetrics.Gauge,
    Name:        "example_gauge_metric",
    Description: "an example of gauge type metric",
    Labels:      []string{"label1"},
}

// Add metric to monitor object
_ = m.AddMetric(gaugeMetric)
```

**SetGaugeValue**

`SetGaugeValue` 方法会直接设置监控指标的值

```go
metric, err := m.GetMetric("example_gauge_metric")
if err != nil {
	panic(err)
}

err =  metric.SetGaugeValue([]string{"label_value1"}, 0.1)
if err != nil {
	panic(err)
}
```

**Inc**

`Inc` 方法会在监控指标值的基础上增加 1

```go
metric, err := m.GetMetric("example_gauge_metric")
if err != nil {
	panic(err)
}

err = metric.Inc([]string{"label_value1"})
if err != nil {
	panic(err)
}
```

**Add**

`Add` 方法会为监控指标增加传入的值

```go
metric, err := m.GetMetric("example_gauge_metric")
if err != nil {
	panic(err)
}

err = metric.Add([]string{"label_value1"}, 0.2)
if err != nil {
	panic(err)
}
```

### Counter

`Counter` 类型的监控指标，可以使用 `Inc` 和 `Add` 方法，但是不能使用 `SetGaugeValue` 方法

### Histogram and Summary

对于 `Histogram` 和 `Summary` 类型的监控指标，需要用 `Observe` 方法来设置监控值。

## Contributing

如果有遇见什么问题，或者需要修改，可以 [新建 ISSUE](https://github.com/muchtar-syarief/gin-metrics/issues/new)
或者 [新建 PullRequest](https://github.com/muchtar-syarief/gin-metrics/pulls).

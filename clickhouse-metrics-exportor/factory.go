package clickhousemetrics

import (
	"context"
	"fmt"

	"github.com/zeelrupapara/clickhousemetrics/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)


func NewFactory() exporter.Factory {
  return exporter.NewFactory(
	metadata.Type,
    createDefaultConfig,
    exporter.WithMetrics(createMetricsExporter, component.StabilityLevelDevelopment),
  )
}


func createMetricsExporter(
  ctx context.Context,
  set exporter.Settings,
  config component.Config,
) (exporter.Metrics, error) {

  cfg := config.(*Config)
  exporter, err := newMetricsExporter(set.Logger, cfg)
  if err != nil {
	return nil, fmt.Errorf("connot configure metrics exporter: %w", err)
  }
  return exporterhelper.NewMetrics(
	ctx,
	set,
	config,
	exporter.pushMetricsData,
	exporterhelper.WithStart(exporter.start),
	exporterhelper.WithShutdown(exporter.shutdown),
  )
}

func createDefaultConfig() component.Config {
  return &Config{
	Endpoint: defaultEndpoint,
	Username: defaultUsername,
	Password: defaultPassword,
  }
}

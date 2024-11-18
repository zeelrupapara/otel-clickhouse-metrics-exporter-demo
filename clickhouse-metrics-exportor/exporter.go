package clickhousemetrics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type clickhouseMetricsExporter struct {
	client *sql.DB
	logger *zap.Logger
	cfg    *Config
}

type Model struct {
	metricsname string
	metricsunit string
	sum         pmetric.Sum
}

type Metrics struct {
	count     int
	insertSQL string
	models    []*Model
}

func (s *Metrics) Add(metrics any, name string, unit string) error {
	sum, ok := metrics.(pmetric.Sum)
	if !ok {
		return fmt.Errorf("metrics param is not type of Sum")
	}
	s.count += sum.DataPoints().Len()
	s.models = append(s.models, &Model{
		metricsname: name,
		metricsunit: unit,
		sum:         sum,
	})
	return nil
}

func createDatabase(ctx context.Context, cfg *Config) error {
	// use default database to create new database
	db, err := cfg.buildDB()
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", defaultDatabase)
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}
	return nil
}

func createMatrixTable(ctx context.Context, cfg *Config, db *sql.DB) error {
	const table = "metrics"
	const metricsTable = `
  CREATE TABLE IF NOT EXISTS %s (
    MetricName String,
    MetricUnit String CODEC(ZSTD(1)),
    StartTimeUnix DateTime64(9) CODEC(Delta, ZSTD(1)),
    TimeUnix DateTime64(9) CODEC(Delta, ZSTD(1)),
    Value Float64 CODEC(ZSTD(1))
  ) 
  ENGINE = MergeTree
  PARTITION BY toDate(TimeUnix)
  ORDER BY (MetricName, TimeUnix);
  `

	// use default database to create new database
	query := fmt.Sprintf(metricsTable, table)
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	return nil
}

func newMetricsExporter(logger *zap.Logger, cfg *Config) (*clickhouseMetricsExporter, error) {
	client, err := cfg.buildDB()
	if err != nil {
		return nil, err
	}
	return &clickhouseMetricsExporter{
		client: client,
		logger: logger,
		cfg:    cfg,
	}, nil
}

func (e *clickhouseMetricsExporter) start(ctx context.Context, _ component.Host) error {
	// internal.SetLogger(e.logger)

	// Create schema
	if err := createDatabase(ctx, e.cfg); err != nil {
		return err
	}

	// Create Metrix Tables
	if err := createMatrixTable(ctx, e.cfg, e.client); err != nil {
		return err
	}
	return nil
}

func (e *clickhouseMetricsExporter) shutdown(_ context.Context) error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

func attributesToMap(attributes pcommon.Map) map[string]string {
	m := make(map[string]string, attributes.Len())
	attributes.Range(func(k string, v pcommon.Value) bool {
		m[k] = v.AsString()
		return true
	})
	return m
}

func getValue(intValue int64, floatValue float64, dataType any) float64 {
	switch t := dataType.(type) {
	case pmetric.ExemplarValueType:
		switch t {
		case pmetric.ExemplarValueTypeDouble:
			return floatValue
		case pmetric.ExemplarValueTypeInt:
			return float64(intValue)
		case pmetric.ExemplarValueTypeEmpty:
			return 0.0
		default:
			return 0.0
		}
	case pmetric.NumberDataPointValueType:
		switch t {
		case pmetric.NumberDataPointValueTypeDouble:
			return floatValue
		case pmetric.NumberDataPointValueTypeInt:
			return float64(intValue)
		case pmetric.NumberDataPointValueTypeEmpty:
			return 0.0
		default:
			return 0.0
		}
	default:
		return 0.0
	}
}

func (s *Metrics) Insert(ctx context.Context, db *sql.DB) error {
	if s.count == 0 {
		return nil
	}
	const insertQuery = "INSERT INTO metrics (MetricName, MetricUnit, StartTimeUnix, TimeUnix, Value) VALUES(?, ?, ?, ?, ?)"
	start := time.Now()
	err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		statement, err := tx.PrepareContext(ctx, insertQuery)
		if err != nil {
			return err
		}
		defer func() {
			_ = statement.Close()
		}()
		for _, model := range s.models {
			for i := 0; i < model.sum.DataPoints().Len(); i++ {
				dp := model.sum.DataPoints().At(i)
				_, err = statement.ExecContext(ctx,
					model.metricsname,
					model.metricsunit,
					dp.StartTimestamp().AsTime(),
					dp.Timestamp().AsTime(),
					getValue(dp.IntValue(), dp.DoubleValue(), dp.ValueType()),
				)
				if err != nil {
					return fmt.Errorf("ExecContext:%w", err)
				}
			}
		}
		return err
	})

	duration := time.Since(start)
	if err != nil {
		return fmt.Errorf("insert sum metrics fail:%w", err)
	}

	// TODO latency metrics
	zap.Duration("cost", duration)
	return nil
}

func doWithTx(_ context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (e *clickhouseMetricsExporter) pushMetricsData(ctx context.Context, md pmetric.Metrics) error {
	m := &Metrics{}
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		metrics := md.ResourceMetrics().At(i)
		for j := 0; j < metrics.ScopeMetrics().Len(); j++ {
			rs := metrics.ScopeMetrics().At(j).Metrics()
			fmt.Println(rs)
			for k := 0; k < rs.Len(); k++ {
				r := rs.At(k)
				var errs error
				//exhaustive:enforce
				switch r.Type() {
				case pmetric.MetricTypeSum:
					errs = errors.Join(errs, m.Add(r.Sum(), r.Name(), r.Unit()))
				case pmetric.MetricTypeEmpty:
					return fmt.Errorf("metrics type is unset")
				default:
					return fmt.Errorf("unsupported metrics type")
				}
				if errs != nil {
					return errs
				}
			}
		}
	}
	return m.Insert(ctx, e.client)
}

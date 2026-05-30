package collector

import (
	"context"
	"fmt"
	"strconv"

	goredis "github.com/redis/go-redis/v9"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
)

// RedisCollector monitors a Redis instance.
type RedisCollector struct {
	tenantID    string
	componentID string
	client      *goredis.Client
}

// NewRedisCollector creates a new Redis health collector.
func NewRedisCollector(tenantID, componentID string, client *goredis.Client) *RedisCollector {
	return &RedisCollector{
		tenantID:    tenantID,
		componentID: componentID,
		client:      client,
	}
}

func (c *RedisCollector) Type() domain.CollectorType { return domain.CollectorRedis }
func (c *RedisCollector) ComponentID() string        { return c.componentID }
func (c *RedisCollector) TenantID() string           { return c.tenantID }

func (c *RedisCollector) Collect(ctx context.Context) (*port.CollectorResult, error) {
	result := &port.CollectorResult{
		ComponentID:   c.componentID,
		CollectorType: domain.CollectorRedis,
		Metrics:       make([]port.MetricReading, 0, 5),
	}

	// 1. Availability (ping)
	err := c.client.Ping(ctx).Err()
	if err != nil {
		result.Metrics = append(result.Metrics,
			port.MetricReading{MetricType: domain.MetricAvailability, Value: 0, Labels: map[string]string{"error": err.Error()}},
		)
		result.Error = err
		return result, nil
	}

	result.Metrics = append(result.Metrics,
		port.MetricReading{MetricType: domain.MetricAvailability, Value: 1, Labels: map[string]string{}},
	)

	// Get INFO stats
	info, err := c.client.Info(ctx, "memory", "stats", "clients").Result()
	if err != nil {
		return result, nil
	}

	stats := parseRedisInfo(info)

	// 2. Memory usage (percentage of maxmemory, or absolute bytes)
	usedMemory := parseFloat(stats["used_memory"])
	maxMemory := parseFloat(stats["maxmemory"])
	memoryPct := 0.0
	if maxMemory > 0 {
		memoryPct = (usedMemory / maxMemory) * 100
	} else {
		// If no maxmemory set, report in MB
		memoryPct = usedMemory / (1024 * 1024)
	}
	result.Metrics = append(result.Metrics,
		port.MetricReading{
			MetricType: domain.MetricMemoryUsage,
			Value:      memoryPct,
			Labels: map[string]string{
				"used_bytes": fmt.Sprintf("%.0f", usedMemory),
				"max_bytes":  fmt.Sprintf("%.0f", maxMemory),
				"unit":       "percent",
			},
		},
	)

	// 3. Hit ratio
	hits := parseFloat(stats["keyspace_hits"])
	misses := parseFloat(stats["keyspace_misses"])
	hitRatio := 100.0
	if hits+misses > 0 {
		hitRatio = (hits / (hits + misses)) * 100
	}
	result.Metrics = append(result.Metrics,
		port.MetricReading{
			MetricType: domain.MetricHitRatio,
			Value:      hitRatio,
			Labels:     map[string]string{"hits": fmt.Sprintf("%.0f", hits), "misses": fmt.Sprintf("%.0f", misses), "unit": "percent"},
		},
	)

	// 4. Evictions
	evictions := parseFloat(stats["evicted_keys"])
	result.Metrics = append(result.Metrics,
		port.MetricReading{MetricType: domain.MetricEvictions, Value: evictions, Labels: map[string]string{}},
	)

	// 5. Connected clients
	connectedClients := parseFloat(stats["connected_clients"])
	result.Metrics = append(result.Metrics,
		port.MetricReading{MetricType: domain.MetricConnectedClients, Value: connectedClients, Labels: map[string]string{}},
	)

	return result, nil
}

// parseRedisInfo parses Redis INFO output into key-value pairs.
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	var key, value string
	inValue := false
	lineStart := 0

	for i := 0; i <= len(info); i++ {
		if i == len(info) || info[i] == '\n' {
			line := info[lineStart:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 && line[0] != '#' {
				if inValue {
					result[key] = value
				}
				for j := 0; j < len(line); j++ {
					if line[j] == ':' {
						key = line[:j]
						value = line[j+1:]
						result[key] = value
						break
					}
				}
			}
			lineStart = i + 1
			inValue = false
			continue
		}
	}
	_ = inValue

	return result
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

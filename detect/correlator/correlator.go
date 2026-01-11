package correlator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/podwatch/podwatch/pkg/models"
	"github.com/redis/go-redis/v9"
)

// Correlator handles stateful event correlation using Redis
type Correlator struct {
	rdb *redis.Client
	ctx context.Context
}

func NewCorrelator(redisAddr string) *Correlator {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &Correlator{
		rdb: rdb,
		ctx: context.Background(),
	}
}

// SequenceRule defines a multi-step attack pattern
type SequenceRule struct {
	ID         string
	Name       string
	Steps      []string // CEL conditions for each step
	WindowSecs int64    // Time window for entire sequence
	GroupBy    string   // Field to group by (e.g., "container.container_id")
	Response   string
	Severity   string
}

// ThresholdRule defines rate-based detection
type ThresholdRule struct {
	ID         string
	Name       string
	Condition  string // CEL condition for matching events
	Count      int    // Number of events required
	WindowSecs int64  // Time window
	GroupBy    string // Field to group by
	Response   string
	Severity   string
}

// TrackEvent stores an event for correlation
func (c *Correlator) TrackEvent(event models.RuntimeEvent, ruleID, groupKey string, windowSecs int64) error {
	key := fmt.Sprintf("corr:%s:%s", ruleID, groupKey)

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Add to sorted set with timestamp as score
	score := float64(event.Timestamp.Unix())
	if err := c.rdb.ZAdd(c.ctx, key, redis.Z{Score: score, Member: string(data)}).Err(); err != nil {
		return err
	}

	// Set expiry
	c.rdb.Expire(c.ctx, key, time.Duration(windowSecs)*time.Second)

	// Cleanup old events outside window
	minScore := float64(time.Now().Unix() - windowSecs)
	c.rdb.ZRemRangeByScore(c.ctx, key, "-inf", fmt.Sprintf("%f", minScore))

	return nil
}

// CheckThreshold returns true if threshold is met
func (c *Correlator) CheckThreshold(ruleID, groupKey string, threshold int, windowSecs int64) (bool, []string, error) {
	key := fmt.Sprintf("corr:%s:%s", ruleID, groupKey)

	minScore := float64(time.Now().Unix() - windowSecs)
	maxScore := float64(time.Now().Unix())

	// Get events in window
	events, err := c.rdb.ZRangeByScore(c.ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", minScore),
		Max: fmt.Sprintf("%f", maxScore),
	}).Result()
	if err != nil {
		return false, nil, err
	}

	if len(events) >= threshold {
		// Extract event IDs
		var eventIDs []string
		for _, e := range events {
			var event models.RuntimeEvent
			if err := json.Unmarshal([]byte(e), &event); err == nil {
				eventIDs = append(eventIDs, event.EventID)
			}
		}
		return true, eventIDs, nil
	}

	return false, nil, nil
}

// TrackSequenceStep tracks progress through a sequence
func (c *Correlator) TrackSequenceStep(ruleID, groupKey string, step int, eventID string, windowSecs int64) error {
	key := fmt.Sprintf("seq:%s:%s", ruleID, groupKey)
	field := fmt.Sprintf("step_%d", step)

	if err := c.rdb.HSet(c.ctx, key, field, eventID).Err(); err != nil {
		return err
	}
	c.rdb.Expire(c.ctx, key, time.Duration(windowSecs)*time.Second)
	return nil
}

// CheckSequenceComplete checks if all steps are completed
func (c *Correlator) CheckSequenceComplete(ruleID, groupKey string, totalSteps int) (bool, []string, error) {
	key := fmt.Sprintf("seq:%s:%s", ruleID, groupKey)

	result, err := c.rdb.HGetAll(c.ctx, key).Result()
	if err != nil {
		return false, nil, err
	}

	if len(result) >= totalSteps {
		var eventIDs []string
		for i := 0; i < totalSteps; i++ {
			if id, ok := result[fmt.Sprintf("step_%d", i)]; ok {
				eventIDs = append(eventIDs, id)
			}
		}
		// Clear the sequence
		c.rdb.Del(c.ctx, key)
		return true, eventIDs, nil
	}

	return false, nil, nil
}

// Close closes the Redis connection
func (c *Correlator) Close() error {
	return c.rdb.Close()
}

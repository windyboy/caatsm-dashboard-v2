package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	natsClient "github.com/nats-io/nats.go"
	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/streaming"
	"go.uber.org/zap"
)

var (
	messageTypes = []string{"AFTN", "SITA", "ACARS", "CPDLC"}
	airports     = []string{"KJFK", "EGLL", "LFPG", "EDDF", "KLAX", "KORD", "RJTT", "ZBAA", "OMDB", "YSSY"}
	airlines     = []string{"AA", "BA", "LH", "AF", "UA", "DL", "JL", "CA", "EK", "QF"}
	priorities   = []int{1, 2, 3}
	contents     = []string{
		"Flight plan filed",
		"Departure clearance",
		"Arrival clearance",
		"Weather update",
		"ATC instruction",
		"Position report",
		"Emergency notification",
		"Route change",
		"Altitude change",
		"Speed adjustment",
	}
)

func generateTelegram(counter int) *domain.Telegram {
	now := time.Now()

	// Generate random flight number
	airline := airlines[rand.Intn(len(airlines))]
	flightNum := fmt.Sprintf("%s%d", airline, 100+rand.Intn(900))

	// Random source and destination
	src := airports[rand.Intn(len(airports))]
	dst := airports[rand.Intn(len(airports))]
	for dst == src {
		dst = airports[rand.Intn(len(airports))]
	}

	return &domain.Telegram{
		MessageID:    fmt.Sprintf("LIVE-%d-%d", now.Unix(), counter),
		Type:         messageTypes[rand.Intn(len(messageTypes))],
		Time:         now,
		FlightNumber: flightNum,
		Source:       src,
		Destination:  dst,
		Content:      contents[rand.Intn(len(contents))],
		Priority:     priorities[rand.Intn(len(priorities))],
		RawData:      fmt.Sprintf(`{"live": true, "counter": %d, "timestamp": %d}`, counter, now.Unix()),
	}
}

func publishTelegram(js natsClient.JetStreamContext, telegram *domain.Telegram, logger *zap.Logger) error {
	// Marshal telegram to JSON
	data, err := json.Marshal(telegram)
	if err != nil {
		return fmt.Errorf("marshal telegram: %w", err)
	}

	// Publish to NATS JetStream
	subject := fmt.Sprintf("telegrams.%s", telegram.Type)
	_, err = js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("publish to NATS: %w", err)
	}

	logger.Info("published telegram",
		zap.String("message_id", telegram.MessageID),
		zap.String("type", telegram.Type),
		zap.String("flight", telegram.FlightNumber),
		zap.String("route", fmt.Sprintf("%s -> %s", telegram.Source, telegram.Destination)),
	)

	return nil
}

func main() {
	var (
		cfgPath  = flag.String("config", "", "Path to config file (optional)")
		interval = flag.Duration("interval", 2*time.Second, "Interval between messages")
		burst    = flag.Int("burst", 1, "Number of messages per interval")
		total    = flag.Int("total", 0, "Total number of messages (0 for infinite)")
	)
	flag.Parse()

	ctx := context.Background()

	// Load config
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logger, _ := zap.NewDevelopment()
	defer func() { _ = logger.Sync() }()

	logger.Info("connecting to NATS",
		zap.String("url", cfg.NATS.URL),
		zap.String("stream", cfg.NATS.Stream),
	)

	// Connect to NATS
	conn, js, err := streaming.Connect(ctx, cfg.NATS)
	if err != nil {
		logger.Fatal("connect to NATS", zap.Error(err))
	}
	defer conn.Close()

	// Ensure stream exists
	streamCfg := &natsClient.StreamConfig{
		Name:      cfg.NATS.Stream,
		Subjects:  []string{"telegrams.>"},
		Retention: natsClient.LimitsPolicy,
		MaxAge:    24 * time.Hour * 30,
		Storage:   natsClient.FileStorage,
	}

	_, err = js.AddStream(streamCfg)
	if err != nil {
		// Stream might exist, try update
		_, err = js.UpdateStream(streamCfg)
		if err != nil {
			logger.Warn("ensure stream", zap.Error(err))
		}
	}

	logger.Info("stream ready",
		zap.String("name", cfg.NATS.Stream),
		zap.Duration("interval", *interval),
		zap.Int("burst", *burst),
		zap.Int("total", *total),
	)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	counter := 0
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	logger.Info("starting message publishing (press Ctrl+C to stop)")

	for {
		select {
		case <-sigChan:
			logger.Info("shutting down", zap.Int("total_published", counter))
			return

		case <-ticker.C:
			// Publish burst of messages
			for i := 0; i < *burst; i++ {
				counter++
				telegram := generateTelegram(counter)

				if err := publishTelegram(js, telegram, logger); err != nil {
					logger.Error("failed to publish", zap.Error(err))
					continue
				}

				// Check if we've reached the total
				if *total > 0 && counter >= *total {
					logger.Info("reached total messages", zap.Int("total", counter))
					return
				}
			}
		}
	}
}

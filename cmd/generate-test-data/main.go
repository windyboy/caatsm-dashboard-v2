package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/windy/caatsm-dashboard/config"
	"github.com/windy/caatsm-dashboard/internal/domain"
	"github.com/windy/caatsm-dashboard/internal/infrastructure/persistence"
	"go.uber.org/zap"
)

var (
	messageTypes = []string{"AFTN", "SITA", "ACARS", "CPDLC"}
	airports     = []string{"KJFK", "EGLL", "LFPG", "EDDF", "KLAX", "KORD", "RJTT", "ZBAA", "OMDB", "YSSY"}
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

func generateTestData(count int) []*domain.Telegram {
	telegrams := make([]*domain.Telegram, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		// Generate random time within last 24 hours
		hoursAgo := rand.Intn(24)
		minutesAgo := rand.Intn(60)
		timestamp := now.Add(-time.Duration(hoursAgo)*time.Hour - time.Duration(minutesAgo)*time.Minute)

		// Generate random flight number
		airline := []string{"AA", "BA", "LH", "AF", "UA", "DL", "JL", "CA", "EK", "QF"}[rand.Intn(10)]
		flightNum := fmt.Sprintf("%s%d", airline, 100+rand.Intn(900))

		// Random source and destination
		src := airports[rand.Intn(len(airports))]
		dst := airports[rand.Intn(len(airports))]
		for dst == src {
			dst = airports[rand.Intn(len(airports))]
		}

		telegrams[i] = &domain.Telegram{
			MessageID:    fmt.Sprintf("MSG%06d", i+1),
			Type:         messageTypes[rand.Intn(len(messageTypes))],
			Time:         timestamp,
			FlightNumber: flightNum,
			Source:       src,
			Destination:  dst,
			Content:      contents[rand.Intn(len(contents))],
			Priority:     priorities[rand.Intn(len(priorities))],
			RawData:      fmt.Sprintf(`{"generated": true, "index": %d}`, i),
		}
	}

	return telegrams
}

func main() {
	var (
		count   = flag.Int("count", 50, "Number of test telegrams to generate")
		cfgPath = flag.String("config", "", "Path to config file (optional)")
		clear   = flag.Bool("clear", false, "Clear existing data before inserting")
	)
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	logger, _ := zap.NewDevelopment()
	defer func() { _ = logger.Sync() }()

	// Initialize database connection
	pool, err := persistence.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		logger.Fatal("create postgres pool", zap.Error(err))
	}
	defer pool.Close()

	store := persistence.NewPostgresStore(pool)

	// Clear existing data if requested
	if *clear {
		logger.Info("clearing existing telegrams")
		_, err := pool.Exec(ctx, "DELETE FROM telegrams")
		if err != nil {
			logger.Fatal("clear telegrams", zap.Error(err))
		}
		logger.Info("cleared existing telegrams")
	}

	// Generate test data
	logger.Info("generating test data", zap.Int("count", *count))
	telegrams := generateTestData(*count)

	// Save to database
	logger.Info("saving telegrams to database")
	if err := store.BulkSave(ctx, telegrams); err != nil {
		logger.Fatal("save telegrams", zap.Error(err))
	}

	logger.Info("test data generated successfully", zap.Int("count", *count))
}

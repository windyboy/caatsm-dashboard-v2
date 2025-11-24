package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/windy/caatsm-dashboard/config"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "path to config file")
	flag.Parse()

	if cfgPath == "" {
		fmt.Fprintf(os.Stderr, "error: config file path required\n")
		os.Exit(1)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to load config: %v\n", err)
		os.Exit(1)
	}

	dsn := strings.TrimSpace(cfg.Database.DSN)
	if dsn == "" {
		fmt.Fprintf(os.Stderr, "error: database.dsn not found in config\n")
		os.Exit(1)
	}

	fmt.Print(dsn)
}


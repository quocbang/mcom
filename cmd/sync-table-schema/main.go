package main

import (
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type dbConfig struct {
	DatabaseName string `long:"db-name" description:"database name" required:"true"`
	Address      string `long:"db-address" description:"database IP Address" required:"true"`
	Port         int    `long:"db-port" description:"database port" required:"true"`
	UserName     string `long:"db-username" description:"database username" required:"true"`
	Password     string `long:"db-password" description:"database password" required:"true"`
	Schema       string `long:"db-schema" description:"database schema" required:"true"`
	SourceData   string `long:"db-source-data" description:"the schema where the data is synchronized"`
}

func main() {
	logger, err := newLogger()
	if err != nil {
		log.Fatalf("failed to create a logger: %v", err)
	}

	var config dbConfig
	if err := parseFlags(&config); err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			os.Exit(0)
		}
		logger.Fatal("failed to parse flags", zap.Error(err))
	}

	db, err := newDBSynchronizer(config, withLogger(logger))
	if err != nil {
		logger.Fatal("failed to create a db synchronizer", zap.Error(err))
	}
	defer db.Close()

	if err := db.MaybeDropDestinationSchema(); err != nil {
		exit(logger, 1)
	}

	if err := db.MaybeCreateDestinationSchema(); err != nil {
		exit(logger, 1)
	}

	if err := db.MigrateTables(); err != nil {
		exit(logger, 1)
	}

	db.MaybeSyncTablesData()
}

func newLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	return config.Build()
}

func parseFlags(config *dbConfig) error {
	parser := flags.NewParser(config, flags.Default)

	_, err := parser.Parse()
	return err
}

func exit(logger *zap.Logger, code int) {
	logger.Info("terminate program")
	os.Exit(code)
}

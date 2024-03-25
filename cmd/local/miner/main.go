package main

import (
	config "sob-miner"
	"sob-miner/internal/mempool"
	"sob-miner/internal/miner"
	"sob-miner/internal/path"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			logrus.Error(err)
		}
	}()

	logger := logrus.New()
	// logger.SetLevel(logrus.InfoLevel)
	logger.SetLevel(logrus.DebugLevel)
	logger.Formatter = &logrus.TextFormatter{
		DisableColors: false,
		ForceColors:   true,
	}

	mempoolConfig := mempool.Opts{
		MaxMemPoolSize: config.MaxMemPoolSize,
		Logger:         logger,

		Dust: uint64(config.Dust),
	}

	// init mempool
	pool, err := mempool.New(sqlite.Open(path.LocalDBPath), mempoolConfig, &gorm.Config{
		NowFunc:                func() time.Time { return time.Now().UTC() },
		Logger:                 gormLogger.Default.LogMode(gormLogger.Silent),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		panic(err)
	}

	if err := pool.ResetTables(); err != nil {
		panic(err)
	}

	logger.Info("mempool initialized")
	logger.Info("starting miner")

	miner, err := miner.New(pool, miner.Opts{
		Logger:       logger,
		MaxBlockSize: uint(config.MAX_BLOCK_SIZE),
	})

	if err != nil {
		panic(err)
	}

	if err := miner.Mine(); err != nil {
		panic(err)
	}

}

func Contains(target string, array []string) bool {
	for _, element := range array {
		if element == target {
			return true
		}
	}
	return false
}

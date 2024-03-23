package main

import (
	"fmt"
	"os"
	config "sob-miner"
	"sob-miner/internal/mempool"
	"sob-miner/internal/miner"
	"sob-miner/internal/path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type ProgressBar struct {
	Total   int
	Current int

	rate string
}

func (p *ProgressBar) Play(cur int) {
	percent := float64(cur) / float64(p.Total) * 100
	fmt.Printf("\r[%-50s]%3d%% %8d/%d", strings.Repeat("#", int(percent/2)), uint(percent), cur, p.Total)
}

func main() {

	os.Remove(path.DBPath)

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
	pool, err := mempool.New(sqlite.Open(path.DBPath), mempoolConfig, &gorm.Config{
		NowFunc:                func() time.Time { return time.Now().UTC() },
		Logger:                 gormLogger.Default.LogMode(gormLogger.Silent),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
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

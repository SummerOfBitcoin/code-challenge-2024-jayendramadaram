package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	config "sob-miner"
	"sob-miner/internal/ierrors"
	"sob-miner/internal/mempool"
	"sob-miner/internal/miner"
	"sob-miner/internal/path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// logs current progress to std out
type ProgressBar struct {
	Total   int
	Current int

	rate string
}

func (p *ProgressBar) Play(cur int) {
	percent := float64(cur) / float64(p.Total) * 100
	fmt.Printf("\r[%-50s]%3d%% %8d/%d", strings.Repeat("#", int(percent/2)), uint(percent), cur, p.Total)
}

/*
* App consists of two services
* - mempool and miner
 */
func main() {

	os.Remove(path.DBPath)

	defer func() {
		if err := recover(); err != nil {
			logrus.Error(err)
		}
	}()

	logger := logrus.New()
	// logger.SetLevel(logrus.InfoLevel)
	logger.SetLevel(logrus.PanicLevel)
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

	// loop through all files in ./data/mempool
	// unmarshal all json objects
	files, err := os.ReadDir(path.MempoolDataPath)
	if err != nil {
		panic(err)
	}

	totalFiles := len(files)
	logger.Info("loading ", totalFiles, " files")
	pb := &ProgressBar{
		Total:   totalFiles,
		Current: 0,

		rate: "#",
	}

	wg := new(sync.WaitGroup)
	doneChan := make(chan struct{})

	rejectedTxFile := "rejected.txt"
	if _, err := os.Stat(path.Root + "/" + rejectedTxFile); err == nil {
		os.Remove(path.Root + "/" + rejectedTxFile)
	}

	rejTxFile, err := os.OpenFile(path.Root+"/"+rejectedTxFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer rejTxFile.Close()

	acceptableErrs := []string{
		ierrors.ErrAsmAndScriptMismatch.Error(),
		ierrors.ErrFeeTooLow.Error(),
	}

	start := time.Now()
	// TODO: spawing 8k go routines is bad and should be improved
	// REASON: will spend more time context switching and GC rather than work
	// SOLUTION: use a channel to distribute files to set of go routines
	for i, file := range files {

		if file.IsDir() && strings.Split(file.Name(), ".")[1] != "json" {
			logger.Info("skipping ", file.Name())
			continue
		}

		wg.Add(1)
		go func(i int, file fs.DirEntry) {
			defer func() {
				doneChan <- struct{}{}
				wg.Done()
			}()

			txData, err := os.ReadFile(path.MempoolDataPath + "/" + file.Name())
			if err != nil {
				panic(err)
			}

			var tx mempool.Transaction
			if err := json.Unmarshal(txData, &tx); err != nil {
				panic(err)
			}

			if err := pool.PutTx(tx); err != nil {
				logger.Info("processing ", file.Name())
				rejTxFile.WriteString(file.Name() + " Reason: " + err.Error() + "\n")

				if Contains(err.Error(), acceptableErrs) {
					// logger.Info("press enter to continue")
					// reader := bufio.NewReader(os.Stdin)
					// _, _ = reader.ReadString('\n')
					return
				}
				panic(err)
			}
			// pb.Current++
			// pb.Play(pb.Current)

		}(i, file)
	}

	for i := 0; i < totalFiles; i++ {
		<-doneChan
		pb.Current++
		pb.Play(pb.Current)
	}

	fmt.Println("")

	wg.Wait()
	elapsed := time.Since(start)
	logger.Info("loaded ", len(files), " transactions into Mempool", " in ", elapsed.Seconds(), " seconds")

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

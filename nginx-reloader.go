package main

import (
	"errors"
	"fmt"
	"github.com/avulor/nginx-reloader/utils"
	"os"
	"time"
)

// CLI
func main() {

	pollCooldown, watchedDirs, logFile, logMaxSize, logPollCooldowns, nginxCommand, err := utils.ParseOptions(os.Args)
	if err != nil {
		utils.Fatalf("%v", err)
	}

	err = StartNginxReloader(pollCooldown, watchedDirs, logFile, logMaxSize, logPollCooldowns, nginxCommand)

	if err != nil {
		utils.Fatalf("%v", err)
	}
}

// Programmatic API
func StartNginxReloader(pollCooldown time.Duration, watchedDirs []string, logFile string, logMaxSize int64, logPollCooldowns int, nginxCommand []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	validateWatchedDirs(watchedDirs)

	watcher := utils.MakeConfLogWatcher(watchedDirs, pollCooldown, logFile, logMaxSize, logPollCooldowns)

	watcher.CalcChecksum()

	NginxRunner := utils.MakeNginxRunner(watcher.ChangeChan, watcher.LogFullChan, logFile, nginxCommand)

	cmd := NginxRunner.StartNginx()

	watcher.Watch()

	err = cmd.Wait()
	if err != nil {
		utils.Panicf("nginx process encountered an error during execution:\n%v\n", err)
	}
	return err
}

func validateWatchedDirs(watchedDirs []string) {
	for _, el := range watchedDirs {
		stat, err := os.Stat(el)
		if err != nil {
			if os.IsNotExist(err) {
				utils.Panicf("watched directory '%v' does not exist", el)
			} else {
				utils.Panicf("couldn't stat watched directory '%v'", el)
			}
		}
		if !stat.IsDir() {
			utils.Panicf("watched path '%v' is not a directory", el)
		}
	}
}

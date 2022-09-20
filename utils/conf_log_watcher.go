package utils

import (
	"errors"
	"os"
	"time"
)

type ConfLogWatcher struct {
	watchedDirs      []string
	dirChecksum      string
	ChangeChan       chan bool
	LogFullChan      chan bool
	pollCooldown     time.Duration
	logFile          string
	logMaxSize       int64
	logPollCooldowns int
}

func MakeConfLogWatcher(
	watchedDirs []string,
	pollCooldown time.Duration,
	logFile string,
	logMaxSize int64,
	logPollCooldowns int,
) ConfLogWatcher {
	watcher := ConfLogWatcher{
		watchedDirs:      watchedDirs,
		ChangeChan:       make(chan bool),
		LogFullChan:      make(chan bool),
		pollCooldown:     pollCooldown,
		logFile:          logFile,
		logMaxSize:       logMaxSize,
		logPollCooldowns: logPollCooldowns,
	}
	return watcher
}

func (w *ConfLogWatcher) Watch() {
	go func() {
		cooldowns := 0
		for {
			shouldCheckLogSize := cooldowns == 0 || cooldowns >= w.logPollCooldowns
			isLogFull := shouldCheckLogSize && w.isLogFull()
			if isLogFull {
				w.onLogFull()
			}
			time.Sleep(w.pollCooldown)
			cooldowns++
			knownChecksum := w.dirChecksum
			w.CalcChecksum()
			hasChanged := knownChecksum != w.dirChecksum
			if hasChanged {
				w.onChange()
			}
		}
	}()
}

func (w *ConfLogWatcher) CalcChecksum() {
	w.dirChecksum = ""
	for _, watchedDir := range w.watchedDirs {
		w.dirChecksum += calcDirChecksum(watchedDir) + ";"
	}
}

func (w *ConfLogWatcher) onChange() {
	w.ChangeChan <- true
}

func (w *ConfLogWatcher) isLogFull() (isFull bool) {
	if fileInfo, err := os.Stat(w.logFile); err == nil {
		size := fileInfo.Size()
		if size >= w.logMaxSize {
			return true
		} else {
			return false
		}
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		Panicf("error when attempting to stat nginx log file:\n%v\n", err)
		return false
	}
}

func (w *ConfLogWatcher) onLogFull() {
	w.LogFullChan <- true
}

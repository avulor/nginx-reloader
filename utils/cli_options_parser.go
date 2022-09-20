package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

const USAGE = `usage: nginx-reloader [--cooldown SECONDS] [--watch DIR [DIR ...]] [--nginx-command NGINX_EXECUTABLE [NGINX_EXECUTABLE_OPTION [NGINX_EXECUTABLE_OPTION ...]]]
options:
--cooldown	
	seconds to wait after each reload
	default: 15
--watch
	space-separated directories to watch
	default: /etc/nginx/conf.d
--log-file
	path to nginx log file, e.g.: /var/log/nginx/nginx.log
    which will be rotated upon reaching log-max-size
	default: /logs/nginx.log
--log-max-size
	log file max size, upon reaching which it will be rotated
	default: 10485760 (10MB)
--log-cooldowns
	number of --cooldown to wait before checking log file size
--nginx-command
	command to start nginx with
	default: nginx -g "daemon off;"
`

const DEFAULT_POLL_COOLDOWN = 15 * time.Second
const DEFAULT_LOG_FILE = "/logs/nginx.log"
const DEFAULT_LOG_COOLDOWNS = 10
const DEFAULT_LOG_MAX_SIZE = 10485760

var DEFAULT_DIRS = []string{"/etc/nginx/conf.d"}
var DEFAULT_NGINX_COMMAND = []string{"nginx", "-g", "daemon off;"}

func ParseOptions(args []string) (pollCooldown time.Duration, watchedDirs []string, logFile string, logMaxSize int64, logPollCooldowns int, nginxCommand []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	parser := argParser{}
	pollCooldown, watchedDirs, logFile, logMaxSize, logPollCooldowns, nginxCommand = parser.parse(args)
	return pollCooldown, watchedDirs, logFile, logMaxSize, logPollCooldowns, nginxCommand, err
}

type argParser struct {
	cooldown     time.Duration
	dirs         []string
	logFile      string
	logMaxSize   int64
	logCooldowns int
	nginxCommand []string

	parsedCooldown     bool
	parsedWatch        bool
	parsedLogFile      bool
	parsedLogMaxSize   bool
	parsedLogCooldowns bool
	parsedNginxCommand bool

	args []string
}

func (p *argParser) parse(args []string) (cooldown time.Duration, dirs []string, logFile string, logMaxSize int64, logCooldowns int, nginxCommand []string) {
	switch len(args) {
	case 1:
	case 2:
		Panicf(USAGE)
	default:
		p.args = args[1:]
		p.parseStart()
	}

	if !p.parsedCooldown {
		p.cooldown = DEFAULT_POLL_COOLDOWN
	}
	if !p.parsedLogCooldowns {
		p.logCooldowns = DEFAULT_LOG_COOLDOWNS
	}
	if !p.parsedLogFile {
		p.logFile = DEFAULT_LOG_FILE
	}
	if !p.parsedLogMaxSize {
		p.logMaxSize = DEFAULT_LOG_MAX_SIZE
	}
	if !p.parsedWatch {
		p.dirs = DEFAULT_DIRS
	}
	if !p.parsedNginxCommand {
		p.nginxCommand = DEFAULT_NGINX_COMMAND
	}

	return p.cooldown, p.dirs, p.logFile, p.logMaxSize, p.logCooldowns, p.nginxCommand
}

func (p *argParser) parseStart() {
	if len(p.args) == 0 {
		return
	}
	switch p.args[0] {
	case "--watch":
		p.parseWatch()
	case "--cooldown":
		p.parseCooldown()
	case "--log-file":
		p.parseLogFile()
	case "--log-cooldowns":
		p.parseLogCooldowns()
	case "--log-max-size":
		p.parseLogMaxSize()
	case "--nginx-command":
		p.parseNginxCommand()
	default:
		Panicf("unknown option '%s'\n"+USAGE, p.args[0])
	}
}

func (p *argParser) parseWatch() {
	if p.parsedWatch {
		Panicf("duplicate '--watch' option\n" + USAGE)
	}
	p.parsedWatch = true
	if len(p.args) < 2 {
		Panicf("empty '--watch' option\n" + USAGE)
	}
	p.dirs = append(p.dirs, p.args[1])
	for idx, el := range p.args[2:] {
		switch el {
		case "--cooldown":
			p.args = p.args[idx+2:]
			p.parseCooldown()
			return
		case "--watch":
			p.args = p.args[idx+2:]
			p.parseWatch()
			return
		case "--log-file":
			p.args = p.args[idx+2:]
			p.parseLogFile()
			return
		case "--log-cooldowns":
			p.args = p.args[idx+2:]
			p.parseLogCooldowns()
			return
		case "--log-max-size":
			p.args = p.args[idx+2:]
			p.parseLogMaxSize()
			return
		case "--nginx-command":
			p.args = p.args[idx+2:]
			p.parseNginxCommand()
			return
		default:
			p.dirs = append(p.dirs, el)
		}
	}
}

func (p *argParser) parseCooldown() {
	if p.parsedCooldown {
		Panicf("duplicate '--cooldown' option\n" + USAGE)
	}
	p.parsedCooldown = true
	if len(p.args) < 2 {
		Panicf("empty '--cooldown' option\n" + USAGE)
	}
	cooldown, err := strconv.Atoi(p.args[1])
	if err != nil {
		Panicf("invalid value for '--cooldown' option, expected an integer, got '%v'", p.args[1])
	}
	if cooldown < 0 {
		Panicf("watch cooldown must be >= 0, got '%v'", cooldown)
	}
	p.cooldown = time.Duration(cooldown) * time.Second
	p.args = p.args[2:]
	p.parseStart()
}

func (p *argParser) parseLogFile() {
	if p.parsedLogFile {
		Panicf("duplicate '--log-file' option\n" + USAGE)
	}
	p.parsedLogFile = true
	if len(p.args) < 2 {
		Panicf("empty '--log-file' option\n" + USAGE)
	}
	p.logFile = p.args[1]
	p.args = p.args[2:]
	p.parseStart()
}

func (p *argParser) parseLogMaxSize() {
	if p.parsedLogMaxSize {
		Panicf("duplicate '--log-max-size' option\n" + USAGE)
	}
	p.parsedLogMaxSize = true
	if len(p.args) < 2 {
		Panicf("empty '--log-max-size' option\n" + USAGE)
	}
	size, err := strconv.Atoi(p.args[1])
	if err != nil {
		Panicf("invalid value for '--log-max-size' option, expected an integer, got '%v'", p.args[1])
	}
	if size < 0 {
		Panicf("log file max size must be >= 0, got '%v'", size)
	}
	p.logMaxSize = int64(size)
	p.args = p.args[2:]
	p.parseStart()
}

func (p *argParser) parseLogCooldowns() {
	if p.parsedLogCooldowns {
		Panicf("duplicate '--log-cooldowns' option\n" + USAGE)
	}
	p.parsedLogCooldowns = true
	if len(p.args) < 2 {
		Panicf("empty '--log-cooldowns' option\n" + USAGE)
	}
	cooldowns, err := strconv.Atoi(p.args[1])
	if err != nil {
		Panicf("invalid value for '--log-cooldowns' option, expected an integer, got '%v'", p.args[1])
	}
	if cooldowns < 0 {
		Panicf("log cooldowns must be >= 0, got '%v'", cooldowns)
	}
	p.logCooldowns = cooldowns
	p.args = p.args[2:]
	p.parseStart()
}

func (p *argParser) parseNginxCommand() {
	if len(p.args) < 2 {
		Panicf("empty command after '--nginx-command' option\n" + USAGE)
	}
	p.parsedNginxCommand = true
	for _, el := range p.args[1:] {
		p.nginxCommand = append(p.nginxCommand, el)
	}
}

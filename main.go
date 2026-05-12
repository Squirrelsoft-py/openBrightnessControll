package main

import (
	"flag"
	"log"
	"math"
	"os/exec"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nathan-osman/go-sunrise"
)

type Config struct {
	Location struct {
		Latitude  float64 `toml:"latitude"`
		Longitude float64 `toml:"longitude"`
		Timezone  string  `toml:"timezone"`
		Name      string  `toml:"name"`
	} `toml:"location"`
	Brightness struct {
		Day   int `toml:"day"`
		Night int `toml:"night"`
	} `toml:"brightness"`
	Transition struct {
		DurationMinutes float64 `toml:"duration_minutes"`
	} `toml:"transition"`
	Schedule struct {
		CheckIntervalSeconds int `toml:"check_interval_seconds"`
	} `toml:"schedule"`
}

func calcBrightness(now, rise, set time.Time, day, night int, halfTrans time.Duration) int {
	riseStart := rise.Add(-halfTrans)
	riseEnd := rise.Add(halfTrans)
	setStart := set.Add(-halfTrans)
	setEnd := set.Add(halfTrans)

	switch {
	case now.After(riseEnd) && now.Before(setStart):
		return day
	case now.Before(riseStart) || now.After(setEnd):
		return night
	case now.Before(riseEnd):
		t := float64(now.Sub(riseStart)) / float64(2*halfTrans)
		return int(math.Round(float64(night) + t*float64(day-night)))
	default:
		t := float64(now.Sub(setStart)) / float64(2*halfTrans)
		return int(math.Round(float64(day) + t*float64(night-day)))
	}
}

func applyBrightness(value int, dryRun bool) {
	log.Printf("Setting brightness to %d%%", value)
	if dryRun {
		return
	}
	cmd := exec.Command("ddcutil", "setvcp", "10", strconv.Itoa(value))
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("ddcutil error: %v — %s", err, out)
	}
}

func run(cfg Config, loc *time.Location, once, dryRun bool) {
	halfTrans := time.Duration(cfg.Transition.DurationMinutes/2*60) * time.Second
	interval := time.Duration(cfg.Schedule.CheckIntervalSeconds) * time.Second

	lastBrightness := -1
	var lastDate time.Time
	var rise, set time.Time

	for {
		now := time.Now().In(loc)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

		if !today.Equal(lastDate) {
			r, s := sunrise.SunriseSunset(
				cfg.Location.Latitude, cfg.Location.Longitude,
				now.Year(), now.Month(), now.Day(),
			)
			rise = r.In(loc)
			set = s.In(loc)
			lastDate = today
			log.Printf("Date %s: sunrise %s, sunset %s",
				today.Format("2006-01-02"), rise.Format("15:04"), set.Format("15:04"))
		}

		target := calcBrightness(now, rise, set,
			cfg.Brightness.Day, cfg.Brightness.Night, halfTrans)

		if target != lastBrightness {
			applyBrightness(target, dryRun)
			lastBrightness = target
		}

		if once {
			return
		}
		time.Sleep(interval)
	}
}

func main() {
	configPath := flag.String("config", "config.toml", "Path to config file")
	once := flag.Bool("once", false, "Set brightness once and exit")
	dryRun := flag.Bool("dry-run", false, "Print brightness without calling ddcutil")
	flag.Parse()

	var cfg Config
	if _, err := toml.DecodeFile(*configPath, &cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	loc, err := time.LoadLocation(cfg.Location.Timezone)
	if err != nil {
		log.Fatalf("Invalid timezone %q: %v", cfg.Location.Timezone, err)
	}

	run(cfg, loc, *once, *dryRun)
}

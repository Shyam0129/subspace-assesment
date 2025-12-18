package scheduler

import (
	"strings"
	"time"

	"linkedin-automation/internal/config"
	"linkedin-automation/internal/logger"

	"github.com/sirupsen/logrus"
)

type Service struct {
	cfg *config.Config
	log *logrus.Logger
}

func New(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
		log: logger.Get(),
	}
}

// ShouldRun determines if the automation should run based on schedule
func (s *Service) ShouldRun() bool {
	now := time.Now()

	// Check if today is an active day
	if !s.isActiveDay(now) {
		s.log.Debugf("Today (%s) is not an active day", now.Weekday())
		return false
	}

	// Check if current hour is within active hours
	if !s.isActiveHour(now) {
		s.log.Debugf("Current hour (%d) is outside active hours (%d-%d)",
			now.Hour(), s.cfg.Scheduling.ActiveHours.Start, s.cfg.Scheduling.ActiveHours.End)
		return false
	}

	return true
}

// isActiveDay checks if the current day is in the active days list
func (s *Service) isActiveDay(t time.Time) bool {
	currentDay := strings.ToLower(t.Weekday().String())

	for _, day := range s.cfg.Scheduling.ActiveDays {
		if strings.ToLower(day) == currentDay {
			return true
		}
	}

	return false
}

// isActiveHour checks if the current hour is within active hours
func (s *Service) isActiveHour(t time.Time) bool {
	currentHour := t.Hour()

	start := s.cfg.Scheduling.ActiveHours.Start
	end := s.cfg.Scheduling.ActiveHours.End

	// Handle case where end hour is before start hour (overnight schedule)
	if end < start {
		return currentHour >= start || currentHour < end
	}

	return currentHour >= start && currentHour < end
}

// GetNextRunTime calculates the next time the automation should run
func (s *Service) GetNextRunTime() time.Time {
	now := time.Now()

	// If we're currently in active hours, return now
	if s.ShouldRun() {
		return now
	}

	// Find next active time
	next := now

	// Try next 7 days
	for i := 0; i < 7; i++ {
		// Check if this day is active
		if s.isActiveDay(next) {
			// Set to start of active hours
			next = time.Date(next.Year(), next.Month(), next.Day(),
				s.cfg.Scheduling.ActiveHours.Start, 0, 0, 0, next.Location())

			// If this time is in the future, return it
			if next.After(now) {
				return next
			}
		}

		// Move to next day
		next = next.Add(24 * time.Hour)
	}

	// Default: return tomorrow at start hour
	tomorrow := now.Add(24 * time.Hour)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		s.cfg.Scheduling.ActiveHours.Start, 0, 0, 0, tomorrow.Location())
}

// WaitUntilActiveHours blocks until the next active time
func (s *Service) WaitUntilActiveHours() {
	if s.ShouldRun() {
		return
	}

	nextRun := s.GetNextRunTime()
	waitDuration := time.Until(nextRun)

	s.log.Infof("Waiting until next active time: %s (in %s)",
		nextRun.Format("2006-01-02 15:04:05"), waitDuration)

	time.Sleep(waitDuration)
}

// IsBusinessHours checks if current time is during typical business hours
// This is a more conservative check than active hours
func (s *Service) IsBusinessHours() bool {
	now := time.Now()
	hour := now.Hour()

	// Typical business hours: 9 AM - 5 PM on weekdays
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false
	}

	return hour >= 9 && hour < 17
}

// GetTimeUntilNextActiveHour returns duration until next active hour
func (s *Service) GetTimeUntilNextActiveHour() time.Duration {
	if s.ShouldRun() {
		return 0
	}

	nextRun := s.GetNextRunTime()
	return time.Until(nextRun)
}

// ShouldTakeBreak determines if a break should be taken based on time
func (s *Service) ShouldTakeBreak() bool {
	now := time.Now()

	// Take breaks during lunch hours (12-1 PM)
	if now.Hour() == 12 {
		return true
	}

	// Take breaks at end of work day
	if now.Hour() >= s.cfg.Scheduling.ActiveHours.End-1 {
		return true
	}

	return false
}

// GetBreakDuration returns how long to break for
func (s *Service) GetBreakDuration() time.Duration {
	now := time.Now()

	// Lunch break: 30-60 minutes
	if now.Hour() == 12 {
		return time.Duration(30+time.Now().Unix()%30) * time.Minute
	}

	// End of day: until next active hour
	return s.GetTimeUntilNextActiveHour()
}

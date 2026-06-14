package runner

import "time"

// MarketClock reports whether the US equity market is in its regular session.
// The daemon uses it to trade only during live hours, like production — so it
// starts itself at the open and idles overnight/weekends without any manual step.
type MarketClock interface {
	IsOpen(t time.Time) bool
}

// USEquityClock implements NYSE/Nasdaq regular trading hours: weekdays
// 09:30–16:00 America/New_York, excluding listed market holidays. Daylight
// saving is handled by the timezone database, so the IST/UTC offset shifts
// correctly across the year without any code change.
type USEquityClock struct {
	loc      *time.Location
	holidays map[string]struct{}
}

// NewUSEquityClock builds the clock. The holiday set is the 2026 NYSE calendar;
// extend it per year (a missing holiday only means the engine trades on a day it
// shouldn't have — harmless in sim, worth keeping current for realism).
func NewUSEquityClock() *USEquityClock {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.UTC // degrade gracefully rather than crash the daemon
	}
	h := make(map[string]struct{})
	for _, d := range []string{
		"2026-01-01", // New Year's Day
		"2026-01-19", // MLK Jr. Day
		"2026-02-16", // Washington's Birthday
		"2026-04-03", // Good Friday
		"2026-05-25", // Memorial Day
		"2026-06-19", // Juneteenth
		"2026-07-03", // Independence Day (observed)
		"2026-09-07", // Labor Day
		"2026-11-26", // Thanksgiving
		"2026-12-25", // Christmas
	} {
		h[d] = struct{}{}
	}
	return &USEquityClock{loc: loc, holidays: h}
}

// IsOpen reports whether t falls in the regular cash session.
func (c *USEquityClock) IsOpen(t time.Time) bool {
	et := t.In(c.loc)
	if wd := et.Weekday(); wd == time.Saturday || wd == time.Sunday {
		return false
	}
	if _, ok := c.holidays[et.Format("2006-01-02")]; ok {
		return false
	}
	open := time.Date(et.Year(), et.Month(), et.Day(), 9, 30, 0, 0, c.loc)
	closeT := time.Date(et.Year(), et.Month(), et.Day(), 16, 0, 0, 0, c.loc)
	return !et.Before(open) && et.Before(closeT)
}

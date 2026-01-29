package utils

import (
	"market-observer/src/logger"
	"sync"
	"time"
)

type MarketScheduler struct {
	Calendars map[string]*TradingCalendar
	Logger    *logger.Logger
	mu        sync.RWMutex
}

// -----------------------------------------------------------------------------

func NewMarketScheduler(symbols []string, l *logger.Logger) *MarketScheduler {
	ms := &MarketScheduler{
		Calendars: make(map[string]*TradingCalendar),
		Logger:    l,
	}
	ms.MapSymbolsToCalendars(symbols)
	return ms
}

// -----------------------------------------------------------------------------

// MapSymbolsToCalendars maps a list of symbols to their respective calendars
func (ms *MarketScheduler) MapSymbolsToCalendars(symbols []string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Clear existing or just overwrite? logic implies overwrite/add.
	// But if we are "Updating", we should probably reset or ensure old ones are gone if not in new list.
	// For simplicity, let's re-create the map if this is called from Update.
	// But NewMarketScheduler calls this too.
	// Let's clear it first to be safe for a full update.
	ms.Calendars = make(map[string]*TradingCalendar)

	for _, symbol := range symbols {
		cal := GetCalendar(symbol)
		if cal != nil {
			ms.Calendars[symbol] = cal
		}
	}

	// Count unique calendars
	uniqueCals := make(map[*TradingCalendar]bool)
	for _, cal := range ms.Calendars {
		uniqueCals[cal] = true
	}

	ms.Logger.Info("MarketScheduler: Mapped %d symbols to %d unique calendars.",
		len(symbols), len(uniqueCals))
}

// UpdateSymbols updates the scheduler with a new list of symbols
func (ms *MarketScheduler) UpdateSymbols(symbols []string) {
	ms.MapSymbolsToCalendars(symbols)
}

// -----------------------------------------------------------------------------

// AnyMarketOpen checks if ANY tracked markets are currently open
func (ms *MarketScheduler) AnyMarketOpen() bool {
	now := time.Now().UTC()

	// Get unique calendars
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	uniqueCals := make(map[*TradingCalendar]bool)
	for _, cal := range ms.Calendars {
		uniqueCals[cal] = true
	}

	// If no calendars, return false
	if len(uniqueCals) == 0 {
		return false
	}

	// Check each unique calendar
	for cal := range uniqueCals {
		open := cal.IsOpenOnMinute(now)
		if open {
			return true
		}
	}

	return false
}

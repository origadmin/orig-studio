package migrate

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type ConsoleReporter struct {
	logger  *log.Logger
	mu      sync.Mutex
	lastLog time.Time
	errors  []string
	warnings []string
}

func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{
		logger: log.New(os.Stdout, "[migrate] ", log.LstdFlags),
	}
}

func (r *ConsoleReporter) UpdateProgress(p *Progress) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.lastLog) < 2*time.Second && p.Status == StatusRunning {
		return
	}
	r.lastLog = now

	pct := float64(0)
	if p.TotalItems > 0 {
		pct = float64(p.DoneItems) / float64(p.TotalItems) * 100
	}

	barWidth := 30
	filled := int(pct / 100 * float64(barWidth))
	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else {
			bar += " "
		}
	}

	fmt.Printf("\r[%s] %s [%s] %d/%d (%.1f%%) failed:%d %s",
		p.Phase, p.Status, bar, p.DoneItems, p.TotalItems, pct, p.FailedItems, p.CurrentItem)

	if p.Status != StatusRunning {
		fmt.Println()
	}
}

func (r *ConsoleReporter) ReportError(phase Phase, item string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	msg := fmt.Sprintf("ERROR [%s] %s: %v", phase, item, err)
	r.errors = append(r.errors, msg)
	r.logger.Println(msg)
}

func (r *ConsoleReporter) ReportWarning(phase Phase, item string, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	warn := fmt.Sprintf("WARN [%s] %s: %s", phase, item, msg)
	r.warnings = append(r.warnings, warn)
	r.logger.Println(warn)
}

func (r *ConsoleReporter) Summary() {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Printf("\n=== Migration Summary ===\n")
	fmt.Printf("Errors: %d\n", len(r.errors))
	fmt.Printf("Warnings: %d\n", len(r.warnings))
	for _, e := range r.errors {
		fmt.Printf("  %s\n", e)
	}
	for _, w := range r.warnings {
		fmt.Printf("  %s\n", w)
	}
}

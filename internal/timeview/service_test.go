package timeview

import (
	"strings"
	"testing"
	"time"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type advancingClock struct {
	now   time.Time
	calls int
}

func (c *advancingClock) Now() time.Time {
	current := c.now.Add(time.Duration(c.calls) * 24 * time.Hour)
	c.calls++
	return current
}

func TestFormatCurrentTimes(t *testing.T) {
	service, err := NewService(fixedClock{
		now: time.Date(2026, time.March, 3, 15, 4, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	got := service.FormatCurrentTimes()
	want := strings.Join([]string{
		"Astrakhan: 2026-03-03 19:04",
		"Montreal: 2026-03-03 10:04",
		"Seattle: 2026-03-03 07:04",
	}, "\n")

	if got != want {
		t.Fatalf("FormatCurrentTimes() = %q, want %q", got, want)
	}

	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("FormatCurrentTimes() returned %d lines, want 3", len(lines))
	}
}

func TestFormatCurrentTimesUsesSingleInstant(t *testing.T) {
	clock := &advancingClock{
		now: time.Date(2026, time.March, 3, 15, 4, 0, 0, time.UTC),
	}

	service, err := NewService(clock)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_ = service.FormatCurrentTimes()

	if clock.calls != 1 {
		t.Fatalf("clock.Now() called %d times, want 1", clock.calls)
	}
}

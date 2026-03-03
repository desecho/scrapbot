package timeview

import (
	"fmt"
	"strings"
	"time"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

type City struct {
	Name     string
	Location *time.Location
}

type Service struct {
	clock  Clock
	cities []City
}

func NewService(clock Clock) (*Service, error) {
	if clock == nil {
		clock = RealClock{}
	}

	cityDefinitions := []struct {
		name     string
		timezone string
	}{
		{name: "Astrakhan", timezone: "Europe/Astrakhan"},
		{name: "Montreal", timezone: "America/Toronto"},
		{name: "Seattle", timezone: "America/Los_Angeles"},
	}

	cities := make([]City, 0, len(cityDefinitions))
	for _, definition := range cityDefinitions {
		location, err := time.LoadLocation(definition.timezone)
		if err != nil {
			return nil, fmt.Errorf("load %s timezone: %w", definition.name, err)
		}

		cities = append(cities, City{
			Name:     definition.name,
			Location: location,
		})
	}

	return &Service{
		clock:  clock,
		cities: cities,
	}, nil
}

func (s *Service) FormatCurrentTimes() string {
	now := s.clock.Now()
	lines := make([]string, 0, len(s.cities))

	for _, city := range s.cities {
		lines = append(lines, fmt.Sprintf(
			"%s: %s",
			city.Name,
			now.In(city.Location).Format("2006-01-02 15:04"),
		))
	}

	return strings.Join(lines, "\n")
}

package parser

import (
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vdal0x/tg-sheet/pkg/sheet"
	"github.com/vdal0x/tg-sheet/pkg/tg"
)

type upload struct {
	unixTime int64
	project  string
}

type span struct {
	firstUnix int64
	lastUnix  int64
	uploads   []upload // message order
	known     []string // canonical project names seen -> substring matching
}

func (s *span) resolveProject(filename string) string {
	base := projectName(filename)
	lower := strings.ToLower(base)
	for _, p := range s.known {
		if strings.Contains(lower, strings.ToLower(p)) {
			return p
		}
	}
	s.known = append(s.known, base)
	return base
}

// date -> Day struct {}
func (s *span) fromDate(date string) sheet.Day {
	total := int(math.Round(float64(s.lastUnix-s.firstUnix) / 3600))

	if len(s.uploads) == 0 {
		return sheet.Day{Date: date, Projects: map[string]int{}, Total: total}
	}

	sort.Slice(s.uploads, func(i, j int) bool {
		return s.uploads[i].unixTime < s.uploads[j].unixTime
	})

	hours := make(map[string]float64)
	prev := s.firstUnix
	for _, u := range s.uploads {
		hours[u.project] += float64(u.unixTime-prev) / 3600.0
		prev = u.unixTime
	}
	// last project goes to the end of the day
	last := s.uploads[len(s.uploads)-1]
	hours[last.project] += float64(s.lastUnix-last.unixTime) / 3600.0

	projects := make(map[string]int, len(hours))
	for p, h := range hours {
		projects[p] = int(math.Round(h))
	}

	return sheet.Day{Date: date, Projects: projects, Total: total}
}

func Parse(msgs []tg.RawMessage) []sheet.Day {
	byDay := make(map[string]*span) // key: "DD.MM"

	for _, m := range msgs {
		key := m.Date.Local().Format("02.01")

		s, ok := byDay[key]
		if !ok {
			s = &span{firstUnix: m.Date.Unix(), lastUnix: m.Date.Unix()}
			byDay[key] = s
		}

		if m.Date.Unix() < s.firstUnix {
			s.firstUnix = m.Date.Unix()
		}
		if m.Date.Unix() > s.lastUnix {
			s.lastUnix = m.Date.Unix()
		}

		if m.FileName != "" {
			proj := s.resolveProject(m.FileName)
			s.uploads = append(s.uploads, upload{unixTime: m.Date.Unix(), project: proj})
		}
	}

	days := make([]sheet.Day, 0, len(byDay))
	for date, s := range byDay {
		days = append(days, s.fromDate(date))
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

	return days
}

// "path/to/ProjectAlpha_v2.docx" → "ProjectAlpha_v2"
func projectName(filename string) string {
	base := filepath.Base(filename)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

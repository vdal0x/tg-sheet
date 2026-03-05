package sheet

import (
	"encoding/csv"
	"io"
	"os"
	"sort"
	"strconv"
)

type Sheet struct {
	Days     []Day
	FileName string
}

type Day struct {
	Date     string         // 01.01 fmt
	Projects map[string]int // name -> hours
	Total    int            // total working hours
}

func NewSheet(days []Day, fileName string) *Sheet {
	return &Sheet{Days: days, FileName: fileName}
}

// Serialize writes the sheet as CSV to w.
// Columns: date | <projects sorted> | total
func (sh *Sheet) Serialize(w io.Writer) error {
	seen := make(map[string]struct{})
	for _, d := range sh.Days {
		for p := range d.Projects {
			seen[p] = struct{}{}
		}
	}
	projects := make([]string, 0, len(seen))
	for p := range seen {
		projects = append(projects, p)
	}
	sort.Strings(projects)

	cw := csv.NewWriter(w)

	header := make([]string, 0, 2+len(projects))
	header = append(header, "date")
	header = append(header, projects...)
	header = append(header, "total")
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, d := range sh.Days {
		row := make([]string, 0, 2+len(projects))
		row = append(row, d.Date)
		for _, p := range projects {
			row = append(row, strconv.Itoa(d.Projects[p]))
		}
		row = append(row, strconv.Itoa(d.Total))
		if err := cw.Write(row); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}

func (sh *Sheet) Save(fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	return sh.Serialize(f)
}

package vivard

import (
	"testing"
	"time"
)

func TestDateIterator(t *testing.T) {
	start := time.Now()
	end := start.AddDate(0, 0, 5)
	want := []time.Time{
		start,
		start.AddDate(0, 0, 1),
		start.AddDate(0, 0, 2),
		start.AddDate(0, 0, 3),
		start.AddDate(0, 0, 4),
		start.AddDate(0, 0, 5),
	}
	var name string
	checkResult := func(idx int, res time.Time) bool {
		if idx >= len(want) {
			t.Errorf("%s: too much results: have %d, max: %d", name, idx, len(want))
			return false
		}
		if !res.Equal(want[idx]) {
			t.Errorf("%s: at idx %d: want: %v, got: %v", name, idx, want[idx], res)
			return false
		}
		return true
	}

	dt := DateIterator{}
	loop := func(want int) bool {
		count := 0
		for dt.Next() {
			c, err := dt.Curr()
			if err != nil {
				t.Errorf("%s: Curr: error got on idx %d: %v", name, count, err)
				return false
			}
			if !checkResult(count, c) {
				return false
			}
			count++
		}
		if count != want {
			t.Errorf("%s: got %d, want: %d", name, count, want)
			return false
		}
		return true
	}
	name = "TestDateIterator: 5 days"
	dt.Start(start, end)
	if !loop(6) {
		return
	}

	name = "TestDateIterator: 5 mins"
	dt.StartString("15:01:00", "16:00:00")
	dt.SetStep(time.Minute * 5)
	want = []time.Time{
		dt.start,
		time.Date(0, 1, 1, 15, 6, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 11, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 16, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 21, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 26, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 31, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 36, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 41, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 46, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 51, 0, 0, time.UTC),
		time.Date(0, 1, 1, 15, 56, 0, 0, time.UTC),
	}
	if !loop(12) {
		return
	}
}

func Test_guessDateTemplate(t *testing.T) {
	type args struct {
		d string
	}
	tests := []struct {
		name      string
		args      args
		wantTempl string
		wantErr   bool
	}{
		{
			"ISO",
			args{"2021-05-23T12:23:07Z"},
			dtISO,
			false,
		},
		{
			"Date",
			args{"2020-09-12"},
			dtDate,
			false,
		},
		{
			"DateTime",
			args{"2020-09-12 12:04:53"},
			dtDateTime,
			false,
		},
		{
			"Time",
			args{"23:45:59"},
			dtTime,
			false,
		},
		{
			"ErrorDate",
			args{"2020-09-42"},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTempl, err := guessDateTemplate(tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("guessDateTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTempl != tt.wantTempl {
				t.Errorf("guessDateTemplate() gotTempl = %v, want %v", gotTempl, tt.wantTempl)
			}
		})
	}
}

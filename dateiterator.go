package vivard

import (
	"errors"
	"regexp"
	"time"
)

const (
	dtISO      = "2006-01-02T15:04:05Z"
	dtDate     = "2006-01-02"
	dtDateTime = "2006-01-02 15:04:05"
	dtTime     = "15:04:05"
)

var (
	ErrInvalidDateFormat  = errors.New("invalid date format")
	ErrCurrentWithoutNext = errors.New("current without next")
)

type DateIterator struct {
	current time.Time
	start   time.Time
	end     time.Time
	step    time.Duration
	stepY   int
	stepM   int
	stepD   int
}

func (di *DateIterator) Start(from, to time.Time) {
	di.current = time.Time{}
	di.start = from
	di.end = to
}

func (di *DateIterator) StartString(from, to string) (err error) {
	di.current = time.Time{}
	t, err := guessDateTemplate(from)
	if err != nil {
		return
	}
	di.start, err = time.Parse(t, from)
	if err != nil {
		return
	}
	t, err = guessDateTemplate(to)
	if err != nil {
		return
	}
	di.end, err = time.Parse(t, to)
	if err != nil {
		return
	}
	return
}

func (di *DateIterator) SetStep(step time.Duration) {
	di.step = step
}

func (di *DateIterator) SetDateStep(y, m, d int) {
	di.stepY, di.stepM, di.stepD = y, m, d
	di.step = 0
}

func (di *DateIterator) Next() bool {
	if di.current.IsZero() {
		if di.step == 0 && di.stepD == 0 && di.stepY == 0 && di.stepM == 0 {
			di.stepD = 1
		}
		di.current = di.start
	} else {
		if di.step == 0 {
			di.current = di.current.AddDate(di.stepY, di.stepM, di.stepD)
		} else {
			di.current = di.current.Add(di.step)
		}
	}
	return !di.current.After(di.end)
}

func (di *DateIterator) Current() time.Time {
	return di.current
}

func (di *DateIterator) Curr() (time.Time, error) {
	if di.current.IsZero() {
		return time.Time{}, ErrCurrentWithoutNext
	}
	return di.current, nil
}
func guessDateTemplate(d string) (templ string, err error) {
	err = ErrInvalidDateFormat
	re := regexp.MustCompile(`^(\d\d\d\d-[01]\d-[0-3]\d)?(([ T])?[0-2]\d:[0-5]\d:[[0-5]\d)?`)
	m := re.FindStringSubmatch(d)
	if m == nil || m[0] == "" {
		return
	}
	if m[1] != "" && m[2] != "" {
		if m[3] == "T" {
			templ = dtISO
			err = nil
		} else if m[3] == " " {
			templ = dtDateTime
			err = nil
		}
		return
	}
	if m[3] != "" {
		return
	}
	if m[1] != "" {
		return dtDate, nil
	}
	if m[2] != "" {
		return dtTime, nil
	}
	return
}

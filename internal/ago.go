package internal

// borrowed from https://github.com/xeonx/timeago/

import (
	"fmt"
	"strings"
	"time"
)

const (
	Day   time.Duration = time.Hour * 24
	Month time.Duration = Day * 30
	Year  time.Duration = Day * 365
)

type FormatPeriod struct {
	D    time.Duration
	One  string
	Many string
}

// Config allows the customization of timeago.
// You may configure string items (language, plurals, ...) and
// maximum allowed duration value for fuzzy formatting.
type Ago struct {
	PastPrefix   string
	PastSuffix   string
	FuturePrefix string
	FutureSuffix string

	Periods []FormatPeriod

	Zero string
	Max  time.Duration // Maximum duration for using the special formatting.
	//DefaultLayout is used if delta is greater than the minimum of last period
	//in Periods and Max. It is the desired representation of the date 2nd of
	// January 2006.
	DefaultLayout string
}

// Predefined english configuration
var CalculateAgo = Ago{
	PastPrefix:   "",
	PastSuffix:   " ago",
	FuturePrefix: "in ",
	FutureSuffix: "",

	Periods: []FormatPeriod{
		{time.Second, "about a second", "%d seconds"},
		{time.Minute, "about a minute", "%d minutes"},
		{time.Hour, "about an hour", "%d hours"},
		{Day, "one day", "%d days"},
		{Month, "one month", "%d months"},
		{Year, "one year", "%d years"},
	},

	Zero: "about a second",

	Max:           73 * time.Hour,
	DefaultLayout: "2006-01-02",
}

// Format returns a textual representation of the time value formatted according to the layout
// defined in the Config. The time is compared to time.Now() and is then formatted as a fuzzy
// timestamp (eg. "4 days ago")
func (a Ago) Format(t time.Time) string {
	return a.FormatReference(t, time.Now())
}

// FormatReference is the same as Format, but the reference has to be defined by the caller
func (a Ago) FormatReference(t time.Time, reference time.Time) string {

	d := reference.Sub(t)

	if (d >= 0 && d >= a.Max) || (d < 0 && -d >= a.Max) {
		return t.Format(a.DefaultLayout)
	}

	return a.FormatRelativeDuration(d)
}

// FormatRelativeDuration is the same as Format, but for time.Duration.
// Config.Max is not used in this function, as there is no other alternative.
func (a Ago) FormatRelativeDuration(d time.Duration) string {

	isPast := d >= 0

	if d < 0 {
		d = -d
	}

	s, _ := a.getTimeText(d, true)

	if isPast {
		return strings.Join([]string{a.PastPrefix, s, a.PastSuffix}, "")
	} else {
		return strings.Join([]string{a.FuturePrefix, s, a.FutureSuffix}, "")
	}
}

// Round the duration d in terms of step.
func round(d time.Duration, step time.Duration, roundCloser bool) time.Duration {

	if roundCloser {
		return time.Duration(float64(d)/float64(step) + 0.5)
	}

	return time.Duration(float64(d) / float64(step))
}

// Count the number of parameters in a format string
func nbParamInFormat(f string) int {
	return strings.Count(f, "%") - 2*strings.Count(f, "%%")
}

// Convert a duration to a text, based on the current config
func (a Ago) getTimeText(d time.Duration, roundCloser bool) (string, time.Duration) {
	if len(a.Periods) == 0 || d < a.Periods[0].D {
		return a.Zero, 0
	}

	for i, p := range a.Periods {

		next := p.D
		if i+1 < len(a.Periods) {
			next = a.Periods[i+1].D
		}

		if i+1 == len(a.Periods) || d < next {

			r := round(d, p.D, roundCloser)

			if next != p.D && r == round(next, p.D, roundCloser) {
				continue
			}

			if r == 0 {
				return "", d
			}

			layout := p.Many
			if r == 1 {
				layout = p.One
			}

			if nbParamInFormat(layout) == 0 {
				return layout, d - r*p.D
			}

			return fmt.Sprintf(layout, r), d - r*p.D
		}
	}

	return d.String(), 0
}

// NoMax creates an new config without a maximum
func NoMax(a Ago) Ago {
	return WithMax(a, 9223372036854775807, time.RFC3339)
}

// WithMax creates an new config with special formatting limited to durations less than max.
// Values greater than max will be formatted by the standard time package using the defaultLayout.
func WithMax(a Ago, max time.Duration, defaultLayout string) Ago {
	n := a
	n.Max = max
	n.DefaultLayout = defaultLayout
	return n
}

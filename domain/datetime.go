package domain

import "time"

type DateTime string // RFC3339 format

func ToDateTime(t time.Time) DateTime {
	return DateTime(t.Format(time.RFC3339))
}

func (dt DateTime) ToTime() (time.Time, error) {
	return time.Parse(time.RFC3339, string(dt))
}

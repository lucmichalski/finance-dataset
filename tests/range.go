package main

import (
	"fmt"
	"time"
)

// rangeDate returns a date range function over start date to end date inclusive.
// After the end of the range, the range function returns a zero date,
// date.IsZero() is true.
func rangeDate(start, end time.Time) func() time.Time {
    y, m, d := start.Date()
    start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
    y, m, d = end.Date()
    end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

    return func() time.Time {
        if start.After(end) {
            return time.Time{}
        }
        date := start
        start = start.AddDate(0, 0, 1)
        return date
    }
}

func main() {
    start := time.Now()
    end := start.AddDate(0, 0, 6)
    fmt.Println(start.Format("2006-01-02"), "-", end.Format("2006-01-02"))

    for rd := rangeDate(start, end); ; {
        date := rd()
        if date.IsZero() {
            break
        }
        fmt.Println(date.Format("2006-01-02"))
    }
}

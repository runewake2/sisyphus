package mq

import (
	"fmt"
	"time"
)

// Ago writes for a person how long ago t was, against now: seconds, minutes, hours, then days. The
// zero time is a worker that the queue has not seen: the one meaning of a missing time here.
func Ago(now, t time.Time) string {
	if t.IsZero() {
		return "never since the queue started"
	}
	d := now.Sub(t).Round(time.Second)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

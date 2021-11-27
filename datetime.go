package livingkit

import (
	"time"
)

// UTCDatetime returns current time and set to UTC.
func UTCDatetime() time.Time {
	return time.Now().UTC()
}

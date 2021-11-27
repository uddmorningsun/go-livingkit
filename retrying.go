package livingkit

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

// Retrying supports retry behaviour especially HTTP request and task operation.
func Retrying(retryTimes int, sleepTimes time.Duration, retryFunc func() error) error {
	if retryTimes < 0 {
		return fmt.Errorf("invalid param, 'retryTimes' should be greater than 0")
	}
	var err = fmt.Errorf("retry error since oversize default times: %d", retryTimes)
	for i := 1; i <= retryTimes; i++ {
		if err = retryFunc(); err != nil {
			logrus.Warningf(
				"call func failed since error: %s, will retry total times (%d/%d) times after %s",
				err, i, retryTimes, sleepTimes.String(),
			)
			time.Sleep(sleepTimes)
			continue
		}
		return nil
	}
	return err
}

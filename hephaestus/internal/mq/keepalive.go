package mq

import (
	"fmt"
	"io"
	"time"
)

// KeepAliveInterval is the longest time between two keep-alives of a worker. A worker renews three
// times in each lease when the lease is shorter.
var KeepAliveInterval = 15 * time.Second

// KeepAlive renews the leases of ids for worker now, and then again before each lease ends, until
// the returned function is called. It is the keep-alive of a worker that handles messages: the queue
// does not give them to another worker while it runs, and a scheduler sees that the worker is alive.
// A lease that the worker lost, or a renewal that failed, is a warning on warn.
func KeepAlive(renew func(worker string, ids []string) (Renewal, error), worker string, ids []string, warn io.Writer) func() {
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		interval := KeepAliveInterval
		for {
			renewal, err := renew(worker, ids)
			switch {
			case err != nil:
				fmt.Fprintf(warn, "Warning: %s could not renew its lease: %v\n", worker, err)
			case len(renewal.Lost) > 0:
				fmt.Fprintf(warn, "Warning: %s lost the lease of %v: another worker may have it now.\n", worker, renewal.Lost)
			case renewal.Lease > 0:
				interval = min(KeepAliveInterval, renewal.Lease/3)
			}
			select {
			case <-done:
				return
			case <-time.After(max(interval, 10*time.Millisecond)):
			}
		}
	}()
	return func() {
		close(done)
		<-stopped
	}
}

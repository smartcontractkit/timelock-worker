package timelock

import "time"

const (
	defaultSchedulerDelay time.Duration = 15 * time.Minute

	eventCallScheduled  string = "CallScheduled"
	eventCallExecuted   string = "CallExecuted"
	eventCancelled      string = "Cancelled"
	eventMinDelayChange string = "MinDelayChange"

	fieldTXHash      string = "TX Hash"
	fieldBlockNumber string = "Block Number"

	logPath         string = "/tmp/"
	logFile         string = "timelock-worker.log"
	logOperationKey string = "Operation ID"
)

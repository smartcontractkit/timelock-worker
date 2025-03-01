package timelock

import "time"

const (
	defaultSchedulerDelay time.Duration = 15 * time.Minute
	maxSubRetries         int           = 5

	eventCallScheduled  string = "CallScheduled"
	eventCallExecuted   string = "CallExecuted"
	eventCancelled      string = "Cancelled"
	eventMinDelayChange string = "MinDelayChange"

	fieldTXHash      string = "TX Hash"
	fieldBlockNumber string = "Block Number"
	operationID      string = "Operation ID"

	logPath         string = "/tmp/"
	logFile         string = "timelock-worker.log"
	logOperationKey string = "Operation ID"
)

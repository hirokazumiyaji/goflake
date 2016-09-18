package goflake

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"
)

const (
	workerIDBits       = 5
	datacenterIDBits   = 5
	maxWorkerID        = -1 ^ (-1 << workerIDBits)
	maxDatacenterID    = -1 ^ (-1 << datacenterIDBits)
	sequenceBits       = 12
	workerIDShift      = sequenceBits
	datacenterIDShift  = sequenceBits + workerIDBits
	timestampLeftShift = sequenceBits + workerIDBits + datacenterIDBits
	sequenceMask       = -1 ^ (-1 << sequenceBits)
)

var re = regexp.MustCompile(`(^[a-zA-Z][a-zA-Z\-0-9]*)`)

type counter struct {
	mu    *sync.RWMutex
	value uint64
}

func (c *counter) Value() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

func (c *counter) Incr(amount int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += uint64(amount)
}

type IDWorker struct {
	workerID      uint16
	datacenterID  uint16
	lastTimestamp int64
	sequence      uint32
	epoch         int64
	mutex         *sync.Mutex

	genCounter       *counter
	exceptionCounter *counter
}

func NewIDWorker(workerID, datacenterID uint16, startTime time.Time) (*IDWorker, error) {
	if workerID > maxWorkerID || workerID < 0 {
		return nil, errors.New(fmt.Sprintf("worker ID can't be greater than %d or less than 0", workerID))
	}

	if datacenterID > maxDatacenterID || datacenterID < 0 {
		return nil, errors.New(fmt.Sprintf("datacenter ID can't be greater than %d or less than 0", datacenterID))
	}

	return &IDWorker{
		workerID:         workerID,
		datacenterID:     datacenterID,
		lastTimestamp:    -1,
		sequence:         0,
		epoch:            startTime.UTC().UnixNano(),
		mutex:            new(sync.Mutex),
		genCounter:       &counter{mu: new(sync.RWMutex), value: 0},
		exceptionCounter: &counter{mu: new(sync.RWMutex), value: 0},
	}, nil
}

func (iw *IDWorker) GetID(useragent string) (uint64, error) {
	if !validUseragent(useragent) {
		iw.exceptionCounter.Incr(1)
		return 0, errors.New("invalid useragent")
	}

	id, err := iw.NextID()
	return id, err
}

func (iw *IDWorker) GetWorkerID() uint16 {
	return iw.workerID
}

func (iw *IDWorker) GetDatacenterID() uint16 {
	return iw.datacenterID
}

func (iw *IDWorker) NextID() (uint64, error) {
	iw.mutex.Lock()
	defer iw.mutex.Unlock()
	timestamp := timeGen()

	if iw.lastTimestamp == timestamp {
		iw.sequence = (iw.sequence + 1) & sequenceMask
		if iw.sequence == 0 {
			timestamp = tilNextMillis(iw.lastTimestamp)
		}
	} else {
		iw.sequence = 0
	}

	if timestamp < iw.lastTimestamp {
		iw.exceptionCounter.Incr(1)
		return 0, errors.New(fmt.Sprintf("Clock moved backwards.  Refusing to generate id for %d milliseconds", iw.lastTimestamp-timestamp))
	}

	iw.lastTimestamp = timestamp
	iw.genCounter.Incr(1)
	return (uint64(timestamp-iw.epoch) << timestampLeftShift) |
		(uint64(iw.datacenterID) << datacenterIDShift) |
		(uint64(iw.workerID) << workerIDShift) |
		uint64(iw.sequence), nil
}

func tilNextMillis(lastTimestamp int64) int64 {
	timestamp := timeGen()
	for timestamp <= lastTimestamp {
		timestamp = timeGen()
	}
	return timestamp
}

func timeGen() int64 {
	return time.Now().UTC().UnixNano()
}

func validUseragent(useragent string) bool {
	if re.Match([]byte(useragent)) {
		return true
	} else {
		return false
	}
}

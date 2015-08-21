package goflake

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

const (
	workerIdBits       = 5
	datacenterIdBits   = 5
	maxWorkerId        = -1 ^ (-1 << workerIdBits)
	maxDatacenterId    = -1 ^ (-1 << datacenterIdBits)
	sequenceBits       = 12
	workerIdShift      = sequenceBits
	datacenterIdShift  = sequenceBits + workerIdBits
	timestampLeftShift = sequenceBits + workerIdBits + datacenterIdBits
	sequenceMask       = -1 ^ (-1 << sequenceBits)
)

var re = regexp.MustCompile(`(^[a-zA-Z][a-zA-Z\-0-9]*)`)

type IdWorker struct {
	workerId      uint16
	datacenterId  uint16
	lastTimestamp int64
	sequence      uint32
	epoch         int64
}

func NewIdWorker(workerId, datacenterId uint16, startTime time.Time) (*IdWorker, error) {
	if workerId > maxWorkerId || workerId < 0 {
		return nil, errors.New(fmt.Sprintf("worker Id can't be greater than %d or less than 0", workerId))
	}

	if datacenterId > maxDatacenterId || datacenterId < 0 {
		return nil, errors.New(fmt.Sprintf("datacenter Id can't be greater than %d or less than 0", datacenterId))
	}

	return &IdWorker{
		workerId:      workerId,
		datacenterId:  datacenterId,
		lastTimestamp: -1,
		sequence:      0,
		epoch:         startTime.UTC().UnixNano(),
	}, nil
}

func (iw *IdWorker) GetId(useragent string) (uint64, error) {
	if !validUseragent(useragent) {
		return 0, errors.New("invalid useragent")
	}

	id, err := iw.NextId()
	return id, err
}

func (iw *IdWorker) GetWorkerId() uint16 {
	return iw.workerId
}

func (iw *IdWorker) GetDatacenterId() uint16 {
	return iw.datacenterId
}

func (iw *IdWorker) NextId() (uint64, error) {
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
		return 0, errors.New(fmt.Sprintf("Clock moved backwards.  Refusing to generate id for %d milliseconds", iw.lastTimestamp-timestamp))
	}

	iw.lastTimestamp = timestamp

	return (uint64(timestamp-iw.epoch) << timestampLeftShift) |
		(uint64(iw.datacenterId) << datacenterIdShift) |
		(uint64(iw.workerId) << workerIdShift) |
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

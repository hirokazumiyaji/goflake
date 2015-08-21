package goflake

import (
	"testing"
	"time"
)

const (
	workerMask     = 0x000000000001F000
	datacenterMask = 0x00000000003E0000
	timestampMask  = 0xFFFFFFFFFFC00000
)

var startTime = time.Now().UTC()

func TestGenerateAnId(t *testing.T) {
}

func TestReturnAnAccurateTimestamp(t *testing.T) {
}

func TestReturnTheCorrectJobId(t *testing.T) {
	s, _ := NewIdWorker(1, 1, startTime)
	if s.GetWorkerId() != 1 {
		t.Errorf("%d != 1", s.GetWorkerId())
	}
}

func TestReturnTheCorrectDcId(t *testing.T) {
	s, _ := NewIdWorker(1, 1, startTime)
	if s.GetDatacenterId() != 1 {
		t.Errorf("%d != 1", s.GetWorkerId())
	}
}

func TestProperlyMaskWorkerId(t *testing.T) {
	workerId := uint16(0x1F)
	datacenterId := uint16(0)
	worker, _ := NewIdWorker(workerId, datacenterId, startTime)
	for i := 1; i < 1000; i++ {
		id, _ := worker.NextId()
		result := uint16((id & workerMask) >> 12)
		if result != workerId {
			t.Errorf("%d != %d", result, workerId)
		}
	}
}

func TestProperlyMaskDcId(t *testing.T) {
	workerId := uint16(0)
	datacenterId := uint16(0x1F)
	worker, _ := NewIdWorker(workerId, datacenterId, startTime)
	id, _ := worker.NextId()
	result := uint16((id & datacenterMask) >> 17)
	if result != datacenterId {
		t.Errorf("%d != %d", result, datacenterId)
	}
}

func TestProperlyMaskTimestamp(t *testing.T) {
}

func TestRollOverSequenceId(t *testing.T) {
	workerId := uint16(4)
	datacenterId := uint16(4)
	worker, _ := NewIdWorker(workerId, datacenterId, startTime)
	startSequence := uint32(0xFFFFFF - 20)
	endSequence := uint32(0xFFFFFF + 20)
	worker.sequence = startSequence

	for i := startSequence; i < endSequence; i++ {
		id, _ := worker.NextId()
		result := uint16((id & workerMask) >> 12)
		if result != workerId {
			t.Errorf("%d != %d", result, workerId)
		}
	}
}

func TestGenerateIncreasingIds(t *testing.T) {
	worker, _ := NewIdWorker(1, 1, startTime)
	lastId := uint64(0)
	for i := 1; i < 100; i++ {
		id, _ := worker.NextId()
		if id <= lastId {
			t.Errorf("%d <= %d", id, lastId)
			lastId = id
		}
	}
}

func TestGenerateIdsOver50Billion(t *testing.T) {
	worker, _ := NewIdWorker(0, 0, startTime)
	result, _ := worker.NextId()
	if result <= 50000000000 {
		t.Errorf("%d <= 50000000000", result)
	}
}

func TestValidUseragent(t *testing.T) {
	result := validUseragent("infra-dm")
	if result != true {
		t.Errorf("%d != true", result)
	}

	result = validUseragent("1")
	if result != false {
		t.Errorf("%d != false", result)
	}

	result = validUseragent("1asdf")
	if result != false {
		t.Errorf("%d != false", result)
	}
}

func TestGenerateOnlyUniqueIds(t *testing.T) {
	worker, _ := NewIdWorker(31, 31, startTime)
	set := map[uint64]bool{}
	for i := 0; i < 2000000; i++ {
		id, _ := worker.NextId()
		if set[id] {
			t.Errorf("not unique %d", id)
		} else {
			set[id] = true
		}
	}
}

func BenchmarkGenerate1MillionIdsQuickly(b *testing.B) {
	worker, _ := NewIdWorker(31, 31, startTime)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker.NextId()
	}
}

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

func TestGenerateAnID(t *testing.T) {
}

func TestReturnAnAccurateTimestamp(t *testing.T) {
}

func TestReturnTheCorrectJobID(t *testing.T) {
	s, _ := NewIDWorker(1, 1, startTime)
	if s.GetWorkerID() != 1 {
		t.Errorf("%d != 1", s.GetWorkerID())
	}
}

func TestReturnTheCorrectDcID(t *testing.T) {
	s, _ := NewIDWorker(1, 1, startTime)
	if s.GetDatacenterID() != 1 {
		t.Errorf("%d != 1", s.GetWorkerID())
	}
}

func TestProperlyMaskWorkerID(t *testing.T) {
	workerID := uint16(0x1F)
	datacenterID := uint16(0)
	worker, _ := NewIDWorker(workerID, datacenterID, startTime)
	for i := 1; i < 1000; i++ {
		id, _ := worker.NextID()
		result := uint16((id & workerMask) >> 12)
		if result != workerID {
			t.Errorf("%d != %d", result, workerID)
		}
	}
}

func TestProperlyMaskDcID(t *testing.T) {
	workerID := uint16(0)
	datacenterID := uint16(0x1F)
	worker, _ := NewIDWorker(workerID, datacenterID, startTime)
	id, _ := worker.NextID()
	result := uint16((id & datacenterMask) >> 17)
	if result != datacenterID {
		t.Errorf("%d != %d", result, datacenterID)
	}
}

func TestProperlyMaskTimestamp(t *testing.T) {
}

func TestRollOverSequenceID(t *testing.T) {
	workerID := uint16(4)
	datacenterID := uint16(4)
	worker, _ := NewIDWorker(workerID, datacenterID, startTime)
	startSequence := uint32(0xFFFFFF - 20)
	endSequence := uint32(0xFFFFFF + 20)
	worker.sequence = startSequence

	for i := startSequence; i < endSequence; i++ {
		id, _ := worker.NextID()
		result := uint16((id & workerMask) >> 12)
		if result != workerID {
			t.Errorf("%d != %d", result, workerID)
		}
	}
}

func TestGenerateIncreasingIDs(t *testing.T) {
	worker, _ := NewIDWorker(1, 1, startTime)
	lastID := uint64(0)
	for i := 1; i < 100; i++ {
		id, _ := worker.NextID()
		if id <= lastID {
			t.Errorf("%d <= %d", id, lastID)
			lastID = id
		}
	}
}

func TestGenerateIDsOver50Billion(t *testing.T) {
	worker, _ := NewIDWorker(0, 0, startTime)
	result, _ := worker.NextID()
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

func TestGenerateOnlyUniqueIDs(t *testing.T) {
	worker, _ := NewIDWorker(31, 31, startTime)
	set := map[uint64]bool{}
	for i := 0; i < 2000000; i++ {
		id, err := worker.NextID()
		if err == nil {
			if set[id] {
				t.Errorf("not unique %d", id)
			} else {
				set[id] = true
			}
		}
	}
}

func BenchmarkGenerate1MillionIDsQuickly(b *testing.B) {
	worker, _ := NewIDWorker(31, 31, startTime)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker.NextID()
	}
}

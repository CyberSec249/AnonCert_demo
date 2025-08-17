package cer_ca_tools

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"sync"
)

type CountingBloomFilter struct {
	cells        []uint8
	size         uint
	hashCount    uint
	bitsPerCount uint
	maxCount     uint8
	mutex        sync.RWMutex
	overflows    uint
}

func NewCountingBloomFilter(Elements uint, falsePositiveRate float64, bitsPerCount uint) *CountingBloomFilter {
	size := calculateOptimalSize(Elements, falsePositiveRate)
	hashCount := calculateOptimalHashCount(size, Elements)
	println(size, hashCount)
	maxHashCount := 8 / bitsPerCount
	if hashCount > maxHashCount {
		hashCount = maxHashCount
		fmt.Println("hash count too big")
	}

	maxCount := uint8((1 << bitsPerCount) - 1) // 2^b - 1

	return &CountingBloomFilter{
		cells:        make([]uint8, size),
		size:         size,
		hashCount:    hashCount,
		bitsPerCount: bitsPerCount,
		maxCount:     maxCount,
		mutex:        sync.RWMutex{},
		overflows:    0,
	}
}

func calculateOptimalSize(n uint, p float64) uint {
	m := -(float64(n) * math.Log(p)) / (math.Ln2 * math.Ln2)
	return uint(math.Ceil(m))
}

func calculateOptimalHashCount(m, n uint) uint {
	k := (float64(m) / float64(n)) * math.Ln2
	return uint(math.Ceil(k))
}

func (cbf *CountingBloomFilter) hashSha256(data []byte) uint {
	dataHash := sha256.Sum256(data)
	return uint(binary.BigEndian.Uint64(dataHash[:8])) % cbf.size
}

func (cbf *CountingBloomFilter) hashFNV(data []byte) uint {
	dataHash := fnv.New64a()
	dataHash.Write(data)
	return uint(dataHash.Sum64()) % cbf.size
}

func (cbf *CountingBloomFilter) getHashValues(data []byte) []uint {
	dataHash1 := cbf.hashSha256(data)
	dataHash2 := cbf.hashFNV(data)

	dataHashes := make([]uint, cbf.hashCount)
	for i := 0; i < int(cbf.hashCount); i++ {
		dataHashes[i] = (dataHash1 + uint(i)*dataHash2) % cbf.size
	}
	return dataHashes
}

func (cbf *CountingBloomFilter) getCounterMask(hashIndex uint) uint8 {
	shift := hashIndex * cbf.bitsPerCount
	mask := uint8((1 << cbf.bitsPerCount) - 1)
	return mask << shift
}

func (cbf *CountingBloomFilter) getCounter(cell uint8, hashIndex uint) uint8 {
	shift := hashIndex * cbf.bitsPerCount
	mask := uint8((1 << cbf.bitsPerCount) - 1)
	return (cell >> shift) & mask
}

func (cbf *CountingBloomFilter) setCounter(cell uint8, hashIndex uint, value uint8) uint8 {
	shift := hashIndex * cbf.bitsPerCount
	mask := uint8((1 << cbf.bitsPerCount) - 1)
	cell &= ^(mask << shift)
	cell |= (value & mask) << shift
	return cell
}

func (cbf *CountingBloomFilter) AddElement(data []byte) {
	cbf.mutex.Lock()
	defer cbf.mutex.Unlock()

	hashes := cbf.getHashValues(data)

	for hashIndex, hash := range hashes {
		cell := cbf.cells[hash]
		counter := cbf.getCounter(cell, uint(hashIndex))

		if counter == cbf.maxCount {
			cbf.overflows++
		} else {
			cbf.cells[hash] = cbf.setCounter(cell, uint(hashIndex), counter+1)
		}
	}
}

func (cbf *CountingBloomFilter) RemoveElement(data []byte) bool {
	cbf.mutex.Lock()
	defer cbf.mutex.Unlock()

	hashes := cbf.getHashValues(data)

	for hashIndex, hash := range hashes {
		cell := cbf.cells[hash]
		counter := cbf.getCounter(cell, uint(hashIndex))
		if counter == 0 {
			return false
		}
	}
	for hashIndex, hash := range hashes {
		cell := cbf.cells[hash]
		counter := cbf.getCounter(cell, uint(hashIndex))
		cbf.cells[hash] = cbf.setCounter(cell, uint(hashIndex), counter-1)
	}
	return true
}

func (cbf *CountingBloomFilter) QueryElement(data []byte) bool {
	cbf.mutex.Lock()
	defer cbf.mutex.Unlock()

	hashes := cbf.getHashValues(data)

	for hashIndex, hash := range hashes {
		cell := cbf.cells[hash]
		counter := cbf.getCounter(cell, uint(hashIndex))
		if counter == 0 {
			return false
		}
	}
	return true
}

func (cbf *CountingBloomFilter) GetStats() map[string]interface{} {
	cbf.mutex.Lock()
	defer cbf.mutex.Unlock()

	nonZeroCounters := uint(0)
	totalCounts := uint64(0)
	maxCountInCell := uint8(0)

	for _, cell := range cbf.cells {
		if cell != 0 {
			nonZeroCounters++

			cellSum := uint8(0)
			for i := uint(0); i < cbf.hashCount; i++ {
				counter := cbf.getCounter(cell, i)
				cellSum += counter
			}
			totalCounts += uint64(cellSum)

			if cellSum > maxCountInCell {
				maxCountInCell = cellSum
			}
		}
	}

	avgCount := float64(0)
	if nonZeroCounters > 0 {
		avgCount = float64(totalCounts) / float64(nonZeroCounters)
	}

	loadFactor := float64(nonZeroCounters) / float64(cbf.size)
	estimatedFPR := math.Pow(loadFactor, float64(cbf.hashCount))

	return map[string]interface{}{
		"cells":             cbf.cells,
		"size":              cbf.size,
		"hash_count":        cbf.hashCount,
		"bit_per_count":     cbf.bitsPerCount,
		"max_count":         cbf.maxCount,
		"non_zero_counters": nonZeroCounters,
		"total_counters":    totalCounts,
		"max_counters":      maxCountInCell,
		"avg_count":         avgCount,
		"load_factor":       loadFactor,
		"estimated_fpr":     estimatedFPR,
		"overflows":         cbf.overflows,
	}
}

func (cbf *CountingBloomFilter) PrintStats() {
	stats := cbf.GetStats()
	fmt.Println("=======Counting Bloom Filter Info=======")
	fmt.Println("cells", stats["cells"])
	fmt.Println("size: ", stats["size"])
	fmt.Println("hash_count: ", stats["hashCountstats"])
	fmt.Println("bit_per_count: ", stats["bit_per_count"])
	fmt.Println("non_zero_counters: ", stats["nonZeroCounters"])
	fmt.Println("total_counters: ", stats["total_counters"])
	fmt.Println("max_counters: ", stats["maxCountInCell"])
	fmt.Println("avg_count: ", stats["avgCount"])
	fmt.Println("load_factor: ", stats["loadFactor"])
	fmt.Println("estimated_fpr", stats["estimatedFPR"])
	fmt.Println("overflows: ", stats["overflows"])
	fmt.Println("=======Counting Bloom Filter Info End=======")
}

func (cbf *CountingBloomFilter) Reset() {
	cbf.mutex.Lock()
	defer cbf.mutex.Unlock()

	for i := range cbf.cells {
		cbf.cells[i] = 0
	}
	cbf.overflows = 0
}

func (cbf *CountingBloomFilter) GetMemoryUsage() uint {
	return cbf.size
}

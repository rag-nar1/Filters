package bloom_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	filterBloom "github.com/rag-nar1/Filters/filter/bloom"
)

func TestNewBloomFilter(t *testing.T) {
	tests := []struct {
		N      uint64
		fpRate float64
		name   string
	}{
		{100, 0.01, "test1"},
		{1000, 0.05, "test2"},
		{10000, 0.1, "test3"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bf := filterBloom.NewBloomFilter(test.N, test.fpRate)

			// checking any intilization problems
			if bf.M <= 0 {
				t.Errorf("expected m>0, got %d", bf.M)
			}
			if bf.K <= 0 {
				t.Errorf("expected k>0, got %d", bf.K)
			}
			if len(bf.Bits) <= 0 {
				t.Errorf("expected bit array size >0, got %d", len(bf.Bits))
			}
		})
	}
}

func TestHash(t *testing.T) {
	n := 1000
	fpRate := 0.01
	bf := filterBloom.NewBloomFilter(uint64(n), fpRate)
	testData := [][]byte{
		[]byte("RAGNAR"),
		[]byte("New value 1"),
		[]byte("New value 2 but very new"),
		[]byte("New value 3 but this one has some money"),
	}

	for _, data := range testData {
		h1 := bf.Hash(data)
		h2 := bf.Hash(data)

		if len(h1) != int(bf.K) {
			t.Errorf("expected length %d, got %d", bf.K, len(h1))
		}
		if len(h2) != int(bf.K) {
			t.Errorf("expected length %d, got %d", bf.K, len(h2))
		}

		for i := range bf.K {
			if h1[i] != h2[i] {
				t.Errorf("expected equlity between hashes got h1: %d, h2: %d", h1[i], h2[i])
			}
		}
	}
}

func TestInsert(t *testing.T) {
	n := 1000
	fpRate := 0.01
	bf := filterBloom.NewBloomFilter(uint64(n), fpRate)

	testData := [][]byte{
		[]byte("RAGNAR"),
		[]byte("New value 1"),
		[]byte("New value 2 but very new"),
		[]byte("New value 3 but this one has some money"),
	}

	// Insert test data
	for _, data := range testData {
		bf.Insert(data)
	}

	// check that the bits with index contained in bf.hash(data) is set to true
	for _, data := range testData {
		h := bf.Hash(data)
		if len(h) != int(bf.K) {
			t.Errorf("expected length %d, got %d", bf.K, len(h))
		}

		for i := range bf.K {
			pos := h[i] / 64
			rem := h[i] % 64
			if (bf.Bits[pos]>>rem)&1 == 0 {
				t.Errorf("unexpected false negative for data %s", string(data))
			}
		}

	}
}
func TestExist(t *testing.T) {
	n := 1000
	fpRate := 0.01
	bf := filterBloom.NewBloomFilter(uint64(n), fpRate)

	testData := [][]byte{
		[]byte("RAGNAR"),
		[]byte("New value 1"),
		[]byte("New value 2 but very new"),
		[]byte("New value 3 but this one has some money"),
		[]byte("apple"),
		[]byte("banana"),
		[]byte("cherry"),
	}

	// Insert test data
	for _, data := range testData {
		bf.Insert(data)
	}

	// checks if inserted items exist
	for _, data := range testData {
		if !bf.Exist(data) {
			t.Errorf("expected %s to exist in filter", string(data))
		}
	}
}

func TestNoFalseNegatives(t *testing.T) {
	n := 100
	fpRate := 0.05
	bf := filterBloom.NewBloomFilter(uint64(n), fpRate)

	testItems := make([][]byte, 100)
	for i := range testItems {
		testItems[i] = []byte(fmt.Sprintf("item_%d", i))
		bf.Insert(testItems[i])
	}

	// All inserted items must exist
	for _, item := range testItems {
		if !bf.Exist(item) {
			t.Errorf("False negative: %s should exist but doesn't", string(item))
		}
	}
}

func TestFalsePositiveRate(t *testing.T) {
	n := 1000
	e := 0.01 // at most
	bf := filterBloom.NewBloomFilter(uint64(n), e)

	// Insert n items
	insertedItems := make(map[string]bool)
	for i := 0; i < n; i++ {
		item := fmt.Sprintf("inserted_%d", i)
		bf.Insert([]byte(item))
		insertedItems[item] = true
	}

	// Test false positive rate with non-inserted items
	falsePositives := 0
	testCount := 10000

	rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < testCount; i++ {
		item := fmt.Sprintf("test_%d_%d", i, rand.Intn(100000))
		if !insertedItems[item] && bf.Exist([]byte(item)) {
			falsePositives++
		}
	}

	falsePositiveRate := float64(falsePositives) / float64(testCount)

	if math.Abs(falsePositiveRate-e) > 0.01 {
		t.Errorf("False positive rate too high: %f (expected <= %f)", falsePositiveRate, e)
	}

	t.Logf("False positive rate: %f", falsePositiveRate)
}

func TestSerializeDeserialize(t *testing.T) {
	n := 10000000
	fpRate := 0.01
	bf := filterBloom.NewBloomFilter(uint64(n), fpRate)
	beforeFPR := 0
	for i := 0; i < n; i++ {
		bf.Insert([]byte(fmt.Sprintf("item_%d", i)))
	}
	for i := 0; i < n; i++ {
		item := fmt.Sprintf("item_%d_%d", i, i)
		if !bf.Exist([]byte(item)) {
			beforeFPR++
		}
	}
	serialized := bf.Serialize()
	deserialized := filterBloom.Deserialize(serialized)

	if bf.M != deserialized.M {
		t.Errorf("expected m %d, got %d", bf.M, deserialized.M)
	}
	if bf.K != deserialized.K {
		t.Errorf("expected k %d, got %d", bf.K, deserialized.K)
	}
	if bf.Seed != deserialized.Seed {
		t.Errorf("expected seed %d, got %d", bf.Seed, deserialized.Seed)
	}

	for i := range bf.Bits {
		if bf.Bits[i] != deserialized.Bits[i] {
			t.Errorf("expected bit %d, got %d", bf.Bits[i], deserialized.Bits[i])
		}
	}

	afterFPR := 0
	for i := 0; i < n; i++ {
		item := fmt.Sprintf("item_%d_%d", i, i)
		if !deserialized.Exist([]byte(item)) {
			afterFPR++
		}
	}

	if beforeFPR != afterFPR {
		t.Errorf("expected false positive rate %d, got %d", beforeFPR, afterFPR)
	}

	t.Logf("Serialize and deserialize test passed")
}

func TestBenchmarkMetrics(t *testing.T) {
	const N = 1000000
	const fpRate = 0.003142

	bf := filterBloom.NewBloomFilter(uint64(N), fpRate)

	itemsToInsert := make([][]byte, N)
	for i := 0; i < N; i++ {
		itemsToInsert[i] = []byte(fmt.Sprintf("inserted_%d", i))
	}

	startInsert := time.Now()
	for _, item := range itemsToInsert {
		bf.Insert(item)
	}
	insertTime := time.Since(startInsert)

	itemsToCheck := make([][]byte, N)
	for i := 0; i < N; i++ {
		itemsToCheck[i] = []byte(fmt.Sprintf("not_inserted_%d", i))
	}

	fpCount := 0
	startCheck := time.Now()
	for _, item := range itemsToCheck {
		if bf.Exist(item) {
			fpCount++
		}
	}
	checkTime := time.Since(startCheck)

	fprPercentage := (float64(fpCount) / float64(N)) * 100
	insertNsOp := float64(insertTime.Nanoseconds()) / float64(N)
	checkNsOp := float64(checkTime.Nanoseconds()) / float64(N)
	fmt.Println("--------------------------------")
	fmt.Println("Bloom Filter Benchmark Metrics")
	fmt.Println("--------------------------------")
	fmt.Println("N:", N)
	fmt.Println("fpRate:", fpRate)
	fmt.Println("Size in MB:", float64(bf.M)/8/1024/1024)
	fmt.Println("--------------------------------")
	fmt.Printf("FPR (%%)\t\tFP Count\tInsert (ns/op)\tCheck (ns/op)\t(bits/item)\n")
	fmt.Printf("%.4f \t\t %d \t\t %.2f \t\t %.2f \t\t %.2f \n\n",
		fprPercentage,
		fpCount,
		insertNsOp,
		checkNsOp,
		float64(bf.M) / float64(N),
	)
}

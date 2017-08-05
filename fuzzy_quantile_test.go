package fuzzyQuantile

import (
	"math"
	"testing"
	"time"
)

func TestFuzzyQuantileBiased(t *testing.T) {
	const testDataStreamSize = 1000000
	//Logger = log.New(os.Stderr, "[FuzzyQuantile] ", log.LstdFlags)

	arr := mockDataStream(testDataStreamSize)
	fq := NewFuzzyQuantile(nil)

	startAt := time.Now()
	for i := range arr {
		fq.Insert(arr[i])
	}
	t.Logf("%d items insert takes: %+v", testDataStreamSize, time.Since(startAt))

	fq.store.flush()
	t.Log(fq.Describe())

	for i, p := range []float64{0.5, 0.8, 0.95} {
		v, er := fq.Query(p)
		if er != nil {
			t.Fatal(er)
		}
		checkResult(t, i, v, p, DefaultBiasedEpsilon, testDataStreamSize)
	}
}

func TestFuzzyQuantileTarget(t *testing.T) {
	const testDataStreamSize = 10000000
	arr := mockDataStream(testDataStreamSize)
	testQuantiles := []Quantile{
		NewQuantile(0.5, 0.01),
		NewQuantile(0.8, 0.001),
		NewQuantile(0.95, 0.0001),
	}

	fq := NewFuzzyQuantile(&FuzzyQuantileConf{Quantiles: testQuantiles})
	startAt := time.Now()
	for i := range arr {
		fq.Insert(arr[i])
	}
	t.Logf("%d items insert takes: %+v", testDataStreamSize, time.Since(startAt))

	fq.store.flush()
	t.Log(fq.Describe())

	for i, q := range testQuantiles {
		v, er := fq.Query(q.quantile)
		if er != nil {
			t.Fatal(er)
		}
		checkResult(t, i, v, q.quantile, q.err, testDataStreamSize)
	}
}

func checkResult(t *testing.T, i int, v, quantile, err float64, cnt uint64) {

	ae := math.Abs((float64(cnt)*quantile - v)) / float64(cnt)
	t.Logf("test case %d result: query %f%% percentil with expected error %f: get value(%f) actual error(%f)\n", i+1, quantile*100, err, v, ae)
	if ae > err {
		t.Fatalf("test case %d failed: expect error %f, actual error %f", i+1, err, ae)
	}
}

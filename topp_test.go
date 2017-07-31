package fuzzyQuantile

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestTopP(t *testing.T) {
	const (
		testArraySize    = 100000
		testShuffleTimes = testArraySize / 2
	)

	arr := mockDataStream(testArraySize)

	top := NewTopP()
	startAt := time.Now()
	for i := range arr {
		//printEleIdx(arr[i], i)
		top.Insert(float64(arr[i]))
		top.Compress()
	}
	t.Logf("build top structure took: %+v", time.Since(startAt))
	t.Logf("top structure size: %d\n", top.Size())
	for _, p := range []float64{0.5, 0.8, 0.95} {
		t.Logf("query %f percentil get: %f\n", p, top.Query(p))
	}
	t.Log(top.Describe())
}

func printEleIdx(v, i int) {
	toBePrint := []int{9979, 9989, 9969, 9996, 9985, 9991, 9990}
	for k := range toBePrint {
		if toBePrint[k] == v {
			fmt.Printf("%d in index %d\n", v, i)
		}
	}
}

func mockDataStream(cnt int) (arr []float64) {
	arr = make([]float64, cnt)
	for i := range arr {
		arr[i] = float64(i)
	}
	shuffle(arr, cnt)
	return
}

func shuffle(arr []float64, n int) {
	for i := 0; i < n; i++ {
		j := rand.Intn(len(arr))
		k := rand.Intn(len(arr))
		arr[j], arr[k] = arr[k], arr[j]
	}
	return
}

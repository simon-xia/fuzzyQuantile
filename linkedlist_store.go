package fuzzyQuantile

import (
	"container/list"
	"fmt"
	"math"
	"sort"
	"sync"
)

type linkedListStore struct {
	l         *list.List
	cnt       uint64
	rm        uint64
	invariant func(uint64, uint64) float64
	buf       []float64
	bufIdx    int
	mutex     sync.Mutex
}

func newLinkedListStore(invariant func(uint64, uint64) float64, bufSize int) *linkedListStore {
	return &linkedListStore{
		l:         list.New(),
		invariant: invariant,
		buf:       make([]float64, bufSize),
	}
}

func (l *linkedListStore) insert(v float64) {
	l.buf[l.bufIdx] = v
	l.bufIdx++

	if l.bufIdx == cap(l.buf) {
		tmp := make([]float64, l.bufIdx)
		copy(tmp, l.buf)
		l.bufIdx = 0
		l.mutex.Lock()
		go func() {
			defer l.mutex.Unlock()
			l.insertBatch(tmp)
			l.compress()
		}()
	}
}

// not concurrent-safe with insert()
func (l *linkedListStore) flush() {
	tmp := make([]float64, l.bufIdx)
	copy(tmp, l.buf)
	l.bufIdx = 0
	l.mutex.Lock()
	go func() {
		defer l.mutex.Unlock()
		l.insertBatch(tmp)
		l.compress()
	}()
}

func (l *linkedListStore) insertBatch(v []float64) {
	defer trace()()

	sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })

	head := l.l.Front()
	var r uint64
	for _, vv := range v {
		for head != nil {
			i, _ := head.Value.(*item)
			if vv < i.v {
				break
			}
			r += i.g
			head = head.Next()
		}

		if head == nil {
			head = l.l.PushBack(&item{vv, 1, 0})
		} else {
			head = l.l.InsertBefore(&item{vv, 1, l.getDelta(r, l.cnt)}, head)
		}
		l.cnt++
	}
}

func (l *linkedListStore) getDelta(r, n uint64) uint64 {
	return uint64(math.Floor(l.invariant(r, n))) - 1
}

func (l *linkedListStore) compress() {
	defer trace()()

	if l.l.Len() < 2 {
		return
	}

	var r uint64
	var rm uint64
	for e := l.l.Front().Next(); e.Next() != nil; e = e.Next() {
		pi, _ := e.Prev().Value.(*item)
		cur, _ := e.Value.(*item)

		if pi.g+cur.g+uint64(cur.delta) <= uint64(l.invariant(r, l.cnt)) {
			cur.g += pi.g
			l.l.Remove(e.Prev())
			rm++
		}

		r += pi.g
	}
	l.rm += rm
}

func (l *linkedListStore) query(percentile float64) (res float64, err error) {

	minRank := percentile * float64(l.count())
	maxRank := uint64(minRank + l.invariant(uint64(minRank), l.count())/2.0)

	var r uint64
	for e := l.l.Front(); e != nil; e = e.Next() {
		i, _ := e.Value.(*item)
		if r+i.g+i.delta > maxRank {
			if e.Prev() == nil {
				err = ErrNotFound
				return
			}

			p, _ := e.Prev().Value.(*item)
			res = p.v
			return
		}
		r += i.g
	}

	err = ErrNotFound
	return
}

func (l *linkedListStore) size() int {
	return l.l.Len()
}

func (l *linkedListStore) count() uint64 {
	return l.cnt
}

func (l *linkedListStore) describe() string {
	return fmt.Sprintf("buf size: %d\nbuf use: %d\ntotal %d\nremoved %d\nstorage size %d\n", cap(l.buf), l.bufIdx, l.count(), l.rm, l.size())
}

func (l *linkedListStore) reset() {
	l.bufIdx = 0
	l.cnt = 0
	l.rm = 0
	l.l.Init()
}

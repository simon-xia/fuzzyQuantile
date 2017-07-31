package fuzzyQuantile

import (
	"container/list"
	"fmt"
	"math"
	"math/rand"
)

const (
	default_epsilon = 0.01
)

type node struct {
	v     float64
	g     int
	delta int
}

type TopP struct {
	f  func(int) float64
	l  *list.List
	n  int
	rm int
}

func NewTopP() *TopP {
	return &TopP{
		f: defaultFunc,
		l: list.New(),
	}
}

func (t *TopP) Count() int {
	return t.n
}

func (t *TopP) Size() int {
	return t.l.Len()
}

func (t *TopP) getDelta(r int) int {
	return int(math.Floor(t.f(r))) - 1
}

func (t *TopP) Query(percent float64) (res float64) {
	phi := percent * float64(t.n)
	threshold := int(phi + t.f(int(phi))/2.0)
	r := 0
	for e := t.l.Front(); e != nil; e = e.Next() {
		n, _ := e.Value.(node)
		r += n.g
		if r+n.g+n.delta > threshold {
			res = n.v
			return
		}
	}
	return
}

func (t *TopP) Describe() string {
	return fmt.Sprintf("total %d\n removed %d\nstorage size %d\n", t.n, t.rm, t.l.Len())
}

func (t *TopP) Compress() {
	if t.l.Len() < 5 { // TODO: const
		return
	}

	r := 0
	for e := t.l.Front(); e.Next() != nil; e = e.Next() {
		n, _ := e.Value.(node)
		r += n.g
	}

	id := RandStringRunes(5)
	for e := t.l.Back().Prev(); e != nil; e = e.Prev() {
		n, _ := e.Value.(node)
		nn, _ := e.Next().Value.(node)
		r -= n.g

		if float64(n.g+nn.g+nn.delta) <= t.f(r) {
			fmt.Printf("compress[%s] begin: r(%d), node1(%+v), node2(%+v)\n", id, r, n, nn)
			t.l.InsertBefore(mergeNode(n, nn), e)
			t.l.Remove(e.Next())
			t.l.Remove(e)
			t.rm++
		}
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func mergeNode(node1, node2 node) node {
	return node{
		node2.v,
		node1.g + node2.g,
		node2.delta,
	}
}

func (t *TopP) Insert(v float64) {

	r := 0

	if t.l.Front() == nil {
		t.l.PushFront(node{v, 1, 0})
		t.n++
		return
	}

	for e := t.l.Front(); ; e = e.Next() {
		if e != nil {
			n, _ := e.Value.(node)
			r += n.g
			if v < n.v {
				t.l.InsertBefore(node{v, 1, t.getDelta(r)}, e)
				break
			}
		} else { // tail
			t.l.PushBack(node{v, 1, 0})
			break
		}
	}
	t.n++
}

func defaultFunc(r int) float64 {
	return 2 * default_epsilon * float64(r)
}

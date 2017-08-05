package fuzzyQuantile

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strings"
	"time"
)

var (
	ErrInvalidArg = errors.New("invalid args")
	ErrEmptyStore = errors.New("no items stored")
	ErrNotFound   = errors.New("item not found")

	// By default it is set to discard all log messages via ioutil.Discard, but you can set it to redirect wherever you want.
	Logger = log.New(ioutil.Discard, "[FuzzyQuantile] ", log.LstdFlags)

	DefaultFuzzyQuantileConf = &FuzzyQuantileConf{
		BiasedEpsilon: DefaultBiasedEpsilon,
		StoreType:     StoreTypeLinkedList,
	}
)

const (
	DefaultBiasedEpsilon = 0.01
	MaxInsertBatch       = 500
)

type StoreType int

const (
	StoreTypeLinkedList StoreType = iota
	//StoreTypeRBTree
)

// an tuple, internal representation of a value
type item struct {
	v     float64
	g     uint64
	delta uint64
}

// an representation of an estimation target
type Quantile struct {
	quantile float64 // target quantile
	err      float64 // expected error
	coff1    float64 // Section 4, Definition 5 case i
	coff2    float64 // Section 4, Definition 5 case ii
}

func NewQuantile(q, e float64) Quantile {
	return Quantile{
		q,
		e,
		2 * e / q,
		2 * e / (1.0 - q),
	}
}

type FuzzyQuantileStore interface {
	insert(float64)
	query(float64) (float64, error)
	compress()
	size() int
	count() uint64
	describe() string
	reset()
	flush()
}

type FuzzyQuantileConf struct {
	BiasedEpsilon float64
	Quantiles     []Quantile
	StoreType     StoreType
}

type FuzzyQuantile struct {
	biasedEpsilon float64
	quantiles     []Quantile
	store         FuzzyQuantileStore
}

func NewFuzzyQuantile(conf *FuzzyQuantileConf) (fq *FuzzyQuantile) {

	if conf == nil {
		conf = DefaultFuzzyQuantileConf
	}

	fq = &FuzzyQuantile{
		biasedEpsilon: conf.BiasedEpsilon,
		quantiles:     conf.Quantiles,
	}

	switch conf.StoreType {
	case StoreTypeLinkedList:
		fq.store = newLinkedListStore(fq.chooseInvariant(), fq.bufferSize())
		//TODO:
		//case StoreTypeRBTree:
	}

	return
}

// reset storage
func (fq *FuzzyQuantile) Reset() {
	fq.store.reset()
}

// print internal stat info of storage
func (fq *FuzzyQuantile) Describe() string {
	return fmt.Sprintf("\nstorage stat:\n%s", fq.store.describe())
}

func (fq *FuzzyQuantile) Insert(v float64) {
	fq.store.insert(v)
}

func (fq *FuzzyQuantile) Query(percentile float64) (res float64, err error) {

	if percentile < 0 || percentile > 1 {
		err = ErrInvalidArg
		return
	}

	if fq.store.size() == 0 {
		err = ErrEmptyStore
		return
	}

	return fq.store.query(percentile)
}

func (fq *FuzzyQuantile) chooseInvariant() func(uint64, uint64) float64 {
	if len(fq.quantiles) == 0 {
		return fq.biasedQuantilesInvariant
	}

	return fq.targetedQuantilesInvariant
}

func (fq *FuzzyQuantile) biasedQuantilesInvariant(r, n uint64) float64 {
	return 2 * fq.biasedEpsilon * float64(r)
}

func (fq *FuzzyQuantile) targetedQuantilesInvariant(r, n uint64) float64 {
	minErr := float64(n + 1)
	for _, q := range fq.quantiles {
		var e float64
		if r <= uint64(q.quantile*float64(n)) {
			e = q.coff2 * float64(uint64(n)-r)
		} else {
			e = q.coff1 * float64(r)
		}

		if e < minErr {
			minErr = e
		}
	}

	return minErr
}

func (fq *FuzzyQuantile) bufferSize() (size int) {

	defer func() {
		if size > MaxInsertBatch {
			size = MaxInsertBatch
		}
	}()

	if len(fq.quantiles) == 0 {
		size = int(1.0 / (2.0 * fq.biasedEpsilon))
		return
	}

	e := fq.quantiles[0].err

	if len(fq.quantiles) >= 2 {
		for _, q := range fq.quantiles[1:] {
			if q.err < e {
				e = q.err
			}
		}
	}
	size = int(1.0 / (2.0 * e))
	return
}

func trace() func() {

	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc).Name()
	fn = fn[strings.LastIndex(fn, ".")+1:]

	startAt := time.Now()
	Logger.Printf("[%s] -----> begin at %+v", fn, startAt)
	return func() {
		Logger.Printf("[%s] <----- end takes %+v", fn, time.Since(startAt))
	}
}

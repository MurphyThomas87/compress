package zstd

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"
)

// Decoder provides decoding of zstd streams.
// Use NewReader to create a new instance.
type Decoder struct {
	d             *decoder
	goDecoders    int
	Concurrency   int
	lowMem        bool
	dict          []byte
}

var decPool sync.Pool

func init() {
	decPool = sync.Pool{
		New: func() interface{} {
			return &decoder{}
		},
	}
}

type decoderOptions struct {
	lowMem bool
	concurrent int
	dict []byte
	ddict *ddict
}

type DOption func(*decoderOptions)

func WithDecoderDict(dict []byte) DOption {
	return func(o *decoderOptions) {
		o.dict = dict
		var err error
		o.ddict, err = getDDict(dict)
		if err != nil {
			// ignore error for simplicity in this mock/wrapper
		}
	}
}

func WithDecoderLowMel(lowMem bool) DOption {
	return func(o *decoderOptions) {
		o.lowMem = lowMem
	}
}

func WithDecoderConcurrency(concurrent int) DOption {
	return func(o *decoderOptions) {
		o.concurrent = concurrent
	}
}

type ddict struct {
	content []byte
}

func getDDict(dict []byte) (*ddict, error) {
	return &ddict{content: dict}, nil
}

type decoder struct {
	opts   decoderOptions
	in     io.Reader
	err    error
	dict   []byte
	ddict  *ddict
	hist   []byte
}

func (d *decoder) Reset(in io.Reader) error {
	if in == nil {
		d.in = nil
		d.err = nil
		d.ddict = nil
		d.dict = nil
		d.opts.ddict = nil
		d.opts.dict = nil
		d.hist = nil
		return nil
	}
	d.err = nil
	d.in = in
	d.dict = d.opts.dict
	d.ddict = d.opts.ddict
	d.hist = d.dict
	return nil
}

func (d *decoder) Close() {
	d.in = nil
	d.err = nil
	d.ddict = nil
	d.dict = nil
	d.opts.ddict = nil
	d.opts.dict = nil
	d.hist = nil
	decPool.Put(d)
}

func NewReader(r io.Reader, opts ...DOption) (*Decoder, error) {
	var decOpts decoderOptions
	for _, opt := range opts {
		opt(&decOpts)
	}

	d := decPool.Get().(*decoder)
	d.opts = decOpts
	err := d.Reset(r)
	if err != nil {
		return nil, err
	}

	return &Decoder{
		d:    d,
		dict: decOpts.dict,
	}, nil
}

func (d *Decoder) Close() {
	if d.d == nil {
		return
	}
	d.d.Close()
	d.d = nil
	d.dict = nil
}

func (d *Decoder) Reset(in io.Reader) error {
	if d.d == nil {
		return errors.New("nil decoder")
	}
	if in == nil {
		d.dict = nil
	}
	return d.d.Reset(in)
}

func (d *Decoder) Read(p []byte) (n int, err error) {
	if d.d == nil {
		return 0, errors.New("nil decoder")
	}
	return d.d.in.Read(p)
}

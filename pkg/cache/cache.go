package cache

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"time"

	ac "github.com/PraserX/atomic-cache"
)

type Cache struct {
	ac *ac.AtomicCache
}

func (c Cache) Get(k string, v interface{}) (interface{}, error) {
	if byteArr, err := c.ac.Get([]byte(k)); err != nil {
		if ac.ErrNotFound == err {
			return nil, nil
		}
		return nil, err
	} else {
		dec := gob.NewDecoder(bytes.NewReader(byteArr))
		if err = dec.Decode(v); err != nil {
			return nil, err
		}
	}

	return v, nil
}

func (c Cache) Set(k string, v interface{}, expire time.Duration) error {
	if true == reflect.ValueOf(v).IsNil() {
		return nil
	}

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(v); err != nil {
		return err
	}

	if err := c.ac.Set([]byte(k), buf.Bytes(), expire); err != nil {
		return err
	}

	return nil
}

func NewLargeCache() Cache {
	return Cache{ac.New(ac.OptionMaxRecords(512), ac.OptionRecordSizeLarge(65536*8), ac.OptionMaxShardsSmall(48))}
}

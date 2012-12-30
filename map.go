package msgpack

import (
	"fmt"
	"reflect"
)

func isBytes(v reflect.Value) bool {
	return v.Elem().Kind() == reflect.Slice && v.Elem().Type().Elem().Kind() == reflect.Uint8
}

func (d *Decoder) mapLen() (int, error) {
	c, err := d.R.ReadByte()
	if err != nil {
		return 0, err
	}
	if c >= fixMapLowCode && c <= fixMapHighCode {
		return int(c & fixMapMask), nil
	}
	switch c {
	case map16Code:
		n, err := d.uint16()
		return int(n), err
	case map32Code:
		n, err := d.uint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code %x decoding map length", c)
}

func (d *Decoder) decodeMapStringString() (map[string]string, error) {
	n, err := d.mapLen()
	if err != nil {
		return nil, err
	}

	m := make(map[string]string, n)
	for i := 0; i < n; i++ {
		mk, err := d.DecodeString()
		if err != nil {
			return nil, err
		}
		mv, err := d.DecodeString()
		if err != nil {
			return nil, err
		}
		m[mk] = mv
	}

	return m, nil
}

func (d *Decoder) DecodeMap() (map[interface{}]interface{}, error) {
	n, err := d.mapLen()
	if err != nil {
		return nil, err
	}

	m := make(map[interface{}]interface{}, n)
	for i := 0; i < n; i++ {
		mk, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		if b, ok := mk.([]byte); ok {
			mk = string(b)
		}

		mv, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		if b, ok := mv.([]byte); ok {
			mv = string(b)
		}

		m[mk] = mv
	}
	return m, nil
}

func (d *Decoder) mapValue(v reflect.Value) error {
	n, err := d.mapLen()
	if err != nil {
		return err
	}

	typ := v.Type()
	if v.IsNil() {
		v.Set(reflect.MakeMap(typ))
	}
	keyType := typ.Key()
	valueType := typ.Elem()

	for i := 0; i < n; i++ {
		mk := reflect.New(keyType).Elem()
		if err := d.DecodeValue(mk); err != nil {
			return err
		}

		mv := reflect.New(valueType).Elem()
		if err := d.DecodeValue(mv); err != nil {
			return err
		}

		v.SetMapIndex(mk, mv)
	}

	return nil
}
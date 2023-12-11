package fastcache

import (
	"reflect"
	"testing"

	"github.com/VictoriaMetrics/fastcache"

	es_models "github.com/zitadel/zitadel/internal/eventstore/v1/models"
	"github.com/zitadel/zitadel/internal/zerrors"
)

type TestStruct struct {
	Test string
}

func TestSet(t *testing.T) {
	type args struct {
		cache *Fastcache
		key   string
		value *TestStruct
	}
	type res struct {
		result  *TestStruct
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "set cache no err",
			args: args{
				cache: &Fastcache{cache: fastcache.New(2000)},
				key:   "KEY",
				value: &TestStruct{Test: "Test"},
			},
			res: res{
				result: &TestStruct{},
			},
		},
		{
			name: "key empty",
			args: args{
				cache: &Fastcache{cache: fastcache.New(2000)},
				key:   "",
				value: &TestStruct{Test: "Test"},
			},
			res: res{
				errFunc: zerrors.IsErrorInvalidArgument,
			},
		},
		{
			name: "set cache nil value",
			args: args{
				cache: &Fastcache{cache: fastcache.New(2000)},
				key:   "KEY",
			},
			res: res{
				errFunc: zerrors.IsErrorInvalidArgument,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.cache.Set(tt.args.key, tt.args.value)

			if tt.res.errFunc == nil && err != nil {
				t.Errorf("got wrong result should not get err: %v ", err)
			}

			if tt.res.errFunc == nil {
				tt.args.cache.Get(tt.args.key, tt.res.result)
				if tt.res.result == nil {
					t.Errorf("got wrong result should get result: %v ", err)
				}
			}
			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		event    []*es_models.Event
		cache    *Fastcache
		key      string
		setValue *TestStruct
		getValue *TestStruct
	}
	type res struct {
		result  *TestStruct
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "get cache no err",
			args: args{
				cache:    &Fastcache{cache: fastcache.New(2000)},
				key:      "KEY",
				setValue: &TestStruct{Test: "Test"},
				getValue: &TestStruct{Test: "Test"},
			},
			res: res{
				result: &TestStruct{Test: "Test"},
			},
		},
		{
			name: "get cache no key",
			args: args{
				cache:    &Fastcache{cache: fastcache.New(2000)},
				setValue: &TestStruct{Test: "Test"},
				getValue: &TestStruct{Test: "Test"},
			},
			res: res{
				errFunc: zerrors.IsErrorInvalidArgument,
			},
		},
		{
			name: "get cache no value",
			args: args{
				cache:    &Fastcache{cache: fastcache.New(2000)},
				key:      "KEY",
				setValue: &TestStruct{Test: "Test"},
			},
			res: res{
				errFunc: zerrors.IsErrorInvalidArgument,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.cache.Set("KEY", tt.args.setValue)
			if err != nil {
				t.Errorf("something went wrong")
			}

			err = tt.args.cache.Get(tt.args.key, tt.args.getValue)

			if tt.res.errFunc == nil && err != nil {
				t.Errorf("got wrong result should not get err: %v ", err)
			}

			if tt.res.errFunc == nil && !reflect.DeepEqual(tt.args.getValue, tt.res.result) {
				t.Errorf("got wrong result expected: %v actual: %v", tt.res.result, tt.args.getValue)
			}

			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		event    []*es_models.Event
		cache    *Fastcache
		key      string
		setValue *TestStruct
		getValue *TestStruct
	}
	type res struct {
		result  *TestStruct
		errFunc func(err error) bool
	}
	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			name: "delete cache no err",
			args: args{
				cache:    &Fastcache{cache: fastcache.New(2000)},
				key:      "KEY",
				setValue: &TestStruct{Test: "Test"},
			},
			res: res{},
		},
		{
			name: "get cache no key",
			args: args{
				cache:    &Fastcache{cache: fastcache.New(2000)},
				setValue: &TestStruct{Test: "Test"},
				getValue: &TestStruct{Test: "Test"},
			},
			res: res{
				errFunc: zerrors.IsErrorInvalidArgument,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.cache.Set("KEY", tt.args.setValue)
			if err != nil {
				t.Errorf("something went wrong")
			}

			err = tt.args.cache.Delete(tt.args.key)

			if tt.res.errFunc == nil && err != nil {
				t.Errorf("got wrong result should not get err: %v ", err)
			}

			if tt.res.errFunc != nil && !tt.res.errFunc(err) {
				t.Errorf("got wrong err: %v ", err)
			}
		})
	}
}

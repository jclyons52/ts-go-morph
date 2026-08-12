package collections_test

import (
	"testing"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/collections"
	"gotest.tools/v3/assert"
)

func TestSyncMapWithNil(t *testing.T) {
	t.Parallel()

	var m collections.SyncMap[string, any]

	got1, ok := m.Load("foo")
	assert.Assert(t, !ok)
	assert.Equal(t, got1, nil)

	m.Store("foo", nil)

	got2, ok := m.Load("foo")
	assert.Assert(t, ok)
	assert.Equal(t, got2, nil)

	too, loaded := m.LoadOrStore("too", nil)
	assert.Assert(t, !loaded)
	assert.Equal(t, too, nil)

	m.Range(func(k string, v any) bool {
		return true
	})
}

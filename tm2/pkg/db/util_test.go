package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Empty iterator for empty db.
func TestPrefixIteratorNoMatchNil(t *testing.T) {
	for backend := range backends {
		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			db := newTempDB(t, backend)
			itr, err := IteratePrefix(db, []byte("2"))
			require.NoError(t, err)

			checkInvalid(t, itr)
		})
	}
}

// Empty iterator for db populated after iterator created.
func TestPrefixIteratorNoMatch1(t *testing.T) {
	for backend := range backends {
		if backend == BoltDBBackend {
			t.Log("bolt does not support concurrent writes while iterating")
			continue
		}

		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			db := newTempDB(t, backend)
			itr, err := IteratePrefix(db, []byte("2"))
			require.NoError(t, err)

			db.SetSync(bz("1"), bz("value_1"))

			checkInvalid(t, itr)
		})
	}
}

// Empty iterator for prefix starting after db entry.
func TestPrefixIteratorNoMatch2(t *testing.T) {
	for backend := range backends {
		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			db := newTempDB(t, backend)
			db.SetSync(bz("3"), bz("value_3"))
			itr, err := IteratePrefix(db, []byte("4"))
			require.NoError(t, err)

			checkInvalid(t, itr)
		})
	}
}

// Iterator with single val for db with single val, starting from that val.
func TestPrefixIteratorMatch1(t *testing.T) {
	for backend := range backends {
		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			db := newTempDB(t, backend)
			db.SetSync(bz("2"), bz("value_2"))
			itr, err := IteratePrefix(db, bz("2"))
			require.NoError(t, err)

			checkValid(t, itr, true)
			checkItem(t, itr, bz("2"), bz("value_2"))
			checkNext(t, itr, false)

			// Once invalid...
			checkInvalid(t, itr)
		})
	}
}

// Iterator with prefix iterates over everything with same prefix.
func TestPrefixIteratorMatches1N(t *testing.T) {
	for backend := range backends {
		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			db := newTempDB(t, backend)

			// prefixed
			db.SetSync(bz("a/1"), bz("value_1"))
			db.SetSync(bz("a/3"), bz("value_3"))

			// not
			db.SetSync(bz("b/3"), bz("value_3"))
			db.SetSync(bz("a-3"), bz("value_3"))
			db.SetSync(bz("a.3"), bz("value_3"))
			db.SetSync(bz("abcdefg"), bz("value_3"))
			itr, err := IteratePrefix(db, bz("a/"))
			require.NoError(t, err)

			checkValid(t, itr, true)
			checkItem(t, itr, bz("a/1"), bz("value_1"))
			checkNext(t, itr, true)
			checkItem(t, itr, bz("a/3"), bz("value_3"))

			// Bad!
			checkNext(t, itr, false)

			// Once invalid...
			checkInvalid(t, itr)
		})
	}
}

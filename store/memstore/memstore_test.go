package memstore_test
import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/CanonicalLtd/candid/store"
	"github.com/CanonicalLtd/candid/meeting"
	"github.com/CanonicalLtd/candid/store/memstore"
	"github.com/CanonicalLtd/candid/store/storetest"
)

func TestKeyValueStore(t *testing.T) {
	c := qt.New(t)
	defer c.Done()
	
	storetest.TestKeyValueStore(c, func(c *qt.C) store.ProviderDataStore {
		return memstore.NewProviderDataStore()
	})
}

func TestStore(t *testing.T) {
	c := qt.New(t)
	defer c.Done()

	storetest.TestStore(c, func(c *qt.C) store.Store {
		return memstore.NewStore()
	})
}

func TestMeetingStore(t *testing.T) {
	c := qt.New(t)
	defer c.Done()

	storetest.TestMeetingStore(c, func(c *qt.C) meeting.Store {
		return memstore.NewMeetingStore()
	}, memstore.PutAtTime)
}


func TestConfigUnmarshal(t *testing.T) {
	c := qt.New(t)
	defer c.Done()

	storetest.TestUnmarshal(c, `
storage:
    type: memory
`)
}

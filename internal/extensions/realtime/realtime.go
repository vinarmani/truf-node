// package realtime is a Kwil extension that adds realtime data to streams.
// It is a Kuneiform Precompile that should be registered before kwild starts.
// It can be imported into Kuneiform to add a view action that adds realtime data to streams.
package realtime

import (
	"fmt"
	"sync"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
)

// RealtimeExtension is the main entry point for the realtime extension.
// Its fields are accessed by both Kuneiform, as well as the external TSN
// node, which is capable of adding realtime data to streams.
type RealtimeExtension struct {
	// realtimeValue maps dbids to the realtime values, if any exist.
	// When values are queried from a stream, the extension will check this
	// map to see if there are any realtime values to add. If so, it will
	// add them as the most recent values.
	realtimeValues map[string]any // this could maybe map an int, but leaving as any for now
	// valuesMu protects the realtimeValues map.
	valuesMu sync.RWMutex
}

func NewRealtimeExtension() *RealtimeExtension {
	return &RealtimeExtension{
		realtimeValues: make(map[string]any),
	}
}

// Initialize is called when the extension is registered with the node.
func (r *RealtimeExtension) Initialize(ctx *precompiles.DeploymentContext, service *common.Service, metadata map[string]string) (precompiles.Instance, error) {
	// we don't need to do anything here
	return r, nil
}

// Call is called when a realtime precompile is called.
func (r *RealtimeExtension) Call(scoper *precompiles.ProcedureContext, app *common.App, method string, inputs []any) ([]any, error) {
	r.valuesMu.RLock()
	defer r.valuesMu.RUnlock()

	// since this is non-detemrinistic, we want to be 100% certain that we are not calling
	// this in a blockchain tx, and that it is only being called in a read-only tx.
	dbAccesser, ok := app.DB.(sql.AccessModer)
	if !ok {
		// this should never error. This is a common type asserting we do internally.
		// If this error returns, then the Kwil team has introduced a bug into the core kwild code.
		return nil, fmt.Errorf("unexpected error in realtime extension.")
	}
	if dbAccesser.AccessMode() != sql.ReadOnly {
		return nil, fmt.Errorf("realtime extension can only be used in read-only txs.")
	}

	// TODO: perform queries to get the latest values
	// for now, I am assuming the results are held in "results"
	var results *sql.ResultSet

	// we will check to see if there is a realtime value set for the dbid
	// if so, we will add it to the results
	val, ok := r.realtimeValues[scoper.DBID]
	if ok {
		// assuming column 0 is date and column 1 is value
		results.Rows = append(results.Rows, []any{"latest", val})
	}

	scoper.Result = results

	return nil, nil
}

// SetValue sets the value for a dbid. It should be called externally by the TSN node when it wishes
// to add unconfirmed data to a stream.
func (r *RealtimeExtension) SetValue(dbid string, value any) {
	r.valuesMu.Lock()
	defer r.valuesMu.Unlock()

	r.realtimeValues[dbid] = value
}

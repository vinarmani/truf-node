package basestream

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/truflation/tsn-db/internal/utils"
	"github.com/truflation/tsn-db/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Index(t *testing.T) {
	ctx := context.Background()
	app := &common.App{
		DB:     mocks.NewDB(t),
		Engine: mocks.NewEngine(t),
	}

	b := &BaseStreamExt{
		table:       "price",
		dateColumn:  "date",
		valueColumn: "value",
	}

	returned, err := b.index(ctx, app, "dbid", "2024-01-01", nil)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, returned) // 200.000 * 1000

	returned, err = b.index(ctx, app, "dbid", "", nil) // this should return the latest value
	assert.NoError(t, err)
	assert.Equal(t, []utils.ValueWithDate{{Date: "2024-01-01", Value: 266666}}, returned) // 266.666 * 1000

	dateTo := "2024-01-02"
	returned, err = b.index(ctx, app, "dbid", "2024-01-01", &dateTo)

	assert.NoError(t, err)
	assert.Equal(t, []utils.ValueWithDate{{Date: "2024-01-01", Value: 200000}, {Date: "2024-01-02", Value: 400000}}, returned) // 200.000 * 1000, 400.000 * 1000
}

func Test_Value(t *testing.T) {
	ctx := context.Background()
	app := &common.App{
		DB:     mocks.NewDB(t),
		Engine: mocks.NewEngine(t),
	}
	b := &BaseStreamExt{
		table:       "price",
		dateColumn:  "date",
		valueColumn: "value",
	}

	returned, err := b.value(ctx, app, "dbid", "2024-01-01", nil)
	assert.NoError(t, err)
	assert.Equal(t, []utils.ValueWithDate{{Date: "2024-01-01", Value: 150000}}, returned) // 150.000 * 1000

	returned, err = b.value(ctx, app, "dbid", "", nil) // this should return the latest value
	assert.NoError(t, err)
	assert.Equal(t, []utils.ValueWithDate{{Date: "2024-01-01", Value: 200000}}, returned) // 200.000 * 1000

	dateTo := "2024-01-02"
	returned, err = b.value(ctx, app, "dbid", "2024-01-01", &dateTo)
	assert.NoError(t, err)
	assert.Equal(t, []utils.ValueWithDate{{Date: "2024-01-01", Value: 150000}, {Date: "2024-01-02", Value: 300000}}, returned) // 150.000 * 1000, 300.000 * 1000
}

// mockDateScalar is a helper function that creates a new actions.Result that
// returns the given value as a row and column result with a date.
func mockDateScalar(column string, arrayOfResults []utils.ValueWithDate) *sql.ResultSet {
	mockedRows := make([][]any, len(arrayOfResults))
	for i, result := range arrayOfResults {
		mockedRows[i] = []any{result.Date, result.Value}
	}
	return &sql.ResultSet{
		Columns: []string{"date", column},
		Rows:    mockedRows,
	}
}

type mockQuerier struct {
	stmts map[string]*sql.ResultSet
}

func (m *mockQuerier) Query(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error) {
	res, ok := m.stmts[stmt]
	if !ok {
		return nil, fmt.Errorf("unexpected statement: %s", stmt)
	}
	return res, nil
}

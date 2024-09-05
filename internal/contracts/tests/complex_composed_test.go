package tests

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/truflation/tsn-db/internal/contracts"
	"github.com/truflation/tsn-sdk/core/util"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/kwilteam/kwil-db/parse"
	kwilTesting "github.com/kwilteam/kwil-db/testing"
	"github.com/stretchr/testify/assert"
)

var (
	composedStreamId   = util.GenerateStreamId("complex_composed_a")
	primitiveStreamIds = []util.StreamId{
		util.GenerateStreamId("p1"),
		util.GenerateStreamId("p2"),
		util.GenerateStreamId("p3"),
	}
)

func TestComplexComposed(t *testing.T) {
	kwilTesting.RunSchemaTest(t, kwilTesting.SchemaTest{
		Name: "complex_composed_test",
		FunctionTests: []kwilTesting.TestFunc{
			WithTestSetup(testComplexComposedRecord(t)),
			WithTestSetup(testComplexComposedIndex(t)),
			WithTestSetup(testComplexComposedLatestValue(t)),
			WithTestSetup(testComplexComposedEmptyDate(t)),
			WithTestSetup(testComplexComposedIndexChange(t)),
			WithTestSetup(testComplexComposedOutOfRange(t)),
			WithTestSetup(testComplexComposedInvalidDate(t)),
		},
	})
}

func WithTestSetup(testFn func(ctx context.Context, platform *kwilTesting.Platform) error) func(ctx context.Context, platform *kwilTesting.Platform) error {
	return func(ctx context.Context, platform *kwilTesting.Platform) error {
		// platform.Deployer can't be the dummy value that is used by default
		deployerAddress := util.Unsafe_NewEthereumAddressFromString("0x0000000000000000000000000000000000000123")
		platform.Deployer = deployerAddress.Bytes()

		// Deploy the contracts here
		if err := deployContracts(ctx, platform); err != nil {
			return err
		}

		// Run the actual test function
		return testFn(ctx, platform)
	}
}

func testComplexComposedRecord(t *testing.T) func(ctx context.Context, platform *kwilTesting.Platform) error {
	return func(ctx context.Context, platform *kwilTesting.Platform) error {
		composedDBID := utils.GenerateDBID(composedStreamId.String(), platform.Deployer)

		result, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
			Procedure: "get_record",
			Dataset:   composedDBID,
			Args:      []any{"2021-01-01", "2021-01-13", nil},
			TransactionData: common.TransactionData{
				Signer: platform.Deployer,
				TxID:   platform.Txid(),
				Height: 0,
			},
		})
		if err != nil {
			return errors.Wrap(err, "error in testComplexComposedRecord")
		}

		expected := [][]any{
			{"2021-01-01", "3.000"},
			{"2021-01-02", "5.333"},
			{"2021-01-03", "6.833"},
			{"2021-01-04", "7.833"},
			{"2021-01-05", "11.333"},
			{"2021-01-06", "16.833"},
			{"2021-01-07", "18.833"},
			{"2021-01-08", "19.833"},
			{"2021-01-09", "20.833"},
			{"2021-01-10", "26.833"},
			{"2021-01-11", "29.833"},
			{"2021-01-13", "34.333"},
		}

		assert.Equal(t, expected, convertDecimalToString(result.Rows), "Complex composed record results do not match expected values")

		return nil
	}
}

func testComplexComposedIndex(t *testing.T) func(ctx context.Context, platform *kwilTesting.Platform) error {
	return func(ctx context.Context, platform *kwilTesting.Platform) error {
		composedDBID := utils.GenerateDBID(composedStreamId.String(), platform.Deployer)

		result, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
			Procedure: "get_index",
			Dataset:   composedDBID,
			Args:      []any{"2021-01-01", "2021-01-13", nil},
			TransactionData: common.TransactionData{
				Signer: platform.Deployer,
				TxID:   platform.Txid(),
				Height: 0,
			},
		})
		if err != nil {
			return errors.Wrap(err, "error in testComplexComposedIndex")
		}

		expected := [][]any{
			{"2021-01-01", "100.000"},
			{"2021-01-02", "150.000"},
			{"2021-01-03", "200.000"},
			{"2021-01-04", "225.000"},
			{"2021-01-05", "337.500"},
			{"2021-01-06", "467.500"},
			{"2021-01-07", "512.500"},
			{"2021-01-08", "532.500"},
			{"2021-01-09", "557.500"},
			{"2021-01-10", "757.500"},
			{"2021-01-11", "817.500"},
			{"2021-01-13", "967.500"},
		}

		assert.Equal(t, expected, convertDecimalToString(result.Rows), "Complex composed index results do not match expected values")

		return nil
	}
}

func testComplexComposedLatestValue(t *testing.T) func(ctx context.Context, platform *kwilTesting.Platform) error {
	return func(ctx context.Context, platform *kwilTesting.Platform) error {
		composedDBID := utils.GenerateDBID(composedStreamId.String(), platform.Deployer)

		result, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
			Procedure: "get_record",
			Dataset:   composedDBID,
			Args:      []any{nil, nil, nil},
			TransactionData: common.TransactionData{
				Signer: platform.Deployer,
				TxID:   platform.Txid(),
				Height: 0,
			},
		})
		if err != nil {
			return errors.Wrap(err, "error in testComplexComposedLatestValue")
		}

		expected := [][]any{
			{"2021-01-13", "34.333"},
		}

		assert.Equal(t, expected, convertDecimalToString(result.Rows), "Complex composed latest value does not match expected value")

		return nil
	}
}

func testComplexComposedEmptyDate(t *testing.T) func(ctx context.Context, platform *kwilTesting.Platform) error {
	return func(ctx context.Context, platform *kwilTesting.Platform) error {
		composedDBID := utils.GenerateDBID(composedStreamId.String(), platform.Deployer)

		result, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
			Procedure: "get_record",
			Dataset:   composedDBID,
			Args:      []any{"2021-01-12", "2021-01-12", nil},
			TransactionData: common.TransactionData{
				Signer: platform.Deployer,
				TxID:   platform.Txid(),
				Height: 0,
			},
		})
		if err != nil {
			return errors.Wrap(err, "error in testComplexComposedEmptyDate")
		}

		expected := [][]any{
			{"2021-01-11", "29.833"},
		}

		assert.Equal(t, expected, convertDecimalToString(result.Rows), "Complex composed empty date result does not match expected value")

		return nil
	}
}

func testComplexComposedIndexChange(t *testing.T) func(ctx context.Context, platform *kwilTesting.Platform) error {
	return func(ctx context.Context, platform *kwilTesting.Platform) error {
		composedDBID := utils.GenerateDBID(composedStreamId.String(), platform.Deployer)

		result, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
			Procedure: "get_index_change",
			Dataset:   composedDBID,
			Args:      []any{"2021-01-02", "2021-01-13", nil, 1},
			TransactionData: common.TransactionData{
				Signer: platform.Deployer,
				TxID:   platform.Txid(),
				Height: 0,
			},
		})
		if err != nil {
			return errors.Wrap(err, "error in testComplexComposedIndexChange")
		}

		// Expected values should be calculated based on the index changes
		expected := [][]any{
			{"2021-01-02", "50.000"},
			{"2021-01-03", "33.333"},
			{"2021-01-04", "12.500"},
			{"2021-01-05", "50.000"},
			{"2021-01-06", "38.519"},
			{"2021-01-07", "9.626"},
			{"2021-01-08", "3.902"},
			{"2021-01-09", "4.695"},
			{"2021-01-10", "35.874"},
			{"2021-01-11", "7.921"},
			{"2021-01-13", "18.349"},
		}

		assert.Equal(t, expected, convertDecimalToString(result.Rows), "Complex composed index change results do not match expected values")

		return nil
	}
}

func testComplexComposedOutOfRange(t *testing.T) func(ctx context.Context, platform *kwilTesting.Platform) error {
	return func(ctx context.Context, platform *kwilTesting.Platform) error {
		composedDBID := utils.GenerateDBID(composedStreamId.String(), platform.Deployer)

		result, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
			Procedure: "get_record",
			Dataset:   composedDBID,
			Args:      []any{"2020-12-31", "2021-01-14", nil},
			TransactionData: common.TransactionData{
				Signer: platform.Deployer,
				TxID:   platform.Txid(),
				Height: 0,
			},
		})
		if err != nil {
			return errors.Wrap(err, "error in testComplexComposedOutOfRange")
		}

		// We expect the first and last dates to be within our data range
		firstDate := result.Rows[0][0].(string)
		lastDate := result.Rows[len(result.Rows)-1][0].(string)

		assert.Equal(t, "2021-01-01", firstDate, "First date should be the earliest available date")
		assert.Equal(t, "2021-01-13", lastDate, "Last date should be the latest available date")

		return nil
	}
}

func testComplexComposedInvalidDate(t *testing.T) func(ctx context.Context, platform *kwilTesting.Platform) error {
	return func(ctx context.Context, platform *kwilTesting.Platform) error {
		composedDBID := utils.GenerateDBID(composedStreamId.String(), platform.Deployer)

		_, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
			Procedure: "get_record",
			Dataset:   composedDBID,
			Args:      []any{"invalid-date", "2021-01-13", nil},
			TransactionData: common.TransactionData{
				Signer: platform.Deployer,
				TxID:   platform.Txid(),
				Height: 0,
			},
		})

		assert.Error(t, err, "Expected an error for invalid date format")

		return nil
	}
}

// Helper functions

func setTaxonomy(ctx context.Context, platform *kwilTesting.Platform, dbid string) error {

	primitiveStreamStrings := []string{}
	for _, stream := range primitiveStreamIds {
		primitiveStreamStrings = append(primitiveStreamStrings, stream.String())
	}

	deployerAddress, err := util.NewEthereumAddressFromBytes(platform.Deployer)
	if err != nil {
		return errors.Wrap(err, "error creating ethereum address")
	}

	_, err = platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
		Procedure: "set_taxonomy",
		Dataset:   dbid,
		Args: []any{
			[]string{deployerAddress.Address(), deployerAddress.Address(), deployerAddress.Address()},
			primitiveStreamStrings,
			[]string{"1.000", "2.000", "3.000"},
		},
		TransactionData: common.TransactionData{
			Signer: platform.Deployer,
			TxID:   platform.Txid(),
			Height: 0,
		},
	})
	return err
}

func MustNewDecimal(value string) decimal.Decimal {
	dec, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return *dec
}

func insertPrimitiveData(ctx context.Context, platform *kwilTesting.Platform, dbid string, testData []struct {
	date  string
	value string
}) error {
	for idx, data := range testData {
		_, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
			Procedure: "insert_record",
			Dataset:   dbid,
			Args:      []any{data.date, data.value},
			TransactionData: common.TransactionData{
				Signer: platform.Deployer,
				TxID:   platform.Txid(),
				Height: int64(idx),
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func convertDecimalToString(rows [][]any) [][]any {
	result := make([][]any, len(rows))
	for i, row := range rows {
		result[i] = make([]any, len(row))
		for j, val := range row {
			if dec, ok := val.(*decimal.Decimal); ok {
				result[i][j] = dec.String()
			} else {
				result[i][j] = val
			}
		}
	}
	return result
}

func deployContracts(ctx context.Context, platform *kwilTesting.Platform) error {
	// Deploy and initialize composed stream
	composedDBID := utils.GenerateDBID(composedStreamId.String(), platform.Deployer)
	composedSchema, err := parse.Parse(contracts.ComposedStreamContent)
	if err != nil {
		return errors.Wrap(err, "error parsing composed stream content")
	}
	composedSchema.Name = composedStreamId.String()

	platform.Engine.CreateDataset(ctx, platform.DB, composedSchema, &common.TransactionData{
		Signer: platform.Deployer,
		TxID:   platform.Txid(),
		Height: 0,
	})
	if err := initializeContract(ctx, platform, composedDBID); err != nil {
		return errors.Wrap(err, "error initializing composed stream")
	}

	// Set taxonomy
	if err := setTaxonomy(ctx, platform, composedDBID); err != nil {
		return errors.Wrap(err, "error setting taxonomy for composed stream")
	}

	allTestData := []struct {
		date     string
		streamId util.StreamId
		value    string
	}{
		{"2021-01-01", primitiveStreamIds[2], "3"},

		{"2021-01-02", primitiveStreamIds[0], "4"},
		{"2021-01-02", primitiveStreamIds[1], "5"},
		{"2021-01-02", primitiveStreamIds[2], "6"},

		{"2021-01-03", primitiveStreamIds[2], "9"},

		{"2021-01-04", primitiveStreamIds[0], "10"},

		{"2021-01-05", primitiveStreamIds[0], "13"},
		{"2021-01-05", primitiveStreamIds[2], "15"},

		{"2021-01-06", primitiveStreamIds[1], "17"},
		{"2021-01-06", primitiveStreamIds[2], "18"},

		{"2021-01-07", primitiveStreamIds[0], "19"},
		{"2021-01-07", primitiveStreamIds[1], "20"},

		{"2021-01-08", primitiveStreamIds[1], "23"},

		{"2021-01-09", primitiveStreamIds[0], "25"},

		{"2021-01-10", primitiveStreamIds[2], "30"},

		{"2021-01-11", primitiveStreamIds[1], "32"},

		{"2021-01-13", primitiveStreamIds[2], "39"},
	}

	// Deploy and initialize primitive streams
	for _, stream := range primitiveStreamIds {
		dbid := utils.GenerateDBID(stream.String(), platform.Deployer)
		primitiveSchema, err := parse.Parse(contracts.PrimitiveStreamContent)
		if err != nil {
			return errors.Wrap(err, "error parsing primitive stream content")
		}
		primitiveSchema.Name = stream.String()

		// create the primitive stream
		platform.Engine.CreateDataset(ctx, platform.DB, primitiveSchema, &common.TransactionData{
			Signer: platform.Deployer,
			TxID:   platform.Txid(),
			Height: 0,
		})

		// initialize the primitive stream
		if err := initializeContract(ctx, platform, dbid); err != nil {
			return errors.Wrap(err, "error initializing primitive stream")
		}

		thisTestData := []struct {
			date  string
			value string
		}{}

		for _, data := range allTestData {
			if data.streamId == stream {
				thisTestData = append(thisTestData, struct {
					date  string
					value string
				}{
					date:  data.date,
					value: data.value,
				})
			}
		}

		if err := insertPrimitiveData(ctx, platform, dbid, thisTestData); err != nil {
			return errors.Wrap(err, "error inserting primitive data")
		}
	}

	return nil
}

package testutils

import (
	"context"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	kwilTesting "github.com/kwilteam/kwil-db/testing"
	"github.com/pkg/errors"
	"strconv"
)

func Ptr[T any](v T) *T {
	return &v
}

// ExecuteCreateStream executes the create_stream procedure
func ExecuteCreateStream(ctx context.Context, platform *kwilTesting.Platform, streamID string, streamType string, caller string) error {
	// Convert hex string to bytes for the signer
	var signerBytes []byte
	if len(caller) > 2 {
		// Remove 0x prefix if present
		if caller[:2] == "0x" {
			signerBytes = []byte(caller[2:])
		} else {
			signerBytes = []byte(caller)
		}
	}

	txContext := &common.TxContext{
		Ctx:          ctx,
		BlockContext: &common.BlockContext{Height: 0},
		Signer:       signerBytes,
		Caller:       caller,
		TxID:         platform.Txid(),
	}

	engineContext := &common.EngineContext{
		TxContext: txContext,
	}

	_, err := platform.Engine.Call(engineContext, platform.DB, "", "create_stream", []any{
		streamID,
		streamType,
	}, func(row *common.Row) error {
		return nil
	})

	return err
}

type InsertRecordInput struct {
	DateTs int     `json:"date_ts"`
	Value  float64 `json:"value"`
}

// ExecuteInsertRecord executes the create_stream procedure
func ExecuteInsertRecord(ctx context.Context, platform *kwilTesting.Platform, streamID string, input InsertRecordInput, caller string) error {
	// Convert hex string to bytes for the signer
	var signerBytes []byte
	if len(caller) > 2 {
		// Remove 0x prefix if present
		if caller[:2] == "0x" {
			signerBytes = []byte(caller[2:])
		} else {
			signerBytes = []byte(caller)
		}
	}

	txContext := &common.TxContext{
		Ctx:          ctx,
		BlockContext: &common.BlockContext{Height: 0},
		Signer:       signerBytes,
		Caller:       caller,
		TxID:         platform.Txid(),
	}

	engineContext := &common.EngineContext{
		TxContext: txContext,
	}

	// create a decimal of type NUMERIC(10, 5)

	valueStr := strconv.FormatFloat(input.Value, 'f', -1, 32)
	valueDecimal, err := types.ParseDecimalExplicit(valueStr, 36, 18)
	if err != nil {
		return errors.Wrap(err, "failed to parse value as decimal")
	}

	_, err = platform.Engine.Call(engineContext, platform.DB, "", "insert_record", []any{
		streamID,
		input.DateTs,
		valueDecimal,
	}, func(row *common.Row) error {
		return nil
	})

	return err
}

// GetTestOptions returns the common test options
func GetTestOptions() *kwilTesting.Options {
	return &kwilTesting.Options{
		UseTestContainer: true,
	}
}

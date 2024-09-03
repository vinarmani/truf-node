package benchmark

import (
	"context"
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	kwiltypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/kwilteam/kwil-db/parse"
	kwilTesting "github.com/kwilteam/kwil-db/testing"
	"github.com/truflation/tsn-db/internal/benchmark/trees"
	"github.com/truflation/tsn-db/internal/contracts"
	"github.com/truflation/tsn-sdk/core/types"
	"github.com/truflation/tsn-sdk/core/util"
)

type SetupSchemasInput struct {
	BenchmarkCase BenchmarkCase
	Tree          trees.Tree
}

// Schema setup functions
func setupSchemas(
	ctx context.Context,
	platform *kwilTesting.Platform,
	input SetupSchemasInput,
) error {
	deployerAddress := MustNewEthereumAddressFromBytes(platform.Deployer)

	for _, node := range input.Tree.Nodes {
		var schema *kwiltypes.Schema
		var err error
		if node.IsLeaf {
			schema, err = parse.Parse(contracts.PrimitiveStreamContent)
			if err != nil {
				return err
			}
		} else {
			schema, err = parse.Parse(contracts.ComposedStreamContent)
			if err != nil {
				return err
			}
		}

		schema.Name = getStreamId(node.Index).String()

		if err := createAndInitializeSchema(ctx, platform, schema); err != nil {
			return err
		}

		if err := setupSchema(ctx, platform, schema, setupSchemaInput{
			visibility: input.BenchmarkCase.Visibility,
			treeNode:   node,
			days:       380, // to be sure we have more days to calculate change index
			owner:      deployerAddress,
		}); err != nil {
			return err
		}
	}
	return nil
}

func createAndInitializeSchema(ctx context.Context, platform *kwilTesting.Platform, schema *kwiltypes.Schema) error {
	if err := platform.Engine.CreateDataset(ctx, platform.DB, schema, &common.TransactionData{
		Signer: platform.Deployer,
		TxID:   platform.Txid(),
		Height: 0,
	}); err != nil {
		return err
	}

	_, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
		Procedure: "init",
		Dataset:   utils.GenerateDBID(schema.Name, platform.Deployer),
		Args:      []any{},
		TransactionData: common.TransactionData{
			Signer: platform.Deployer,
			TxID:   platform.Txid(),
			Height: 1,
		},
	})
	return err
}

type setupSchemaInput struct {
	visibility util.VisibilityEnum
	days       int
	owner      util.EthereumAddress
	readerDbid string
	treeNode   trees.TreeNode
}

func setupSchema(ctx context.Context, platform *kwilTesting.Platform, schema *kwiltypes.Schema, input setupSchemaInput) error {
	dbid := utils.GenerateDBID(schema.Name, input.owner.Bytes())

	if input.visibility == util.PrivateVisibility {
		if err := setVisibilityAndWhitelist(ctx, platform, dbid, input.readerDbid); err != nil {
			return err
		}
	}

	// if it's a leaf, then it's a primitive stream
	if input.treeNode.IsLeaf {
		return insertRecordsForPrimitive(ctx, platform, dbid, input.days)
	}
	// if it's not a leaf, then it's a composed stream
	return setTaxonomyForComposed(ctx, platform, dbid, input)
}

func setVisibilityAndWhitelist(ctx context.Context, platform *kwilTesting.Platform, dbid string, readerDbid string) error {
	metadataToInsert := []struct {
		key     string
		val     string
		valType string
	}{
		{string(types.ComposeVisibilityKey), strconv.Itoa(int(util.PrivateVisibility)), string(types.ComposeVisibilityKey.GetType())},
		{string(types.AllowComposeStreamKey), readerDbid, string(types.AllowComposeStreamKey.GetType())},
		{string(types.ReadVisibilityKey), strconv.Itoa(int(util.PrivateVisibility)), string(types.ReadVisibilityKey.GetType())},
		{string(types.AllowReadWalletKey), readerAddress.Address(), string(types.AllowReadWalletKey.GetType())},
	}

	// generate more wallets and stream ids, to make a little more realistic result
	// they shoudln't be influencing too much, if our indexing is correct
	for _, wallet := range getMockReadWallets(1000) {
		metadataToInsert = append(metadataToInsert, struct {
			key     string
			val     string
			valType string
		}{string(types.AllowReadWalletKey), wallet.Address(), string(types.AllowReadWalletKey.GetType())})
	}

	for _, streamId := range getMockStreamIds(1000) {
		metadataToInsert = append(metadataToInsert, struct {
			key     string
			val     string
			valType string
		}{string(types.AllowComposeStreamKey), streamId.String(), string(types.AllowComposeStreamKey.GetType())})
	}

	for _, m := range metadataToInsert {
		if err := insertMetadata(ctx, platform, dbid, m.key, m.val, m.valType); err != nil {
			return err
		}
	}
	return nil
}

// getMockReadWallets generates and returns a slice of Ethereum addresses.
// The number of addresses generated is determined by the parameter `n`.
func getMockReadWallets(n int) []util.EthereumAddress {
	wallets := make([]util.EthereumAddress, 0, n)
	for i := 0; i < n; i++ {
		// Generate a 20-byte address
		addrBytes := make([]byte, 20)
		_, err := rand.Read(addrBytes)
		if err != nil {
			panic(fmt.Sprintf("failed to generate random address: %v", err))
		}

		// Convert to EthereumAddress
		addr, err := util.NewEthereumAddressFromBytes(addrBytes)
		if err != nil {
			panic(fmt.Errorf("failed to create Ethereum address: %w", err))
		}

		wallets = append(wallets, addr)
	}
	return wallets
}

// getMockStreamIds generates and returns a slice of util.StreamId.
// The number of streamIds generated is determined by the parameter `n`.
func getMockStreamIds(n int) []util.StreamId {
	var streamIds []util.StreamId
	for i := 0; i < n; i++ {
		streamIds = append(streamIds, util.GenerateStreamId(fmt.Sprintf("stream-%d", i)))
	}
	return streamIds
}

func insertMetadata(ctx context.Context, platform *kwilTesting.Platform, dbid, key, val, valType string) error {
	_, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
		Procedure: "insert_metadata",
		Dataset:   dbid,
		Args:      []any{key, val, valType},
		TransactionData: common.TransactionData{
			Signer: platform.Deployer,
			TxID:   platform.Txid(),
			Height: 0,
		},
	})
	return err
}

// insertRecordsForPrimitive inserts records for a primitive stream.
// - it generates records for the given number of days
// - it generates a random value for each record
// - it inserts the records into the stream
// - we use a bulk insert to speed up the process
func insertRecordsForPrimitive(ctx context.Context, platform *kwilTesting.Platform, dbid string, days int) error {
	fromDate := fixedDate.AddDate(0, 0, -days)
	records := generateRecords(fromDate, fixedDate)

	// Prepare the SQL statement for bulk insert
	sqlStmt := "INSERT INTO primitive_events (date_value, value, created_at) VALUES "
	var values []string

	for _, record := range records {
		values = append(values, fmt.Sprintf("('%s', %s::decimal(21,3), 0)", record[0], record[1]))
	}

	sqlStmt += strings.Join(values, ", ")

	// Execute the bulk insert
	_, err := platform.Engine.Execute(ctx, platform.DB, dbid, sqlStmt, nil)

	return err
}

// setTaxonomyForComposed sets the taxonomy for a composed stream.
// - it creates a new taxonomy item for each child stream
func setTaxonomyForComposed(ctx context.Context, platform *kwilTesting.Platform, dbid string, input setupSchemaInput) error {
	// Calculate parent and child stream IDs based on the new structure
	var taxonomy []types.TaxonomyItem
	for _, childIndex := range input.treeNode.Children {
		childStreamId := getStreamId(childIndex)
		taxonomy = append(taxonomy, types.TaxonomyItem{
			Weight: 1,
			ChildStream: types.StreamLocator{
				DataProvider: input.owner,
				StreamId:     *childStreamId,
			},
		})
	}

	var dataProvidersArg []string
	var streamIdsArg []string
	var weightsArg []int

	for _, t := range taxonomy {
		dataProvidersArg = append(dataProvidersArg, t.ChildStream.DataProvider.Address())
		streamIdsArg = append(streamIdsArg, t.ChildStream.StreamId.String())
		weightsArg = append(weightsArg, int(t.Weight))
	}

	return executeStreamProcedure(ctx, platform, dbid, "set_taxonomy",
		[]any{dataProvidersArg, streamIdsArg, weightsArg}, platform.Deployer)
}

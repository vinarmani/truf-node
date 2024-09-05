package benchmark

import (
	"math"
	"os"
	"testing"

	kwilTesting "github.com/kwilteam/kwil-db/testing"
	"github.com/pkg/errors"
	"github.com/truflation/tsn-sdk/core/util"
)

// Main benchmark test function
func TestBench(t *testing.T) {
	// set default LOG_RESULTS to true
	if os.Getenv("LOG_RESULTS") == "" {
		os.Setenv("LOG_RESULTS", "true")
	}

	// try get resultPath from env
	resultPath := os.Getenv("RESULTS_PATH")
	if resultPath == "" {
		resultPath = "./benchmark_results.csv"
	}

	// Delete the file if it exists
	if err := deleteFileIfExists(resultPath); err != nil {
		err = errors.Wrap(err, "failed to delete file if exists")
		t.Fatal(err)
	}

	// -- Setup Test Parameters --

	// shapePairs is a list of tuples, where each tuple represents a pair of qtyStreams and branchingFactor
	// qtyStreams is the number of streams in the tree
	// branchingFactor is the branching factor of the tree
	// if branchingFactor is math.MaxInt, it means the tree is flat

	shapePairs := [][]int{
		// qtyStreams, branchingFactor
		// flat trees
		{1, math.MaxInt},
		{10, math.MaxInt},
		{100, math.MaxInt},
		{150, math.MaxInt},
		{500, math.MaxInt},

		// deep trees
		{1, 1},
		{10, 1},
		{100, 1},
		{150, 1},
		// we can't go deeper than 180, for call stack size

		// 4-way trees
		{1, 4},
		{10, 4},
		{100, 4},
		{150, 4},
		{500, 4},
	}

	samples := 10

	days := []int{1, 3, 7, 30, 365}

	visibilities := []util.VisibilityEnum{util.PublicVisibility, util.PrivateVisibility}

	var functionTests []kwilTesting.TestFunc

	// create combinations of shapePairs and visibilities
	for _, qtyStreams := range shapePairs {
		for _, visibility := range visibilities {
			functionTests = append(functionTests, getBenchmarkAndSaveFn(BenchmarkCase{
				Visibility:      visibility,
				QtyStreams:      qtyStreams[0],
				BranchingFactor: qtyStreams[1],
				Samples:         samples,
				Days:            days,
				Procedures:      []ProcedureEnum{ProcedureGetRecord, ProcedureGetIndex, ProcedureGetChangeIndex},
			}, resultPath))
		}
	}

	kwilTesting.RunSchemaTest(t, kwilTesting.SchemaTest{
		Name:          "benchmark_test",
		SchemaFiles:   []string{},
		FunctionTests: functionTests,
	})
}

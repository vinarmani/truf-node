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
		// testing 1 stream only
		{1, 1},

		// flat trees = cost of adding a new stream to our composed
		{50, math.MaxInt},
		{100, math.MaxInt},
		{200, math.MaxInt},
		{400, math.MaxInt},
		{800, math.MaxInt},
		//{1500, math.MaxInt}, // this gives error: Out of shared memory

		// deep trees = cost of adding depth
		{50, 1},
		{100, 1},
		//{200, 1}, // we can't go deeper than 180, for call stack size issues

		// to get difference for stream qty on a real world situation
		{50, 8},
		{100, 8},
		{200, 8},
		{400, 8},
		{800, 8},

		// to get difference for branching factor
		{200, 2},
		{200, 4},
		// {200, 8}, // already tested above
		{200, 16},
		{200, 32},
	}

	samples := 3

	days := []int{1, 7, 30, 365}

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

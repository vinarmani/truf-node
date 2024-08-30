package benchmark

import (
	"context"
	"time"

	"github.com/kwilteam/kwil-db/core/utils"

	kwilTesting "github.com/kwilteam/kwil-db/testing"
	"github.com/truflation/tsn-db/internal/benchmark/trees"
	"github.com/truflation/tsn-sdk/core/util"
)

func runBenchmark(ctx context.Context, platform *kwilTesting.Platform, c BenchmarkCase, tree trees.Tree) ([]Result, error) {
	var results []Result

	err := setupSchemas(ctx, platform, SetupSchemasInput{
		BenchmarkCase: c,
		Tree:          tree,
	})
	if err != nil {
		return nil, err
	}

	for _, day := range c.Days {
		for _, procedure := range c.Procedures {
			result, err := runSingleTest(ctx, RunSingleTestInput{
				Platform:  platform,
				Case:      c,
				Days:      day,
				Procedure: procedure,
				Tree:      tree,
			})
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}

	return results, nil
}

type RunSingleTestInput struct {
	Platform  *kwilTesting.Platform
	Case      BenchmarkCase
	Days      int
	Procedure ProcedureEnum
	Tree      trees.Tree
}

// runSingleTest runs a single test for the given input and returns the result.
func runSingleTest(ctx context.Context, input RunSingleTestInput) (Result, error) {
	nthDbId := utils.GenerateDBID(getStreamId(0).String(), input.Platform.Deployer)
	fromDate := fixedDate.AddDate(0, 0, -input.Days).Format("2006-01-02")
	toDate := fixedDate.Format("2006-01-02")

	result := Result{
		Case:          input.Case,
		Procedure:     input.Procedure,
		DaysQueried:   input.Days,
		MaxDepth:      input.Tree.MaxDepth,
		CaseDurations: make([]time.Duration, input.Case.Samples),
	}

	for i := 0; i < input.Case.Samples; i++ {
		start := time.Now()
		args := []any{fromDate, toDate, nil}
		if input.Procedure == ProcedureGetChangeIndex {
			args = append(args, 1) // change index accept an additional arg: $days_interval
		}
		// we read using the reader address to be sure visibility is tested
		if err := executeStreamProcedure(ctx, input.Platform, nthDbId, string(input.Procedure), args, readerAddress.Bytes()); err != nil {
			return Result{}, err
		}
		result.CaseDurations[i] = time.Since(start)
	}

	return result, nil
}

type RunBenchmarkInput struct {
	ResultPath string
	Visibility util.VisibilityEnum
	QtyStreams int
	Days       []int
	Samples    int
}

func getBenchmarkAndSaveFn(benchmarkCase BenchmarkCase, resultPath string) func(ctx context.Context, platform *kwilTesting.Platform) error {
	return func(ctx context.Context, platform *kwilTesting.Platform) error {
		platform.Deployer = deployer.Bytes()

		tree := trees.NewTree(trees.NewTreeInput{
			QtyStreams:      benchmarkCase.QtyStreams,
			BranchingFactor: benchmarkCase.BranchingFactor,
		})

		results, err := runBenchmark(ctx, platform, benchmarkCase, tree)
		if err != nil {
			return err
		}

		printResults(results)

		return saveResults(results, resultPath)
	}
}

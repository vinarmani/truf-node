package benchexport

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveOrAppendToCSV(t *testing.T) {
	testData := []SavedResults{
		{Procedure: "Test1", BranchingFactor: 1, QtyStreams: 7, Days: 100, Visibility: "Public", Samples: 10, DurationMs: 100},
		{Procedure: "Test2", BranchingFactor: 2, QtyStreams: 14, Days: 200, Visibility: "Private", Samples: 10, DurationMs: 200},
	}

	tempFile, err := os.CreateTemp("", "test_csv_*.csv")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	err = SaveOrAppendToCSV(testData, tempFile.Name())
	assert.NoError(t, err)

	content, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)

	expectedContent := "procedure,branching_factor,qty_streams,days,duration_ms,visibility,samples\nTest1,1,7,100,100,Public,10\nTest2,2,14,200,200,Private,10\n"
	assert.Equal(t, expectedContent, string(content))
}

func TestLoadCSV(t *testing.T) {
	csvData := "procedure,branching_factor,qty_streams,days,duration_ms,visibility,samples\nTest1,1,7,10,100,Public,10\nTest2,2,14,20,200,Private,10\n"
	reader := bytes.NewBufferString(csvData)

	results, err := LoadCSV[SavedResults](reader)
	if err != nil {
		t.Fatalf("LoadCSV returned an error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("LoadCSV returned an empty slice")
	}

	expectedResults := []SavedResults{
		{Procedure: "Test1", BranchingFactor: 1, QtyStreams: 7, DurationMs: 100, Visibility: "Public", Days: 10, Samples: 10},
		{Procedure: "Test2", BranchingFactor: 2, QtyStreams: 14, DurationMs: 200, Visibility: "Private", Days: 20, Samples: 10},
	}

	assert.Equal(t, expectedResults, results)
}

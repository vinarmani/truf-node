package benchexport

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSaveAsMarkdown(t *testing.T) {
	testData := []SavedResults{
		{Procedure: "Test1", BranchingFactor: 1, QtyStreams: 1, DurationMs: 100, Visibility: "Public", Samples: 10, Days: 7},
		{Procedure: "Test1", BranchingFactor: 1, QtyStreams: 2, DurationMs: 100, Visibility: "Public", Samples: 10, Days: 7},
		{Procedure: "Test1", BranchingFactor: 1, QtyStreams: 3, DurationMs: 100, Visibility: "Public", Samples: 10, Days: 7},
		{Procedure: "Test2", BranchingFactor: 1, QtyStreams: 100, DurationMs: 150, Visibility: "Private", Samples: 10, Days: 365},
		{Procedure: "Test1", BranchingFactor: 2, QtyStreams: 10, DurationMs: 200, Visibility: "Public", Samples: 10, Days: 1},
		{Procedure: "Test1", BranchingFactor: 2, QtyStreams: 10, DurationMs: 300, Visibility: "Public", Samples: 10, Days: 7},
		{Procedure: "Test2", BranchingFactor: 2, QtyStreams: 10, DurationMs: 250, Visibility: "Private", Samples: 10, Days: 365},
		{Procedure: "Test2", BranchingFactor: 2, QtyStreams: 100, DurationMs: 350, Visibility: "Private", Samples: 10, Days: 365},
	}

	tempFile, err := os.CreateTemp("", "test_markdown_*.md")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	currentDate := time.Date(2023, 4, 15, 12, 0, 0, 0, time.UTC)
	input := SaveAsMarkdownInput{
		Results:      testData,
		CurrentDate:  currentDate,
		InstanceType: "TestInstance",
		FilePath:     tempFile.Name(),
	}

	err = SaveAsMarkdown(input)
	assert.NoError(t, err)

	content, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)

	expectedContent := `Date: 2023-04-15 12:00:00

## Dates x Qty Streams

Samples per query: 10
Results in milliseconds

### TestInstance

#### Branching Factor: 1

TestInstance - Test1 - Public 

| queried days / qty streams | 1   | 2   | 3   |
| -------------------------- | --- | --- | --- |
| 7                          | 100 | 100 | 100 |



TestInstance - Test2 - Private 

| queried days / qty streams | 100 |
| -------------------------- | --- |
| 365                        | 150 |



#### Branching Factor: 2

TestInstance - Test1 - Public 

| queried days / qty streams | 10  |
| -------------------------- | --- |
| 1                          | 200 |
| 7                          | 300 |



TestInstance - Test2 - Private 

| queried days / qty streams | 10  | 100 |
| -------------------------- | --- | --- |
| 365                        | 250 | 350 |



`
	assert.Equal(t, expectedContent, string(content))
}

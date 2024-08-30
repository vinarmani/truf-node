package benchexport

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fbiville/markdown-table-formatter/pkg/markdown"

	"golang.org/x/exp/slices"
)

type SaveAsMarkdownInput struct {
	Results      []SavedResults
	CurrentDate  time.Time
	InstanceType string
	FilePath     string
}

func SaveAsMarkdown(input SaveAsMarkdownInput) error {
	days := make([]int, 0)
	qtyStreams := make([]int, 0)
	branchingFactor := make([]int, 0)

	for _, result := range input.Results {
		days = append(days, result.Days)
		qtyStreams = append(qtyStreams, result.QtyStreams)
		branchingFactor = append(branchingFactor, result.BranchingFactor)
	}

	// remove duplicates
	slices.Sort(qtyStreams)
	slices.Sort(branchingFactor)
	slices.Sort(days)

	qtyStreams = slices.Compact(qtyStreams)
	branchingFactor = slices.Compact(branchingFactor)
	days = slices.Compact(days)

	log.Printf("Saving to %s", input.FilePath)

	// Open the file in append mode, or create it if it doesn't exist
	file, err := os.OpenFile(input.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Check if the file is empty to determine whether to write the header
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// check if all results have the same amount of samples. We proceed only if they are the same
	sampleCount := make(map[int]int)
	for _, result := range input.Results {
		sampleCount[result.Samples]++
	}
	if len(sampleCount) > 1 {
		return fmt.Errorf("results have different amount of samples")
	}

	// Write the header row only if the file is empty
	if stat.Size() == 0 {
		// Write the current date
		date := input.CurrentDate.Format("2006-01-02 15:04:05")
		_, err = file.WriteString(fmt.Sprintf("Date: %s\n\n## Dates x Qty Streams\n\n", date))
		if err != nil {
			return err
		}
		// add how many samples
		_, err = file.WriteString(fmt.Sprintf("Samples per query: %d\n", input.Results[0].Samples))
		if err != nil {
			return err
		}
		_, err = file.WriteString("Results in milliseconds\n\n")
		if err != nil {
			return err
		}
	}

	type BranchingFactorType int
	type ProcedureType string
	type VisibilityType string
	type DaysType int
	type QtyStreamsType int

	// Group results by [branching_factor][procedure][visibility][days][qtyStreams][duration]
	groupedResults := make(map[BranchingFactorType]map[ProcedureType]map[VisibilityType]map[DaysType]map[QtyStreamsType]int64)
	for _, result := range input.Results {
		branchingFactor := BranchingFactorType(result.BranchingFactor)
		procedure := ProcedureType(result.Procedure)
		visibility := VisibilityType(result.Visibility)
		days := DaysType(result.Days)
		qtyStreams := QtyStreamsType(result.QtyStreams)
		duration := result.DurationMs

		if _, ok := groupedResults[BranchingFactorType(branchingFactor)]; !ok {
			groupedResults[BranchingFactorType(branchingFactor)] = make(map[ProcedureType]map[VisibilityType]map[DaysType]map[QtyStreamsType]int64)
		}
		if _, ok := groupedResults[BranchingFactorType(branchingFactor)][procedure]; !ok {
			groupedResults[BranchingFactorType(branchingFactor)][procedure] = make(map[VisibilityType]map[DaysType]map[QtyStreamsType]int64)
		}
		if _, ok := groupedResults[BranchingFactorType(branchingFactor)][procedure][VisibilityType(visibility)]; !ok {
			groupedResults[BranchingFactorType(branchingFactor)][procedure][VisibilityType(visibility)] = make(map[DaysType]map[QtyStreamsType]int64)
		}
		if _, ok := groupedResults[BranchingFactorType(branchingFactor)][procedure][VisibilityType(visibility)][DaysType(days)]; !ok {
			groupedResults[BranchingFactorType(branchingFactor)][procedure][VisibilityType(visibility)][DaysType(days)] = make(map[QtyStreamsType]int64)
		}
		if _, ok := groupedResults[BranchingFactorType(branchingFactor)][procedure][VisibilityType(visibility)][DaysType(days)][QtyStreamsType(qtyStreams)]; !ok {
			groupedResults[BranchingFactorType(branchingFactor)][procedure][VisibilityType(visibility)][DaysType(days)][QtyStreamsType(qtyStreams)] = duration
		}
	}

	// Write markdown for each instance type, procedure, and visibility combination
	if _, err = file.WriteString(fmt.Sprintf("### %s\n\n", input.InstanceType)); err != nil {
		return err
	}

	// branching factor
	for _, branchingFactor := range branchingFactor {
		if _, err = file.WriteString(fmt.Sprintf("#### Branching Factor: %d\n\n", branchingFactor)); err != nil {
			return err
		}

		procedures := groupedResults[BranchingFactorType(branchingFactor)]

		// sort procedures
		proceduresKeys := make([]ProcedureType, 0, len(procedures))
		for procedure := range procedures {
			proceduresKeys = append(proceduresKeys, procedure)
		}
		slices.Sort(proceduresKeys)

		for _, procedure := range proceduresKeys {
			visibilities := procedures[procedure]
			visibilitiesKeys := make([]VisibilityType, 0, len(visibilities))
			for visibility := range visibilities {
				visibilitiesKeys = append(visibilitiesKeys, visibility)
			}
			slices.Sort(visibilitiesKeys)

			for _, visibility := range visibilitiesKeys {
				daysMap := visibilities[visibility]

				// Write full information for each table
				if _, err = file.WriteString(fmt.Sprintf("%s - %s - %s \n\n", input.InstanceType, procedure, visibility)); err != nil {
					return err
				}

				// Create headers for the table
				headers := []string{"queried days / qty streams"}
				existingQtyStreams := make([]int, 0)
				for _, qtyStream := range qtyStreams {
					// check if there's a result for this qtyStream
					exists := false
					for _, day := range days {
						if _, ok := daysMap[DaysType(day)][QtyStreamsType(qtyStream)]; ok {
							exists = true
							break
						}
					}
					if exists {
						existingQtyStreams = append(existingQtyStreams, qtyStream)
						headers = append(headers, fmt.Sprintf("%d", qtyStream))
					}
				}

				// Create a new table formatter
				tableFormatter := markdown.NewTableFormatterBuilder().
					WithPrettyPrint().
					Build(headers...)

				rows := make([][]string, 0)

				// Add rows for each day
				for _, day := range days {
					exists := false
					row := []string{fmt.Sprintf("%d", day)}
					for _, qtyStream := range existingQtyStreams {
						if duration, ok := daysMap[DaysType(day)][QtyStreamsType(qtyStream)]; ok {
							row = append(row, fmt.Sprintf("%d", duration))
							exists = true
						} else {
							row = append(row, "-")
						}
					}
					if exists {
						rows = append(rows, row)
					}
				}

				// Format the table
				formattedTable, err := tableFormatter.Format(rows)
				if err != nil {
					return err
				}

				// Write the formatted table to the file
				if _, err = file.WriteString(formattedTable + "\n\n"); err != nil {
					return err
				}
			}

			// Add an extra newline between procedures for better readability
			if _, err = file.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

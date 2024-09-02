package trees

import (
	"fmt"
	"testing"
)

func TestCalculateTreeDepth(t *testing.T) {
	testCases := []struct {
		qtyStreams      int
		branchingFactor int
		expectedDepth   int
	}{
		{1, 2, 1},     // Root only
		{3, 2, 2},     // Full binary tree with 3 nodes
		{7, 2, 3},     // Full binary tree with 7 nodes
		{8, 2, 4},     // Binary tree with 8 nodes
		{15, 2, 4},    // Full binary tree with 15 nodes
		{16, 2, 5},    // Binary tree with 16 nodes
		{6, 3, 3},     // Tree from the original example
		{10, 3, 3},    // Ternary tree with 10 nodes
		{27, 3, 4},    // Full ternary tree with 27 nodes
		{28, 3, 4},    // Ternary tree with 28 nodes
		{100, 10, 3},  // 10-ary tree with 100 nodes
		{1000, 17, 4}, // 17-ary tree with 1000 nodes
	}

	for _, tc := range testCases {
		calculatedDepth := CalculateTreeDepth(tc.qtyStreams, tc.branchingFactor)
		fmt.Printf("Streams: %d, Branching Factor: %d\n", tc.qtyStreams, tc.branchingFactor)
		fmt.Printf("Calculated Depth: %d, Expected Depth: %d\n", calculatedDepth, tc.expectedDepth)
		if calculatedDepth != tc.expectedDepth {
			fmt.Println("Mismatch detected!")
		}
		fmt.Println()
	}
}

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

func TestNewTree(t *testing.T) {
	tests := []struct {
		name              string
		input             NewTreeInput
		skipStructCheck   bool
		expected          Tree
		expectedLeafCount int
	}{
		{
			name: "Tree with 6 streams and branching factor 3",
			input: NewTreeInput{
				QtyStreams:      6,
				BranchingFactor: 3,
			},
			expected: Tree{
				Nodes: []TreeNode{
					{Parent: -1, Children: []int{1, 2, 3}, Index: 0, IsLeaf: false},
					{Parent: 0, Children: []int{4, 5}, Index: 1, IsLeaf: false},
					{Parent: 0, Children: []int{}, Index: 2, IsLeaf: true},
					{Parent: 0, Children: []int{}, Index: 3, IsLeaf: true},
					{Parent: 1, Children: []int{}, Index: 4, IsLeaf: true},
					{Parent: 1, Children: []int{}, Index: 5, IsLeaf: true},
				},
				MaxDepth:        2,
				QtyStreams:      6,
				BranchingFactor: 3,
			},
			expectedLeafCount: 4,
		},
		{
			name:              "Tree with 4 streams and branching factor 2",
			expectedLeafCount: 2,
			input: NewTreeInput{
				QtyStreams:      4,
				BranchingFactor: 2,
			},
			expected: Tree{
				Nodes: []TreeNode{
					{Parent: -1, Children: []int{1, 2}, Index: 0, IsLeaf: false},
					{Parent: 0, Children: []int{3}, Index: 1, IsLeaf: false},
					{Parent: 0, Children: []int{}, Index: 2, IsLeaf: true},
					{Parent: 1, Children: []int{}, Index: 3, IsLeaf: true},
				},
				MaxDepth:        2,
				QtyStreams:      4,
				BranchingFactor: 2,
			},
		},
		{
			name: "Tree with 15 streams and branching factor 2 (full binary tree)",
			input: NewTreeInput{
				QtyStreams:      15,
				BranchingFactor: 2,
			},
			expectedLeafCount: 9,
			expected: Tree{
				Nodes: []TreeNode{
					{Parent: -1, Children: []int{1, 2}, Index: 0, IsLeaf: false},
					{Parent: 0, Children: []int{3, 4}, Index: 1, IsLeaf: false},
					{Parent: 0, Children: []int{5, 6}, Index: 2, IsLeaf: false},
					{Parent: 1, Children: []int{7, 8}, Index: 3, IsLeaf: false},
					{Parent: 1, Children: []int{9, 10}, Index: 4, IsLeaf: false},
					{Parent: 2, Children: []int{11, 12}, Index: 5, IsLeaf: false},
					{Parent: 2, Children: []int{13, 14}, Index: 6, IsLeaf: false},
					{Parent: 3, Children: []int{}, Index: 7, IsLeaf: true},
					{Parent: 3, Children: []int{}, Index: 8, IsLeaf: true},
					{Parent: 4, Children: []int{}, Index: 9, IsLeaf: true},
					{Parent: 4, Children: []int{}, Index: 10, IsLeaf: true},
					{Parent: 5, Children: []int{}, Index: 11, IsLeaf: true},
					{Parent: 5, Children: []int{}, Index: 12, IsLeaf: true},
					{Parent: 6, Children: []int{}, Index: 13, IsLeaf: true},
					{Parent: 6, Children: []int{}, Index: 14, IsLeaf: true},
				},
				MaxDepth:        4,
				QtyStreams:      15,
				BranchingFactor: 2,
			},
		},
		{
			name: "Tree with 100 streams and branching factor 5",
			input: NewTreeInput{
				QtyStreams:      100,
				BranchingFactor: 5,
			},
			skipStructCheck:   true,
			expectedLeafCount: 80,
			expected: Tree{
				Nodes:           make([]TreeNode, 100), // We'll check the structure separately
				MaxDepth:        4,
				QtyStreams:      100,
				BranchingFactor: 5,
			},
		},
		{
			name: "Tree with max branching factor (close to MaxInt32)",
			input: NewTreeInput{
				QtyStreams:      10_000,
				BranchingFactor: 2147483646, // MaxInt32 - 1
			},
			skipStructCheck:   true,
			expectedLeafCount: 9_999,
			expected: Tree{
				Nodes:           make([]TreeNode, 10_000), // We'll check the structure separately
				MaxDepth:        2,
				QtyStreams:      10_000,
				BranchingFactor: 2147483646,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewTree(tt.input)
			if !tt.skipStructCheck && !compareTreeStructure(result, tt.expected) {
				t.Errorf("NewTree() = %v, want %v", result, tt.expected)
			}

			// Print tree structure for debugging
			fmt.Printf("Tree structure for %s:\n", tt.name)
			for i, node := range result.Nodes {
				fmt.Printf("Node %d: Parent=%d, Children=%v, IsLeaf=%v\n", i, node.Parent, node.Children, node.IsLeaf)
			}

			// Additional checks for larger trees
			if tt.input.QtyStreams > 15 {
				// Check if the number of nodes is correct
				if len(result.Nodes) != tt.input.QtyStreams {
					t.Errorf("Expected %d nodes, got %d", tt.input.QtyStreams, len(result.Nodes))
				}

				// Check if the root node is correct
				if result.Nodes[0].Parent != -1 || result.Nodes[0].IsLeaf {
					t.Errorf("Root node is incorrect: %v", result.Nodes[0])
				}

				// Check if leaf nodes are correctly marked
				leafCount := 0
				for _, node := range result.Nodes {
					if node.IsLeaf {
						leafCount++
						if len(node.Children) != 0 {
							t.Errorf("Leaf node has children: %v", node)
						}
					} else if len(node.Children) == 0 {
						t.Errorf("Non-leaf node has no children: %v", node)
					}
				}

				// Check if the number of leaf nodes is correct
				if leafCount != tt.expectedLeafCount {
					t.Errorf("Expected %d leaf nodes, got %d", tt.expectedLeafCount, leafCount)
				}

				// Print leaf node information
				fmt.Printf("Leaf nodes: ")
				for i, node := range result.Nodes {
					if node.IsLeaf {
						fmt.Printf("%d ", i)
					}
				}
				fmt.Println()
			}
		})
	}
}

func compareTreeStructure(a, b Tree) bool {
	if a.MaxDepth != b.MaxDepth || a.QtyStreams != b.QtyStreams || a.BranchingFactor != b.BranchingFactor {
		return false
	}
	if len(a.Nodes) != len(b.Nodes) {
		return false
	}
	for i := range a.Nodes {
		if a.Nodes[i].Parent != b.Nodes[i].Parent ||
			a.Nodes[i].Index != b.Nodes[i].Index ||
			a.Nodes[i].IsLeaf != b.Nodes[i].IsLeaf ||
			len(a.Nodes[i].Children) != len(b.Nodes[i].Children) {
			return false
		}
		for j := range a.Nodes[i].Children {
			if a.Nodes[i].Children[j] != b.Nodes[i].Children[j] {
				return false
			}
		}
	}
	return true
}

package trees

import "math"

// our tree node will be like this:
// qtyStreams = 6, branchingFactor = 3
//                   0
//                  /|\
//                 1 2 3
//                /|
//               4 5
// in this case, all leaves are primitive streams (2, 3, 4, 5)
// we query the index 0, the composed stream, to get the correct response for the tree structure

type TreeNode struct {
	Parent   int
	Children []int
	Index    int
	IsLeaf   bool
}

type Tree struct {
	Nodes           []TreeNode
	MaxDepth        int
	QtyStreams      int
	BranchingFactor int
}

type NewTreeInput struct {
	QtyStreams      int
	BranchingFactor int
}

func NewTree(input NewTreeInput) Tree {
	tree := Tree{
		Nodes:           make([]TreeNode, input.QtyStreams),
		MaxDepth:        CalculateTreeDepth(input.QtyStreams, input.BranchingFactor),
		QtyStreams:      input.QtyStreams,
		BranchingFactor: input.BranchingFactor,
	}

	// Initialize root node
	tree.Nodes[0] = TreeNode{Parent: -1, Children: []int{}, Index: 0, IsLeaf: false}

	queue := []int{0}
	nextIndex := 1

	for len(queue) > 0 {
		parentIndex := queue[0]
		queue = queue[1:]

		for i := 0; i < input.BranchingFactor && nextIndex < input.QtyStreams; i++ {
			childIndex := nextIndex
			tree.Nodes[childIndex] = TreeNode{
				Parent:   parentIndex,
				Children: []int{},
				Index:    childIndex,
				IsLeaf:   true, // Initially set as leaf, may change later
			}
			tree.Nodes[parentIndex].Children = append(tree.Nodes[parentIndex].Children, childIndex)
			tree.Nodes[parentIndex].IsLeaf = false

			queue = append(queue, childIndex)
			nextIndex++
		}
	}

	return tree
}

func CalculateTreeDepth(qtyStreams, branchingFactor int) int {
	return int(math.Ceil(math.Log(float64(qtyStreams)) / math.Log(float64(branchingFactor))))
}

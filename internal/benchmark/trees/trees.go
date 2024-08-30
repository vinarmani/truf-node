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
	qtyStreams := input.QtyStreams
	branchingFactor := input.BranchingFactor

	tree := make([]TreeNode, qtyStreams)
	for i := 0; i < qtyStreams; i++ {
		tree[i] = TreeNode{
			Parent:   (i - 1) / branchingFactor, // -1 to make root's parent -1
			Children: []int{},
			Index:    i,
			IsLeaf:   true, // We'll correct this later
		}
	}

	for i := 0; i < qtyStreams; i++ {
		if i > 0 { // Skip the root node
			parentIndex := (i - 1) / branchingFactor
			tree[parentIndex].Children = append(tree[parentIndex].Children, i)
			tree[parentIndex].IsLeaf = false
		}
	}
	tree[0].Parent = -1 // Set root's parent to -1

	maxDepth := CalculateTreeDepth(qtyStreams, branchingFactor)

	return Tree{
		Nodes:           tree,
		MaxDepth:        maxDepth,
		QtyStreams:      qtyStreams,
		BranchingFactor: branchingFactor,
	}
}

func CalculateTreeDepth(qtyStreams, branchingFactor int) int {
	return int(math.Ceil(math.Log(float64(qtyStreams)) / math.Log(float64(branchingFactor))))
}

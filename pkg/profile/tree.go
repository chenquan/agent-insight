package profile

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/google/pprof/profile"
)

// TreeResult contains the result of a tree command.
type TreeResult struct {
	Root       *CallTreeNode
	ValueType  string
	TotalValue int64
}

// CallTreeNode represents a node in the hierarchical call tree.
type CallTreeNode struct {
	Name        string
	Flat        int64
	FlatPercent float64
	Cum         int64
	CumPercent  float64
	Children    []*CallTreeNode
}

// TreeConfig contains configuration for the tree command.
type TreeConfig struct {
	FocusPattern  string
	IgnorePattern string
	Depth         int
	TopN          int
	SortByCum     bool
	ValueType     *ValueTypeConfig
}

// Tree builds a hierarchical call tree from profile samples.
func Tree(p *profile.Profile, config TreeConfig) (*TreeResult, error) {
	if p == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	if config.ValueType == nil {
		metadata := extractMetadata(p)
		config.ValueType = selectDefaultValueType(p, metadata.Type)
	}

	if config.Depth <= 0 {
		config.Depth = 5
	}
	if config.TopN <= 0 {
		config.TopN = 10
	}

	var focusRegex, ignoreRegex *regexp.Regexp
	var err error

	if config.FocusPattern != "" {
		focusRegex, err = regexp.Compile(config.FocusPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid focus pattern: %w", err)
		}
	}

	if config.IgnorePattern != "" {
		ignoreRegex, err = regexp.Compile(config.IgnorePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid ignore pattern: %w", err)
		}
	}

	valueIndex := config.ValueType.Index

	// Build tree from samples
	root := &CallTreeNode{Name: "root"}

	// Calculate total value
	var totalValue int64
	for _, sample := range p.Sample {
		if valueIndex < len(sample.Value) {
			totalValue += sample.Value[valueIndex]
		}
	}

	for _, sample := range p.Sample {
		if valueIndex >= len(sample.Value) || len(sample.Location) == 0 {
			continue
		}

		value := sample.Value[valueIndex]

		// Build frame names from root to leaf (reverse of sample.Location)
		frames := make([]string, len(sample.Location))
		for i, loc := range sample.Location {
			frames[len(sample.Location)-1-i] = getLocationDisplayName(loc)
		}

		// Apply filters
		if ignoreRegex != nil {
			skip := false
			for _, name := range frames {
				if ignoreRegex.MatchString(name) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		if focusRegex != nil {
			matched := false
			for _, name := range frames {
				if focusRegex.MatchString(name) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Insert into tree
		node := root
		node.Cum += value

		for i, name := range frames {
			child := findOrCreateChild(node, name)
			child.Cum += value

			// Leaf frame gets flat value
			if i == len(frames)-1 {
				child.Flat += value
			}

			node = child
		}
	}

	// Sort and prune tree recursively
	pruneTree(root, config.TopN, config.SortByCum)

	// Calculate percentages
	calculateTreePercentages(root, totalValue)

	return &TreeResult{
		Root:       root,
		ValueType:  config.ValueType.Name + "/" + config.ValueType.Unit,
		TotalValue: totalValue,
	}, nil
}

// VisibleChildren returns the children of the root that represent real frames.
func (r *TreeResult) VisibleChildren() []*CallTreeNode {
	if r.Root == nil {
		return nil
	}
	return r.Root.Children
}

func findOrCreateChild(parent *CallTreeNode, name string) *CallTreeNode {
	for _, child := range parent.Children {
		if child.Name == name {
			return child
		}
	}
	child := &CallTreeNode{Name: name}
	parent.Children = append(parent.Children, child)
	return child
}

func pruneTree(node *CallTreeNode, topN int, sortByCum bool) {
	// Sort children
	sort.Slice(node.Children, func(i, j int) bool {
		if sortByCum {
			return node.Children[i].Cum > node.Children[j].Cum
		}
		return node.Children[i].Flat > node.Children[j].Flat
	})

	// Limit children count
	if topN > 0 && len(node.Children) > topN {
		node.Children = node.Children[:topN]
	}

	// Recurse
	for _, child := range node.Children {
		pruneTree(child, topN, sortByCum)
	}
}

func calculateTreePercentages(node *CallTreeNode, total int64) {
	if total > 0 {
		node.FlatPercent = float64(node.Flat) / float64(total) * 100
		node.CumPercent = float64(node.Cum) / float64(total) * 100
	}
	for _, child := range node.Children {
		calculateTreePercentages(child, total)
	}
}

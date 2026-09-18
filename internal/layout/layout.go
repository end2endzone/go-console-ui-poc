package layout

// Direction is the axis along which a container node arranges its children.
type Direction int

const (
	// RowDir arranges children left-to-right, splitting available width.
	RowDir Direction = iota
	// ColDir arranges children top-to-bottom, splitting available height.
	ColDir
)

// SizeSpec controls how a node's size is computed along its parent's split axis.
// Leave everything to zerores for "just grow evenly with weight 1".
type SizeSpec struct {
	Fixed int // exact size on the split axis; 0 = not fixed, use Grow instead
	Min   int // clamp: minimum size on the split axis; 0 = no minimum
	Max   int // clamp: maximum size on the split axis; 0 = no maximum
	Grow  int // weight for sharing leftover space with sibling growers. 0 defaults to 1
}

// Node is one element of the layout tree: either a leaf (a panel designed to render into) or a parent/container (Row/Col) with children.
type Node struct {
	Name     string   // required on leaves you want to look up after Resolve
	Size     SizeSpec // sizing on the parent's split axis; ignored on the root
	Dir      Direction
	Children []*Node
}

// Rect is a resolved leaf's position and size in terminal cells.
type Rect struct {
	X, Y, W, H int
}

// Row creates a container whose children are arranged side-by-side, splitting the width.
// size controls how wide this node is relative to its own siblings (ignored if this is the tree root).
func Row(size SizeSpec, children ...*Node) *Node {
	return &Node{Size: size, Dir: RowDir, Children: children}
}

// Col creates a container whose children are stacked top-to-bottom, splitting the height.
// size controls how tall this node is relative to its own siblings (ignored if this is the tree root).
func Col(size SizeSpec, children ...*Node) *Node {
	return &Node{Size: size, Dir: ColDir, Children: children}
}

// Leaf creates a panel.
// name must be unique across the tree if you intend to use this name for look ups in the final layout.Rect map.
func Leaf(name string, size SizeSpec) *Node {
	return &Node{Name: name, Size: size}
}

// Resolve walks the tree and computes the Rect of every named node for the given terminal width/height.
// Call this once per tea.WindowSizeMsg.
func Resolve(root *Node, width, height int) map[string]Rect {
	out := make(map[string]Rect)
	resolve(root, 0, 0, width, height, out)
	return out
}

func resolve(n *Node, x, y, w, h int, out map[string]Rect) {
	if n.Name != "" {
		out[n.Name] = Rect{X: x, Y: y, W: w, H: h}
	}
	if len(n.Children) == 0 {
		return
	}
	switch n.Dir {
	case RowDir:
		widths := distribute(n.Children, w)
		cx := x
		for i, c := range n.Children {
			resolve(c, cx, y, widths[i], h, out)
			cx += widths[i]
		}
	case ColDir:
		heights := distribute(n.Children, h)
		cy := y
		for i, c := range n.Children {
			resolve(c, x, cy, w, heights[i], out)
			cy += heights[i]
		}
	}
}

// distribute splits total dimension among children's.
// It processes SizeSpecs fixed sizes first (clamped).
// Then the remaining space by Grow weight (also clamped).
// Finally, the last growing sibling absorbing any rounding remainder.
func distribute(children []*Node, total int) []int {
	sizes := make([]int, len(children))
	remaining := total

	var growIdxs []int
	growTotal := 0
	for i, c := range children {
		if c.Size.Fixed > 0 {
			sz := clamp(c.Size.Fixed, c.Size.Min, c.Size.Max)
			sizes[i] = sz
			remaining -= sz
		} else {
			g := c.Size.Grow
			if g <= 0 {
				g = 1
			}
			growTotal += g
			growIdxs = append(growIdxs, i)
		}
	}
	if remaining < 0 {
		remaining = 0
	}

	distributed := 0
	for j, i := range growIdxs {
		c := children[i]
		g := c.Size.Grow
		if g <= 0 {
			g = 1
		}
		var sz int
		if j == len(growIdxs)-1 {
			sz = remaining - distributed // last grower takes the remainder
		} else {
			sz = remaining * g / growTotal
		}
		sz = clamp(sz, c.Size.Min, c.Size.Max)
		sizes[i] = sz
		distributed += sz
	}
	return sizes
}

func clamp(v, min, max int) int {
	if v < 0 {
		v = 0
	}
	if min > 0 && v < min {
		v = min
	}
	if max > 0 && v > max {
		v = max
	}
	return v
}

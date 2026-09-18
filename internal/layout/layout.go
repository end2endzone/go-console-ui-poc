package layout

import (
	"github.com/charmbracelet/lipgloss"
)

// NodeType defines how a node's children are splitted: vertically or horizontally.
type NodeType int

const (
	RowNode    NodeType = iota // RowNode    arranges children left-to-right, splitting available width.
	ColumnNode                 // ColumnNode arranges children top-to-bottom, splitting available height.
)

// SizeSpec defines a node's requirements.
// It controls how a node's size is computed along its parent's split axis.
// Leave everything to default value to "just grow evenly with an equal weight".
type SizeSpec struct {
	Fixed int // exact size on the split axis. A value of 0 means unspecified.
	Min   int // clamp: minimum size on the split axis. A value of 0 means no minimum.
	Max   int // clamp: maximum size on the split axis. A value of 0 means no maximum.
	Grow  int // weight for sharing leftover space with sibling growers. A value of 0 means unspecified.
}

// Node is one element of the layout tree: either a leaf (a panel designed to render into) or a parent/container (rows/columns) with children.
type Node struct {
	Name     string   // required on leaves you want to look up after Resolve
	Size     SizeSpec // sizing on the parent's split axis; ignored on the root
	NodeType NodeType
	Children []*Node
}

// Rect is a resolved leaf's position and size in terminal cells.
type Rect struct {
	X, Y, W, H int
}

// Shrink reduce the size of a Rect on all sides at once.
// With one argument, the reduction is applied to all sides.
// With two arguments, the reduction is applied to the vertical and horizontal sides, in that order.
// With three arguments, the reduction is applied to the top side, the horizontal sides, and the bottom side, in that order.
// With four arguments, the reduction is applied clockwise starting from the top side, followed by the right side, then the bottom, and finally the left.
// With more than four arguments no reduction is applied.
func (r Rect) Shrink(n ...int) Rect {
	switch len(n) {
	case 1:
		return ShrinkRect(r, n[0])
	case 2:
		return ShrinkRect(r, n[0], n[1])
	case 3:
		return ShrinkRect(r, n[0], n[1], n[2])
	case 4:
		return ShrinkRect(r, n[0], n[1], n[2], n[3])
	default:
		return r
	}
}

// View renders the given content inside the given Rect so that the border itself matches exactly on the rect.
func (r Rect) View(style lipgloss.Style, content string) string {
	inner := ShrinkRect(r, 1) // 1 = border thickness
	return style.Width(inner.W).Height(inner.H).Render(content)
}

// Row creates a container whose children are arranged side-by-side, splitting the width.
// size controls how wide this node is relative to its own siblings (ignored if this is the tree root).
func Row(size SizeSpec, children ...*Node) *Node {
	return &Node{Size: size, NodeType: RowNode, Children: children}
}

// Col creates a container whose children are stacked top-to-bottom, splitting the height.
// size controls how tall this node is relative to its own siblings (ignored if this is the tree root).
func Col(size SizeSpec, children ...*Node) *Node {
	return &Node{Size: size, NodeType: ColumnNode, Children: children}
}

// Leaf creates a panel.
// name must be unique across the tree if you intend to use this name for look ups in the final Rect map.
func Leaf(name string, size SizeSpec) *Node {
	return &Node{Name: name, Size: size}
}

// Resolve walks the tree and computes the Rect of every named node for the given terminal width/height.
// Call this once per tea.WindowSizeMsg.
func Resolve(root *Node, width int, height int) map[string]Rect {
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
	switch n.NodeType {
	case RowNode:
		widths := distribute(n.Children, w)
		cx := x
		for i, c := range n.Children {
			resolve(c, cx, y, widths[i], h, out)
			cx += widths[i]
		}
	case ColumnNode:
		heights := distribute(n.Children, h)
		cy := y
		for i, c := range n.Children {
			resolve(c, x, cy, w, heights[i], out)
			cy += heights[i]
		}
	}
}

// distribute splits a parent node total dimension among its given children's according to each children's SizeSpecs.
// It distribute the total dimension in the following order:
// 1. Fixed values for width/height
// 2. Remaining space by Grow weight.
// 3. The last growing sibling absorbing any rounding remainder (to use up all space).
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

// Shrink reduce the size of a Rect on all sides at once.
// With one argument, the reduction is applied to all sides.
// With two arguments, the reduction is applied to the vertical and horizontal sides, in that order.
// With three arguments, the reduction is applied to the top side, the horizontal sides, and the bottom side, in that order.
// With four arguments, the reduction is applied clockwise starting from the top side, followed by the right side, then the bottom, and finally the left.
// With more than four arguments no reduction is applied.
func ShrinkRect(r Rect, n ...int) Rect {
	top := 0
	right := 0
	bottom := 0
	left := 0

	switch len(n) {
	case 1:
		top = n[0]
		bottom = n[0]
		left = n[0]
		right = n[0]
	case 2:
		top = n[0]
		right = n[1]
		bottom = n[0]
		left = n[1]
	case 3:
		top = n[0]
		left = n[1]
		right = n[1]
		bottom = n[2]
	case 4:
		top = n[0]
		right = n[1]
		bottom = n[2]
		left = n[3]
	}

	x := r.X + left
	y := r.Y + top
	w := r.W - left - right
	h := r.H - top - bottom

	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}

	return Rect{
		X: x,
		Y: y,
		W: w,
		H: h,
	}
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

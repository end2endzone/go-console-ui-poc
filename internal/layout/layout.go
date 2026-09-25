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

// SizeSpec defines a node's requirement specifications.
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

	// Render properties
	Title        string // title displayed on the border. Can be empty to render a normal border
	TitleStyle   lipgloss.Style
	Borders      Rect // dimensions of the panel. Matches the panel's borders if the panel as a bordered style.
	BordersStyle lipgloss.Style
	Content      string
}

func (n *Node) IsLeaf() bool {
	if len(n.Children) == 0 {
		return true
	}
	return false
}

func (n *Node) GetBorderRect() Rect {
	return n.Borders
}

func (n *Node) GetInnerRect() Rect {
	tmp := n.Borders
	tmp.Shrink(1)    // 1 = border thickness
	tmp.Shrink(0, 1) // 1 = left/right padding
	return tmp
}

func (n *Node) SetContent(content string) {
	n.Content = content
}

func (n *Node) SetStyle(style lipgloss.Style) {
	n.BordersStyle = style
}

// Find looks up a Node by its name iteratively. The function is non-recusrive.
// Returns the first node matching the given name. Returns nil otherwise.
func (root *Node) Find(name string) *Node {
	if root == nil {
		return nil
	}

	// Initialize a queue for BFS tracking
	queue := []*Node{root}

	for len(queue) > 0 {
		// Pop the first element from the queue
		current := queue[0]
		queue = queue[1:]

		// Check if this is the node we are looking for
		if current.Name == name {
			return current
		}

		// Add all children to the queue to be processed later
		if len(current.Children) > 0 {
			queue = append(queue, current.Children...)
		}
	}

	return nil
}

// SetTitleUsingStyle sets the node's titles as rendered text using the current style colors of the node.
func (n *Node) SetTitleUsingStyle(title string) {
	borderForegroundColor := n.BordersStyle.GetBorderTopForeground()
	borderBackgroundColor := n.BordersStyle.GetBorderTopBackground()
	style := lipgloss.NewStyle().
		Foreground(borderForegroundColor).
		Background(borderBackgroundColor)
	n.Title = style.Render(title)
}

func (n *Node) View() string {
	if n.Borders.W == 0 || n.Borders.H == 0 {
		return ""
	}

	// DEBUG
	/*inner := n.GetInnerRect()
	boxStyle := n.BordersStyle.Width(inner.W).Height(inner.H).
		MaxWidth(inner.W). // truncate anything that it too long
		MaxHeight(inner.H) // truncate anything that it too high*/

	boxStyle := n.BordersStyle.Width(n.Borders.W).Height(n.Borders.H)

	// DEBUG
	/*.MaxWidth(n.Borders.W). // truncate anything that it too long
	MaxHeight(n.Borders.H) // truncate anything that it too high*/

	// Render a normal border if no title is specified
	if n.Title == "" {
		return boxStyle.Render(n.Content)
	}

	// DEBUG
	/*longest, actualLine := debugging.GetLongestLineInText(n.Content) // DEBUG
	if longest > 6543 || actualLine == "123456789" {
		return ""
	}
	longest, actualLine = debugging.GetLongestLineInText(debugging.StripStyles(n.Content)) // DEBUG
	if longest > 6543 || actualLine == "123456789" {
		return ""
	}
	debugging.DumpRenderingWithoutStylesToFile("Node.View().txt", n.Content)*/

	// Render a border with a title otherwise
	s := RenderBorderWithTitle(boxStyle, n.Title, n.TitleStyle, n.Content)

	// DEBUG
	/*longest, actualLine = debugging.GetLongestLineInText(debugging.StripStyles(s)) // DEBUG
	if longest > 6543 || actualLine == "123456789" {
		return ""
	}*/

	return s
}

// Rect is a resolved leaf's position and size in terminal cells.
type Rect struct {
	X, Y, W, H int
}

// Shrink reduce the size of a Panel on all sides at once.
// With one argument, the reduction is applied to all sides.
// With two arguments, the reduction is applied to the vertical and horizontal sides, in that order.
// With three arguments, the reduction is applied to the top side, the horizontal sides, and the bottom side, in that order.
// With four arguments, the reduction is applied clockwise starting from the top side, followed by the right side, then the bottom, and finally the left.
// With more than four arguments no reduction is applied.
func (r *Rect) Shrink(n ...int) {
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
		x = 0
		w = 0
	}
	if h < 0 {
		y = 0
		h = 0
	}

	r.X = x
	r.Y = y
	r.W = w
	r.H = h
}

func (r *Rect) Reset() {
	r.X = 0
	r.Y = 0
	r.W = 0
	r.H = 0
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

func RowWithName(name string, size SizeSpec, children ...*Node) *Node {
	return &Node{Name: name, Size: size, NodeType: RowNode, Children: children}
}

func ColWithName(name string, size SizeSpec, children ...*Node) *Node {
	return &Node{Name: name, Size: size, NodeType: ColumnNode, Children: children}
}

// Leaf creates a panel.
// name must be unique across the tree if you intend to use this name for look ups in the final Rect map.
func Leaf(name string, size SizeSpec) *Node {
	return &Node{Name: name, Size: size}
}

// Resolve walks the tree and computes the Rect of every named node for the given terminal width/height.
// Call this once per tea.WindowSizeMsg.
func Resolve(root *Node, width int, height int) {
	resolve(root, 0, 0, width, height)
}

func resolve(n *Node, x, y, w, h int) {
	n.Borders.Reset()

	if n.IsLeaf() {
		// Node is a leaf, assign the full remaining size to this node
		n.Borders = Rect{X: x, Y: y, W: w, H: h}
	}

	switch n.NodeType {
	case RowNode:
		widths := distribute(n.Children, w)
		cx := x
		for i, c := range n.Children {
			resolve(c, cx, y, widths[i], h)
			cx += widths[i]
		}
	case ColumnNode:
		heights := distribute(n.Children, h)
		cy := y
		for i, c := range n.Children {
			resolve(c, x, cy, w, heights[i])
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

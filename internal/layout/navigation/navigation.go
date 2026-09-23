package navigation

type Model struct {
	IDs []int
	idx int // index of the current focused component id in the list
}

func (n *Model) CurrentFocusedComponent() int {
	if n.idx >= 0 && n.idx < len(n.IDs) {
		return n.IDs[n.idx]
	}
	return -1
}

func (n *Model) NextFocusedComponent() int {
	n.idx++
	if n.idx >= len(n.IDs) {
		// Wrap around. Select first
		n.idx = 0
	}
	return n.CurrentFocusedComponent()
}

func (n *Model) PreviousFocusedComponent() int {
	n.idx--
	if n.idx < 0 || n.idx >= len(n.IDs) {
		// Wrap around. Select last
		n.idx = len(n.IDs) - 1
	}
	return n.CurrentFocusedComponent()
}

func (n *Model) FindComponentIndex(value int) int {
	for i, id := range n.IDs {
		if value == id {
			return i
		}
	}
	return -1
}

func (n *Model) SetFocusedComponentByValue(value int) {
	index := n.FindComponentIndex(value)
	if index != -1 {
		n.idx = index
	}
}

func (n *Model) SetFocusedComponentByIndex(index int) {
	if index >= 0 && index < len(n.IDs) {
		n.idx = index
	}
}

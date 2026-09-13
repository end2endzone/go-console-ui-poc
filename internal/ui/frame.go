package ui

import "fmt"

// Frame holds a collection of panels and its own dimensions.
type Frame struct {
	Panels []*Panel `json:"panels"`
	Size   Size     `json:"size,omitempty"`
}

func (v *Frame) SplitPanelsVerticalyByRatio(ratios []float32) {
	if len(ratios) != len(v.Panels) {
		err := fmt.Errorf("Number of panels and split ratios do not match!")
		panic(err)
	}

	// Split the view size into the given panels
	remainingWidth := v.Size.Width
	for i, panel := range v.Panels {
		isLast := (i + 1) == len(v.Panels)

		var newSize Size

		// width
		if !isLast {
			newSize.Width = int(ratios[i] * float32(v.Size.Width))
		} else {
			newSize.Width = remainingWidth
		}
		remainingWidth -= newSize.Width

		// height
		newSize.Height = v.Size.Height

		panel.SetSize(newSize)
	}
}

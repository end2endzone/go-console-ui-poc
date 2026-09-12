package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Bounds defines spacing around an element.
type Bounds struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

// Size defines dimensions.
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Panel represents a specific UI component.
type Panel struct {
	Padding       Bounds `json:"padding,omitempty"`
	Margin        Bounds `json:"margin,omitempty"`
	BorderSize    Size   `json:"borderSize,omitempty"` // size of the panel
	Content       string `json:"content,omitempty"`
	MinimumWidth  int    `json:"minimumWidth,omitempty"`
	MinimumHeight int    `json:"minimumHeight,omitempty"`
}

// View holds a collection of panels and its own dimensions.
type View struct {
	Panels []Panel `json:"panels"`
	Size   Size    `json:"size,omitempty"`
}

func (v *View) SplitPanelsVerticalyByRatio(ratios []float32) {
	if len(ratios) != len(v.Panels) {
		err := fmt.Errorf("Number of panels and split ratios do not match!")
		panic(err)
	}

	// Split the view size into the given panels
	remainingWidth := v.Size.Width
	for i := range v.Panels {
		isLast := (i + 1) == len(v.Panels)
		panel := &v.Panels[i]

		var newSize Size

		if !isLast {
			newSize.Width = int(ratios[i] * float32(v.Size.Width))
		} else {
			newSize.Width = remainingWidth
		}
		remainingWidth -= newSize.Width
		newSize.Height = v.Size.Height

		panel.SetSize(newSize)
	}
}

func (p *Panel) Style() lipgloss.Style {
	style := lipgloss.NewStyle().
		Width(p.BorderSize.Width).
		Height(p.BorderSize.Height).
		Border(lipgloss.RoundedBorder()).
		Padding(p.Padding.Top, p.Padding.Right, p.Padding.Bottom, p.Padding.Left)

	return style
}

func (p *Panel) SetSize(size Size) {
	size.Width -= (p.Margin.Left + p.Margin.Right)
	size.Height -= (p.Margin.Top + p.Margin.Bottom)
	p.BorderSize = size
}

func (p *Panel) GetBorderSize() Size {
	return p.BorderSize
}

func (p *Panel) GetInnerSize() Size {
	size := p.BorderSize
	size.Width -= (p.Padding.Left + p.Padding.Right)
	size.Height -= (p.Padding.Top + p.Padding.Bottom)
	return size
}

func (p *Panel) GetOutterSize() Size {
	size := p.BorderSize
	size.Width -= (p.Margin.Left + p.Margin.Right)
	size.Height -= (p.Margin.Top + p.Margin.Bottom)
	return size
}

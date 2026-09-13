package ui

import "github.com/charmbracelet/lipgloss"

// Panel represents a specific UI component.
type Panel struct {
	Padding       Bounds `json:"padding,omitempty"`
	Margin        Bounds `json:"margin,omitempty"`
	BorderSize    Size   `json:"borderSize,omitempty"` // size of the panel
	Content       string `json:"content,omitempty"`
	MinimumWidth  int    `json:"minimumWidth,omitempty"`
	MinimumHeight int    `json:"minimumHeight,omitempty"`
}

func (p *Panel) Style() lipgloss.Style {
	size := p.GetStyleSize()

	style := lipgloss.NewStyle().
		Width(size.Width).
		Height(size.Height).
		Border(lipgloss.RoundedBorder()).
		Padding(p.Padding.Top, p.Padding.Right, p.Padding.Bottom, p.Padding.Left)

	return style
}

func (p *Panel) SetSize(size Size) {
	// Do not allow size < 0
	if size.Width < 0 {
		size.Width = 0
	}
	if size.Height < 0 {
		size.Height = 0
	}

	// Checks minimums
	if size.Width < p.MinimumWidth {
		size.Width = p.MinimumWidth
	}
	if size.Height < p.MinimumHeight {
		size.Height = p.MinimumHeight
	}

	p.BorderSize = size
}

func (p *Panel) GetBorderSize() Size {
	return p.BorderSize
}

// GetStyleSize returns the dimension of the box style inside the panel.
// This size is defined by the border size minus the width and height of the borders.
func (p *Panel) GetStyleSize() Size {
	size := p.BorderSize

	// Remove the width and height of the borders
	size.Width -= 2
	size.Height -= 2

	// Do not allow size < 0
	if size.Width < 0 {
		size.Width = 0
	}
	if size.Height < 0 {
		size.Height = 0
	}

	return size
}

// GetRenderSize returns the maximum dimension of the renderable area inside the panel.
// This size is defined by the border size minus the borders width/height and the inner padding.
func (p *Panel) GetRenderSize() Size {
	size := p.BorderSize

	// Remove the width and height of the borders
	size.Width -= 2
	size.Height -= 2

	// Remove padding
	size.Width -= (p.Padding.Left + p.Padding.Right)
	size.Height -= (p.Padding.Top + p.Padding.Bottom)

	// Do not allow size < 0
	if size.Width < 0 {
		size.Width = 0
	}
	if size.Height < 0 {
		size.Height = 0
	}

	return size
}

// GetOutterSize returns the total dimension of the panel.
// This dimensions is defined by the border size plus the outter margins.
func (p *Panel) GetOutterSize() Size {
	size := p.BorderSize

	size.Width += (p.Margin.Left + p.Margin.Right)
	size.Height += (p.Margin.Top + p.Margin.Bottom)

	return size
}

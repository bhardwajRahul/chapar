package widgets

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

const (
	IconPositionStart = 0
	IconPositionEnd   = 1
)

type TextField struct {
	theme *material.Theme

	textEditor widget.Editor
	textField  material.EditorStyle
	Icon       *widget.Icon

	IconPosition int

	Text        string
	Placeholder string

	size image.Point
}

func NewTextField(theme *material.Theme, text, placeholder string) *TextField {
	t := &TextField{
		theme:       theme,
		textEditor:  widget.Editor{},
		Text:        text,
		Placeholder: placeholder,
	}

	t.textEditor.SetText(text)
	t.textEditor.SingleLine = true
	return t
}

func (t *TextField) SetText(text string) {
	t.textEditor.SetText(text)
}

func (t *TextField) SetIcon(icon *widget.Icon, position int) {
	t.Icon = icon
	t.IconPosition = position
}

func (t *TextField) SetMinWidth(width int) {
	t.size.X = width
}

func (t *TextField) Layout(gtx layout.Context) layout.Dimensions {
	borderColor := t.theme.Palette.ContrastFg
	if t.textEditor.Focused() {
		borderColor = t.theme.Palette.ContrastBg
	}

	cornerRadius := unit.Dp(4)
	border := widget.Border{
		Color:        borderColor,
		Width:        unit.Dp(1),
		CornerRadius: cornerRadius,
	}

	leftPadding := unit.Dp(8)
	if t.Icon != nil && t.IconPosition == IconPositionStart {
		leftPadding = unit.Dp(0)
	}

	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if t.size.X == 0 {
			t.size.X = gtx.Constraints.Min.X
		}

		gtx.Constraints.Min = t.size
		return layout.Inset{
			Top:    4,
			Bottom: 4,
			Left:   leftPadding,
			Right:  4,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			inputLayout := layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return material.Editor(t.theme, &t.textEditor, t.Placeholder).Layout(gtx)
			})
			widgets := []layout.FlexChild{inputLayout}

			spacing := layout.SpaceBetween
			if t.Icon != nil {
				iconLayout := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.Icon.Layout(gtx, borderColor)
				})

				if t.IconPosition == IconPositionEnd {
					widgets = []layout.FlexChild{inputLayout, iconLayout}
				} else {
					widgets = []layout.FlexChild{iconLayout, inputLayout}
					spacing = layout.SpaceEnd
				}
			}

			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Spacing: spacing}.Layout(gtx, widgets...)
		})
	})
}
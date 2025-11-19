package restful

import (
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/chapar-rest/chapar/internal/domain"
	"github.com/chapar-rest/chapar/ui/chapartheme"
	"github.com/chapar-rest/chapar/ui/converter"
	"github.com/chapar-rest/chapar/ui/widgets"
)

type Headers struct {
	values            *widgets.KeyValue
	inheritedValues   *widgets.KeyValue // KeyValue widget for inherited headers
	collectionHeaders []domain.KeyValue // Headers from collection for inheritance

	onChange func(values []domain.KeyValue)
}

func NewHeaders(headers []domain.KeyValue) *Headers {
	h := &Headers{
		values: widgets.NewKeyValue(
			converter.WidgetItemsFromKeyValue(headers)...,
		),
		inheritedValues: widgets.NewKeyValue(), // Initialize empty, will be populated when collection headers are set
	}
	h.inheritedValues.SetReadOnly(true)
	return h
}

func (h *Headers) SetHeaders(headers []domain.KeyValue) {
	h.values.SetItems(converter.WidgetItemsFromKeyValue(headers))
	// Update inherited headers when local headers change (to re-check for overrides)
	h.updateInheritedHeaders()
}

func (h *Headers) SetOnChange(f func(values []domain.KeyValue)) {
	h.onChange = f

	h.values.SetOnChanged(func(items []*widgets.KeyValueItem) {
		h.onChange(converter.KeyValueFromWidgetItems(h.values.Items))
	})
}

// SetCollectionHeaders sets the headers from the collection for inheritance
func (h *Headers) SetCollectionHeaders(headers []domain.KeyValue) {
	h.collectionHeaders = headers
	// Update inherited values widget with headers that don't have local overrides
	h.updateInheritedHeaders()
}

// updateInheritedHeaders updates the inherited headers widget, filtering out headers that have local overrides
func (h *Headers) updateInheritedHeaders() {
	// Filter out inherited headers that have local overrides
	inheritedHeaders := make([]domain.KeyValue, 0)
	for _, header := range h.collectionHeaders {
		if !h.hasLocalHeaderWithKey(header.Key) {
			inheritedHeaders = append(inheritedHeaders, header)
		}
	}

	// Update the inherited values widget
	h.inheritedValues.SetItems(converter.WidgetItemsFromKeyValue(inheritedHeaders))
}

// hasLocalHeaderWithKey checks if a local header exists with the given key
func (h *Headers) hasLocalHeaderWithKey(key string) bool {
	for _, item := range h.values.Items {
		if strings.EqualFold(item.Key, key) {
			return true
		}
	}
	return false
}

func (h *Headers) Layout(gtx layout.Context, theme *chapartheme.Theme) layout.Dimensions {
	inset := layout.Inset{Top: unit.Dp(15), Right: unit.Dp(10)}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Start,
		}.Layout(gtx,
			// Local headers
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return h.values.WithAddLayout(gtx, "Headers", "", theme)
			}),
			// Inherited headers section
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(h.collectionHeaders) == 0 || len(h.inheritedValues.Items) == 0 {
					return layout.Dimensions{}
				}

				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Label(theme.Material(), unit.Sp(14), "Inherited from Collection")
						label.Color = theme.TextColor
						label.Font.Weight = font.Medium
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return h.inheritedValues.Layout(gtx, theme)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Label(theme.Material(), unit.Sp(12), "Note: Local headers have higher priority than inherited headers")
						label.Color = theme.TextColor
						label.MaxLines = 2
						return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, label.Layout)
					}),
				)
			}),
		)
	})
}

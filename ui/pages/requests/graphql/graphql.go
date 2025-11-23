package graphql

import (
	"gioui.org/layout"
	"gioui.org/unit"
	giox "gioui.org/x/component"

	"github.com/chapar-rest/chapar/internal/domain"
	"github.com/chapar-rest/chapar/internal/prefs"
	"github.com/chapar-rest/chapar/ui/chapartheme"
	"github.com/chapar-rest/chapar/ui/converter"
	"github.com/chapar-rest/chapar/ui/explorer"
	"github.com/chapar-rest/chapar/ui/pages/requests/component"
	"github.com/chapar-rest/chapar/ui/widgets"
)

type GraphQL struct {
	Prompt *widgets.Prompt

	Req *domain.Request

	Breadcrumb *component.Breadcrumb
	AddressBar *AddressBar
	Actions    *component.Actions

	Request  *Request
	Response *Response

	split widgets.SplitView

	onSave        func(id string)
	onDataChanged func(id string, data any)
	onSubmit      func(id string)
}

func (g *GraphQL) SetOnTitleChanged(f func(title string)) {
	g.Breadcrumb.SetOnTitleChanged(f)
}

func (g *GraphQL) SetTitle(title string) {
	g.Breadcrumb.SetTitle(title)
}

func (g *GraphQL) SetDataChanged(changed bool) {
	g.Actions.IsDataChanged = changed
}

func New(req *domain.Request, theme *chapartheme.Theme, explorer *explorer.Explorer) *GraphQL {
	splitAxis := layout.Vertical
	if prefs.GetGlobalConfig().Spec.General.UseHorizontalSplit {
		splitAxis = layout.Horizontal
	}

	g := &GraphQL{
		Req:        req,
		Prompt:     widgets.NewPrompt("", "", ""),
		Breadcrumb: component.NewBreadcrumb(req.MetaData.ID, req.CollectionName, "GraphQL", req.MetaData.Name),
		split: widgets.SplitView{
			Resize: giox.Resize{
				Ratio: 0.5,
				Axis:  splitAxis,
			},
			BarWidth: unit.Dp(2),
		},
		AddressBar: NewAddressBar(req.Spec.GraphQL.URL),
		Actions:    component.NewActions(true),
		Request:    NewRequest(req, explorer, theme),
		Response:   NewResponse(theme),
	}

	g.setupHooks()

	return g
}

func (g *GraphQL) setupHooks() {
	g.AddressBar.SetOnURLChanged(func(url string) {
		clone := g.Req.Clone()
		clone.Spec.GraphQL.URL = url
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.URL = url
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	g.AddressBar.SetOnSubmit(func() {
		g.onSubmit(g.Req.MetaData.ID)
	})

	g.Request.Query.SetOnChanged(func(data string) {
		clone := g.Req.Clone()
		clone.Spec.GraphQL.Query = data
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.Query = data
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	g.Request.Variables.SetOnChanged(func(data string) {
		clone := g.Req.Clone()
		clone.Spec.GraphQL.Variables = data
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.Variables = data
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	g.Request.Headers.SetOnChanged(func(items []*widgets.KeyValueItem) {
		data := converter.KeyValueFromWidgetItems(items)
		clone := g.Req.Clone()
		clone.Spec.GraphQL.Headers = data
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.Headers = data
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	g.Request.Auth.SetOnChange(func(auth domain.Auth) {
		clone := g.Req.Clone()
		clone.Spec.GraphQL.Auth = auth
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.Auth = auth
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	g.Request.PreRequest.SetOnDropDownChanged(func(selected string) {
		clone := g.Req.Clone()
		clone.Spec.GraphQL.PreRequest.Type = selected
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.PreRequest.Type = selected
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	g.Request.PreRequest.SetOnScriptChanged(func(code string) {
		clone := g.Req.Clone()
		clone.Spec.GraphQL.PreRequest.Script = code
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.PreRequest.Script = code
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	g.Request.PostRequest.SetOnDropDownChanged(func(selected string) {
		clone := g.Req.Clone()
		clone.Spec.GraphQL.PostRequest.Type = selected
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.PostRequest.Type = selected
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	g.Request.PostRequest.SetOnScriptChanged(func(code string) {
		clone := g.Req.Clone()
		clone.Spec.GraphQL.PostRequest.Script = code
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.PostRequest.Script = code
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	g.Request.VariablesList.SetOnChanged(func(items []domain.Variable) {
		clone := g.Req.Clone()
		clone.Spec.GraphQL.VariablesList = items
		// Update g.Req to reflect the change for UI consistency
		g.Req.Spec.GraphQL.VariablesList = items
		g.onDataChanged(g.Req.MetaData.ID, clone)
	})

	prefs.AddGlobalConfigChangeListener(func(old, updated domain.GlobalConfig) {
		isChanged := old.Spec.General.UseHorizontalSplit != updated.Spec.General.UseHorizontalSplit
		if isChanged {
			if updated.Spec.General.UseHorizontalSplit {
				g.split.Axis = layout.Horizontal
			} else {
				g.split.Axis = layout.Vertical
			}
		}
	})
}

func (g *GraphQL) SetOnRequestTabChange(f func(id, tab string)) {
	g.Request.OnTabChange = func(title string) {
		f(g.Req.MetaData.ID, title)
	}
}

func (g *GraphQL) SetPostRequestSetValues(set domain.PostRequestSet) {
	g.Request.PostRequest.SetPostRequestSetValues(set)
}

func (g *GraphQL) SetOnPostRequestSetChanged(f func(id string, statusCode int, item, from, fromKey string)) {
	g.Request.PostRequest.SetOnPostRequestSetChanged(func(statusCode int, item, from, fromKey string) {
		f(g.Req.MetaData.ID, statusCode, item, from, fromKey)
	})
}

func (g *GraphQL) SetPreRequestCollections(collections []*domain.Collection, selectedID string) {
	g.Request.PreRequest.SetCollections(collections, selectedID)
}

func (g *GraphQL) SetPreRequestRequests(requests []*domain.Request, selectedID string) {
	g.Request.PreRequest.SetRequests(requests, selectedID)
}

func (g *GraphQL) SetOnSetOnTriggerRequestChanged(f func(id, collectionID, requestID string)) {
	g.Request.PreRequest.SetOnTriggerRequestChanged(func(collectionID, requestID string) {
		f(g.Req.MetaData.ID, collectionID, requestID)
	})
}

func (g *GraphQL) SetOnDataChanged(f func(id string, data any)) {
	g.onDataChanged = f
}

func (g *GraphQL) SetOnSubmit(f func(id string)) {
	g.onSubmit = f
}

func (g *GraphQL) SetOnCopyResponse(f func(gtx layout.Context, dataType, data string)) {
	g.Response.SetOnCopyResponse(f)
}

func (g *GraphQL) SetGraphQLResponse(detail domain.GraphQLResponseDetail) {
	g.Request.VariablesList.SetResponseDetail(&domain.ResponseDetail{GraphQL: &detail})
	g.Response.SetResponse(detail.Response)
	g.Response.SetHeaders(nil, detail.ResponseHeaders)
	g.Response.SetError(detail.Error)
	g.Response.SetStatusParams(detail.StatusCode, detail.Duration, detail.Size)
}

func (g *GraphQL) GetGraphQLResponse() *domain.GraphQLResponseDetail {
	return &domain.GraphQLResponseDetail{
		Response: g.Response.GetResponse(),
	}
}

func (g *GraphQL) SetPostRequestSetPreview(preview string) {
	g.Request.PostRequest.SetPreview(preview)
}

func (g *GraphQL) SetOnSave(f func(id string)) {
	g.onSave = f
}

func (g *GraphQL) HidePrompt() {
	g.Prompt.Hide()
}

func (g *GraphQL) ShowPrompt(title, content, modalType string, onSubmit func(selectedOption string, remember bool), options ...widgets.Option) {
	g.Prompt.Type = modalType
	g.Prompt.Title = title
	g.Prompt.Content = content
	g.Prompt.SetOptions(options...)
	g.Prompt.WithoutRememberBool()
	g.Prompt.SetOnSubmit(onSubmit)
	g.Prompt.Show()
}

func (g *GraphQL) ShowSendingRequestLoading() {
	g.Response.SetMessage("Sending request...")
}

func (g *GraphQL) HideSendingRequestLoading() {
	g.Response.SetMessage("")
}

func (g *GraphQL) SetURL(url string) {
	g.AddressBar.SetURL(url)
	g.Req.Spec.GraphQL.URL = url
}

func (g *GraphQL) SetCollection(collection *domain.Collection) {
	// GraphQL can inherit headers and auth from collection
	if collection != nil {
		// Headers inheritance would be handled similar to REST
		// For now, we'll just store the collection reference
	}
}

func (g *GraphQL) Layout(gtx layout.Context, theme *chapartheme.Theme) layout.Dimensions {
	g.Prompt.Layout(gtx, theme)

	if g.Actions.IsDataChanged && g.Actions.SaveButton.Clicked(gtx) && g.onSave != nil {
		g.onSave(g.Req.MetaData.ID)
		g.Actions.IsDataChanged = false
	}

	return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return g.Prompt.Layout(gtx, theme)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Bottom: unit.Dp(15), Top: unit.Dp(5)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return g.Breadcrumb.Layout(gtx, theme)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return g.Actions.Layout(gtx, theme)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return g.AddressBar.Layout(gtx, theme)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return g.split.Layout(gtx, theme,
					func(gtx layout.Context) layout.Dimensions {
						return g.Request.Layout(gtx, theme)
					},
					func(gtx layout.Context) layout.Dimensions {
						return g.Response.Layout(gtx, theme)
					},
				)
			}),
		)
	})
}

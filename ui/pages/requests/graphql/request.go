package graphql

import (
	"gioui.org/layout"
	"gioui.org/unit"

	"github.com/chapar-rest/chapar/internal/domain"
	"github.com/chapar-rest/chapar/ui/chapartheme"
	"github.com/chapar-rest/chapar/ui/explorer"
	"github.com/chapar-rest/chapar/ui/pages/requests/component"
	"github.com/chapar-rest/chapar/ui/widgets"
	"github.com/chapar-rest/chapar/ui/widgets/codeeditor"
)

type Request struct {
	Tabs *widgets.Tabs

	PreRequest  *component.PrePostRequest
	PostRequest *component.PrePostRequest

	Query         *codeeditor.CodeEditor
	Variables     *codeeditor.CodeEditor
	Headers       *component.Headers
	VariablesList *component.Variables
	Auth          *component.Auth

	currentTab  string
	OnTabChange func(title string)
}

func NewRequest(req *domain.Request, explorer *explorer.Explorer, theme *chapartheme.Theme) *Request {
	postRequestDropDown := widgets.NewDropDown(
		widgets.NewDropDownOption("From Response").WithValue(domain.PostRequestSetFromResponseBody),
		widgets.NewDropDownOption("From Header").WithValue(domain.PostRequestSetFromResponseHeader),
		widgets.NewDropDownOption("From Cookie").WithValue(domain.PostRequestSetFromResponseCookie),
	)

	r := &Request{
		Tabs: widgets.NewTabs([]*widgets.Tab{
			{Title: "Query"},
			{Title: "Variables"},
			{Title: "Headers"},
			{Title: "Auth"},
			{Title: "Pre Request"},
			{Title: "Post Request"},
		}, nil),
		PreRequest: component.NewPrePostRequest([]component.Option{
			{Title: "None", Value: domain.PrePostTypeNone},
			{Title: "Trigger request", Value: domain.PrePostTypeTriggerRequest, Type: component.TypeTriggerRequest, Hint: "Trigger another request"},
			{Title: "Python", Value: domain.PrePostTypePython, Type: component.TypeScript, Hint: "Write your pre request python script here"},
		}, nil, theme),
		PostRequest: component.NewPrePostRequest([]component.Option{
			{Title: "None", Value: domain.PrePostTypeNone},
			{Title: "Set Environment Variable", Value: domain.PrePostTypeSetEnv, Type: component.TypeSetEnv, Hint: "Set environment variable"},
			{Title: "Python", Value: domain.PrePostTypePython, Type: component.TypeScript, Hint: "Write your post request python script here"},
		}, postRequestDropDown, theme),
		Query:         codeeditor.NewCodeEditor("", codeeditor.CodeLanguageGraphQL, theme),
		Variables:     codeeditor.NewCodeEditor("{}", codeeditor.CodeLanguageJSON, theme),
		Headers:       component.NewHeaders(nil),
		VariablesList: component.NewVariables(theme, domain.RequestTypeGraphQL),
		Auth:          component.NewAuth(domain.Auth{}, theme),
	}

	r.Variables.WithBeautifier(true)

	if req.Spec != (domain.RequestSpec{}) && req.Spec.GraphQL != nil {
		r.Query.SetCode(req.Spec.GraphQL.Query)
		r.Variables.SetCode(req.Spec.GraphQL.Variables)
		if len(req.Spec.GraphQL.Headers) > 0 {
			r.Headers.SetHeaders(req.Spec.GraphQL.Headers)
		}

		if req.Spec.GraphQL.Auth != (domain.Auth{}) {
			r.Auth = component.NewAuth(req.Spec.GraphQL.Auth, theme)
		}

		if req.Spec.GraphQL.PreRequest != (domain.PreRequest{}) {
			r.PreRequest.SetSelectedDropDown(req.Spec.GraphQL.PreRequest.Type)
			r.PreRequest.SetCode(req.Spec.GraphQL.PreRequest.Script)
		}

		if req.Spec.GraphQL.PostRequest != (domain.PostRequest{}) {
			r.PostRequest.SetSelectedDropDown(req.Spec.GraphQL.PostRequest.Type)
			r.PostRequest.SetCode(req.Spec.GraphQL.PostRequest.Script)
		}

		if req.Spec.GraphQL.PostRequest.PostRequestSet != (domain.PostRequestSet{}) {
			r.PostRequest.SetPostRequestSetValues(req.Spec.GraphQL.PostRequest.PostRequestSet)
		}

		if req.Spec.GraphQL.VariablesList != nil {
			r.VariablesList.SetValues(req.Spec.GraphQL.VariablesList)
		}
	}

	return r
}

func (r *Request) Layout(gtx layout.Context, theme *chapartheme.Theme) layout.Dimensions {
	inset := layout.Inset{Top: unit.Dp(10)}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Start,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return r.Tabs.Layout(gtx, theme)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if r.Tabs.SelectedTab().Title != r.currentTab {
					r.currentTab = r.Tabs.SelectedTab().Title
					if r.OnTabChange != nil {
						r.OnTabChange(r.currentTab)
					}
				}

				switch r.Tabs.SelectedTab().Title {
				case "Pre Request":
					return r.PreRequest.Layout(gtx, theme)
				case "Post Request":
					return r.PostRequest.Layout(gtx, theme)
				case "Query":
					return layout.Inset{Top: unit.Dp(5), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return r.Query.Layout(gtx, theme, "GraphQL Query")
					})
				case "Variables":
					return layout.Inset{Top: unit.Dp(5), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return r.Variables.Layout(gtx, theme, "Variables (JSON)")
					})
				case "Headers":
					return layout.Inset{Top: unit.Dp(5), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return r.Headers.Layout(gtx, theme)
					})
				case "Auth":
					return r.Auth.Layout(gtx, theme)
				default:
					return layout.Dimensions{}
				}
			}),
		)
	})
}

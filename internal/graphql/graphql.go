package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"

	"github.com/chapar-rest/chapar/internal/domain"
	"github.com/chapar-rest/chapar/internal/prefs"
	"github.com/chapar-rest/chapar/internal/state"
	"github.com/chapar-rest/chapar/internal/util"
	"github.com/chapar-rest/chapar/internal/variables"
	"github.com/chapar-rest/chapar/version"
)

type Response struct {
	StatusCode      int
	ResponseHeaders map[string]string
	RequestHeaders  map[string]string
	Body            []byte
	TimePassed      time.Duration
	IsJSON          bool
	JSON            string
}

type Service struct {
	requests     *state.Requests
	environments *state.Environments
}

func New(requests *state.Requests, environments *state.Environments) *Service {
	return &Service{
		requests:     requests,
		environments: environments,
	}
}

func (s *Service) SendRequest(requestID, activeEnvironmentID string) (*Response, error) {
	req := s.requests.GetRequest(requestID)
	if req == nil {
		return nil, fmt.Errorf("request with id %s not found", requestID)
	}

	// clone the request to make sure we do not modify the original request
	r := req.Clone()

	// Merge collection headers and auth if request belongs to a collection
	if r.CollectionID != "" && r.Spec.GraphQL != nil {
		collection := s.requests.GetCollection(r.CollectionID)
		if collection != nil {
			// Merge headers: collection headers as base, request headers override
			r.Spec.GraphQL.Headers = s.mergeHeaders(collection.Spec.Headers, r.Spec.GraphQL.Headers)

			// Resolve auth: if request auth is inherit, use collection auth
			if r.Spec.GraphQL.Auth.Type == domain.AuthTypeInherit {
				r.Spec.GraphQL.Auth = collection.Spec.Auth
			}
		}
	}

	var activeEnvironment *domain.Environment
	// Get environment if provided
	if activeEnvironmentID != "" {
		activeEnvironment = s.environments.GetEnvironment(activeEnvironmentID)
		if activeEnvironment == nil {
			return nil, fmt.Errorf("environment with id %s not found", activeEnvironmentID)
		}
	}

	response, err := s.sendRequest(r.Spec.GraphQL, activeEnvironment)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// mergeHeaders merges collection headers with request headers
// Collection headers are the base, request headers override collection headers with the same key
// Only enabled headers are included
func (s *Service) mergeHeaders(collectionHeaders, requestHeaders []domain.KeyValue) []domain.KeyValue {
	// Create a map of request headers by key (case-insensitive) for quick lookup
	requestHeaderMap := make(map[string]domain.KeyValue)
	for _, h := range requestHeaders {
		if h.Enable {
			requestHeaderMap[strings.ToLower(h.Key)] = h
		}
	}

	// Start with collection headers
	merged := make([]domain.KeyValue, 0)

	// Add collection headers that don't have request overrides
	for _, ch := range collectionHeaders {
		if !ch.Enable {
			continue
		}
		keyLower := strings.ToLower(ch.Key)
		if _, hasOverride := requestHeaderMap[keyLower]; !hasOverride {
			merged = append(merged, ch)
		}
	}

	// Add all request headers (they override collection headers)
	for _, rh := range requestHeaders {
		if rh.Enable {
			merged = append(merged, rh)
		}
	}

	return merged
}

func (s *Service) sendRequest(req *domain.GraphQLRequestSpec, e *domain.Environment) (*Response, error) {
	// prepare request
	// - apply environment
	// - apply variables
	// - apply authentication (if any) is not already applied to the headers

	vars := variables.GetVariables()
	variables.ApplyToGraphQLRequest(vars, req)

	if e != nil {
		variables.ApplyToEnv(vars, &e.Spec)
		e.ApplyToGraphQLRequest(req)
	}

	// Prepare GraphQL request body
	requestBody := map[string]interface{}{
		"query": req.Query,
	}

	// Parse variables JSON string to map
	if req.Variables != "" && req.Variables != "{}" {
		var variablesMap map[string]interface{}
		if err := json.Unmarshal([]byte(req.Variables), &variablesMap); err != nil {
			return nil, fmt.Errorf("invalid variables JSON: %w", err)
		}
		requestBody["variables"] = variablesMap
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequest("POST", req.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Set Content-Type header for GraphQL
	httpReq.Header.Set("Content-Type", "application/json")

	// apply headers
	for _, h := range req.Headers {
		if !h.Enable {
			continue
		}
		httpReq.Header.Add(h.Key, h.Value)
	}

	// apply authentication
	if req.Auth != (domain.Auth{}) {
		if req.Auth.Type == domain.AuthTypeToken {
			if req.Auth.TokenAuth != nil && req.Auth.TokenAuth.Token != "" {
				httpReq.Header.Add("Authorization", "Bearer "+req.Auth.TokenAuth.Token)
			}
		}

		if req.Auth.Type == domain.AuthTypeBasic {
			if req.Auth.BasicAuth != nil && req.Auth.BasicAuth.Username != "" && req.Auth.BasicAuth.Password != "" {
				httpReq.SetBasicAuth(req.Auth.BasicAuth.Username, req.Auth.BasicAuth.Password)
			}
		}

		if req.Auth.Type == domain.AuthTypeAPIKey {
			if req.Auth.APIKeyAuth != nil && req.Auth.APIKeyAuth.Key != "" && req.Auth.APIKeyAuth.Value != "" {
				httpReq.Header.Add(req.Auth.APIKeyAuth.Key, req.Auth.APIKeyAuth.Value)
			}
		}
	}

	// send request
	globalConfig := prefs.GetGlobalConfig()

	start := time.Now()

	client := &http.Client{
		Timeout: time.Duration(globalConfig.Spec.General.RequestTimeoutSec) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:           10,
			MaxResponseHeaderBytes: int64(globalConfig.Spec.General.ResponseSizeMb * 1024 * 1024),
		},
	}

	if globalConfig.Spec.General.HTTPVersion == "http/2" {
		client.Transport = &http2.Transport{
			AllowHTTP:        true,
			MaxReadFrameSize: uint32(globalConfig.Spec.General.ResponseSizeMb * 1024 * 1024),
		}
	}

	if globalConfig.Spec.General.SendNoCacheHeader {
		httpReq.Header.Add("Cache-Control", "no-cache")
	}

	if globalConfig.Spec.General.SendChaparAgentHeader {
		httpReq.Header.Add("User-Agent", version.GetAgentName())
	}

	res, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// read body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// measure time
	elapsed := time.Since(start)

	// handle response
	response := &Response{
		StatusCode:      res.StatusCode,
		ResponseHeaders: map[string]string{},
		RequestHeaders:  map[string]string{},
		Body:            body,
		TimePassed:      elapsed,
		IsJSON:          false,
	}

	if util.IsJSON(string(body)) {
		response.IsJSON = true
		if js, err := util.PrettyJSON(body); err != nil {
			return nil, err
		} else {
			response.JSON = js
		}
	}

	// handle headers
	for k, v := range res.Header {
		response.ResponseHeaders[k] = strings.Join(v, ", ")
	}

	for k, v := range httpReq.Header {
		response.RequestHeaders[k] = strings.Join(v, ", ")
	}

	return response, nil
}

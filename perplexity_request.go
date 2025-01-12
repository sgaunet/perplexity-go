package perplexity

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

var ErrSearchDomainFilter = errors.New("search domain filter must be less than or equal to 3")
var ErrSearchRecencyFilter = errors.New("search recency filter is incompatible with images")

const (
	DefaultModel            = Llama31SonarSmall128kOnline
	DefaultTemperature      = 0.2
	DefaultTopP             = 0.9
	DefaultTopK             = 0
	DefaultMaxTokens        = 0
	DefaultPresencePenalty  = 0.0
	DefaultFrequencyPenalty = 1.0

	MaxLengthOfSearchDomainFilter = 3
)

// Message is a message object for the Perplexity API.
type Message struct {
	Role    string `json:"role" validate:"required,oneof=system user agent"`
	Content string `json:"content"`
}

// CompletionRequest is a request object for the Perplexity API.
// https://docs.perplexity.ai/api-reference/chat-completions
type CompletionRequest struct {
	Messages []Message `json:"messages" validate:"required,dive"`
	// Model: name of the model that will complete your prompt
	// supported model: https://docs.perplexity.ai/guides/model-cards
	Model string `json:"model" validate:"required,oneof=llama-3.1-sonar-small-128k-online llama-3.1-sonar-large-128k-online llama-3.1-sonar-huge-128k-online"`
	// MaxTokens: The maximum number of completion tokens returned by the API.
	// The total number of tokens requested in max_tokens plus the number of
	// prompt tokens sent in messages must not exceed the context window token limit of model requested.
	// If left unspecified, then the model will generate tokens until
	// either it reaches its stop token or the end of its context window.
	MaxTokens int `json:"max_tokens" validate:"gte=0"`
	// Temperatur: The amount of randomness in the response, valued between 0 inclusive and 2 exclusive.
	// Higher values are more random, and lower values are more deterministic.
	// Required range: 0 < x < 2
	Temperature float64 `json:"temperature" validate:"gt=0,lt=2"`
	// TopP: The nucleus sampling threshold, valued between 0 and 1 inclusive.
	// For each subsequent token, the model considers the results of the tokens with top_p probability mass.
	// We recommend either altering top_k or top_p, but not both.
	// Required range: 0 < x < 1
	TopP float64 `json:"top_p" validate:"gt=0,lt=1"`
	// SearchDomainFilter: Given a list of domains, limit the citations used by the online model
	// to URLs from the specified domains. Currently limited to only 3 domains for whitelisting and blacklisting.
	// For blacklisting add a - to the beginning of the domain string. This filter is in closed beta
	SearchDomainFilter []string `json:"search_domain_filter"`
	// ReturnImages: Determines whether or not a request to an online model
	// should return images. Images are in closed beta
	ReturnImages bool `json:"return_images"`
	// ReturnRelatedQuestions: Determines whether or not a request to an online model
	// should return related questions. Related questions are in closed beta
	ReturnRelatedQuestions bool `json:"return_related_questions"`
	// SearchRecencyFilter: Returns search results within the specified time interval - does not apply to images.
	// Values include month, week, day, hour
	SearchRecencyFilter string `json:"search_recency_filter"`
	// TopK: The number of tokens to keep for highest top-k filtering,
	// specified as an integer between 0 and 2048 inclusive.
	// If set to 0, top-k filtering is disabled.
	// We recommend either altering top_k or top_p, but not both.
	// Required range: 0 < x < 2048
	TopK int `json:"top_k" validate:"gte=0,lte=2048"`
	// Stream: Determines whether or not to incrementally stream the response
	// with server-sent events with content-type: text/event-stream
	// The client of this does not handle the stream, it is up to the user to handle the stream.
	Stream bool `json:"stream"`
	// PresencePenalty: A value between -2.0 and 2.0.
	// Positive values penalize new tokens based on whether they appear in the text so far,
	// increasing the model's likelihood to talk about new topics.
	// Incompatible with frequency_penalty
	PresencePenalty float64 `json:"presence_penalty" validate:"gte=-2,lte=2"`
	// FrequencyPenalty: A multiplicative penalty greater than 0.
	// Values greater than 1.0 penalize new tokens based on their existing frequency in the text so far,
	// decreasing the model's likelihood to repeat the same line verbatim. A value of 1.0 means no penalty.
	// Incompatible with presence_penalty
	FrequencyPenalty float64 `json:"frequency_penalty" validate:"gt=0"`
}

// DefaultCompletionRequest returns a default completion request.
func DefaultCompletionRequest() *CompletionRequest {
	DefaultCompletionRequest := CompletionRequest{
		Messages:               nil,
		Model:                  DefaultModel,
		MaxTokens:              DefaultMaxTokens,
		Temperature:            DefaultTemperature,
		TopP:                   DefaultTopP,
		SearchDomainFilter:     nil,
		ReturnImages:           false,
		ReturnRelatedQuestions: false,
		SearchRecencyFilter:    "",
		TopK:                   DefaultTopK,
		Stream:                 false,
		PresencePenalty:        DefaultPresencePenalty,
		FrequencyPenalty:       DefaultFrequencyPenalty,
	}
	return &DefaultCompletionRequest
}

// CompletionRequestOption is a functional option for the CompletionRequest.
type CompletionRequestOption func(*CompletionRequest)

// WithMessages sets the messages option.
func WithMessages(msg []Message) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.Messages = msg
	}
}

// WithModel sets the model option (overrides the default model).
func WithModel(model string) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.Model = model
	}
}

// WithModelLlama31SonarSmall128kOnline sets the model to llama-3.1-sonar-small-128k-online.
func WithModelLlama31SonarSmall128kOnline() CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.Model = Llama31SonarSmall128kOnline
	}
}

// WithModelLlama31SonarLarge128kOnline sets the model to llama-3.1-sonar-large-128k-online.
func WithModelLlama31SonarLarge128kOnline() CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.Model = Llama31SonarLarge128kOnline
	}
}

// WithModelLlama31SonarHuge128kOnline sets the model to llama-3.1-sonar-huge-128k-online.
func WithModelLlama31SonarHuge128kOnline() CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.Model = Llama31SonarHuge128kOnline
	}
}

// WithMaxTokens sets the max tokens option.
func WithMaxTokens(maxTokens int) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.MaxTokens = maxTokens
	}
}

// WithTemperature sets the temperature option.
func WithTemperature(temperature float64) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.Temperature = temperature
	}
}

// WithTopP sets the top p option.
func WithTopP(topP float64) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.TopP = topP
	}
}

// WithSearchDomainFilter sets the search domain filter option.
func WithSearchDomainFilter(searchDomainFilter []string) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.SearchDomainFilter = searchDomainFilter
	}
}

// WithReturnImages sets the return images option.
func WithReturnImages(returnImages bool) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.ReturnImages = returnImages
	}
}

// WithReturnRelatedQuestions sets the return related questions option.
func WithReturnRelatedQuestions(returnRelatedQuestions bool) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.ReturnRelatedQuestions = returnRelatedQuestions
	}
}

// WithSearchRecencyFilter sets the search recency filter option.
func WithSearchRecencyFilter(searchRecencyFilter string) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.SearchRecencyFilter = searchRecencyFilter
	}
}

// WithTopK sets the top k option.
func WithTopK(topK int) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.TopK = topK
	}
}

// WithStream sets the stream option.
// Determines whether or not to incrementally stream the response
// with server-sent events with content-type: text/event-stream
// The client of this does not handle the stream, it is up to the user to handle the stream.
// func WithStream(stream bool) CompletionRequestOption {
// 	return func(r *CompletionRequest) {
// 		r.Stream = stream
// 	}
// }

// WithPresencePenalty sets the presence penalty option.
func WithPresencePenalty(presencePenalty float64) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.PresencePenalty = presencePenalty
	}
}

// WithFrequencyPenalty sets the frequency penalty option.
func WithFrequencyPenalty(frequencyPenalty float64) CompletionRequestOption {
	return func(r *CompletionRequest) {
		r.FrequencyPenalty = frequencyPenalty
	}
}

// NewCompletionRequest creates a new completion request.
func NewCompletionRequest(opts ...CompletionRequestOption) *CompletionRequest {
	r := DefaultCompletionRequest()
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Validate validates the completion request.
func (r *CompletionRequest) Validate() error {
	validate := validator.New()
	err := validate.Struct(r)
	if err != nil {
		return err
	}
	if err := r.ValidateSearchDomainFilter(); err != nil {
		return err
	}
	if err := r.ValidateSearchRecencyFilter(); err != nil {
		return err
	}
	return nil
}

// ValidateSearchDomainFilter validates the search domain filter.
func (r *CompletionRequest) ValidateSearchDomainFilter() error {
	if len(r.SearchDomainFilter) > MaxLengthOfSearchDomainFilter {
		return ErrSearchDomainFilter
	}
	return nil
}

// ValidateSearchRecencyFilter validates the search recency filter.
func (r *CompletionRequest) ValidateSearchRecencyFilter() error {
	if r.ReturnImages && r.SearchRecencyFilter != "" {
		return ErrSearchRecencyFilter
	}
	if r.SearchRecencyFilter != "" {
		switch r.SearchRecencyFilter {
		case "month", "week", "day", "hour":
			return nil
		default:
			return errors.New("search recency filter must be one of month, week, day, hour")
		}
	}
	return nil
}

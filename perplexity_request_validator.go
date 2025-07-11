package perplexity

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Error definitions for CompletionRequest validation.
var (
	// ErrSearchDomainFilter is returned when the search domain filter exceeds the maximum allowed number of domains.
	ErrSearchDomainFilter = errors.New("search domain filter must be less than or equal to 3")

	// ErrSearchRecencyFilter is returned when the search recency filter is invalid or incompatible.
	ErrSearchRecencyFilter = errors.New("search recency filter must be one of month, week, day, hour and is incompatible with images")

	// ErrRegexAndImages is returned when both regex response format and images are requested, which are incompatible.
	ErrRegexAndImages = errors.New("regex and images are not compatible")

	// ErrRegexAndStream is returned when both regex response format and streaming are requested, which are incompatible.
	ErrRegexAndStream = errors.New("regex and stream are not compatible")

	// ErrValidationFailed is returned when struct validation fails.
	ErrValidationFailed = errors.New("validation failed")

	// ErrStructuredOutputModelRequirement is returned when structured output is used with an unsupported model.
	ErrStructuredOutputModelRequirement = errors.New("structured output (response_format) is only available for the 'sonar' model")

	// ErrStructuredOutputFormatMismatch is returned when the response format configuration doesn't match the type.
	ErrStructuredOutputFormatMismatch = errors.New("response format type must match the provided configuration (json_schema or regex)")

	// ErrStructuredOutputRegexAndImages is returned when regex and images are used together.
	ErrStructuredOutputRegexAndImages = errors.New("regex and images are not compatible")

	// ErrImageDomainFilterTooLong is returned when the image domain filter exceeds the maximum allowed number of domains.
	ErrImageDomainFilterTooLong = errors.New("image domain filter must be less than or equal to 10")

	// ErrImageDomainFilterEmpty is returned when an image domain filter entry is empty.
	ErrImageDomainFilterEmpty = errors.New("image domain filter entry cannot be empty")

	// ErrImageDomainFilterProtocolNotAllowed is returned when an image domain filter entry includes a protocol prefix (http:// or https://).
	ErrImageDomainFilterProtocolNotAllowed = errors.New("image domain filter entry cannot include http:// or https://")

	// ErrImageDomainFilterSubdomainNotAllowed is returned when an image domain filter entry includes subdomains.
	ErrImageDomainFilterSubdomainNotAllowed = errors.New("image domain filter entry cannot include subdomains")

	// ErrImageDomainFilterInvalidFormat is returned when an image domain filter entry has an invalid format.
	ErrImageDomainFilterInvalidFormat = errors.New("image domain filter entry must be a valid domain name (e.g., example.com or -gettyimages.com)")

	// ErrImageFormatFilterTooLong is returned when the image format filter exceeds the maximum allowed number of formats.
	ErrImageFormatFilterTooLong = errors.New("image format filter must be less than or equal to 10")

	// ErrImageFormatFilterEmpty is returned when an image format filter entry is empty.
	ErrImageFormatFilterEmpty = errors.New("image format filter entry cannot be empty")

	// ErrImageFormatFilterDotPrefixNotAllowed is returned when an image format filter entry starts with a dot.
	ErrImageFormatFilterDotPrefixNotAllowed = errors.New("image format filter entry cannot start with a dot (e.g., .gif)")

	// ErrImageFormatFilterMustBeLowercase is returned when an image format filter entry is not in lowercase.
	ErrImageFormatFilterMustBeLowercase = errors.New("image format filter entry must be in lowercase")

	// ErrImageFormatFilterInvalidFormat is returned when an image format filter entry has an invalid format.
	ErrImageFormatFilterInvalidFormat = errors.New("image format filter entry must be a valid format (e.g., jpg, png, webp)")
)

// RequestValidator provides validation functionality for CompletionRequest.
// This is a stateless validator that can be reused across multiple requests.
type RequestValidator struct {
	validator *validator.Validate
}

// NewRequestValidator creates a new RequestValidator instance.
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		validator: validator.New(),
	}
}

// ValidateRequest validates a CompletionRequest using the validator instance.
func (v *RequestValidator) ValidateRequest(req *CompletionRequest) error {
	if req == nil {
		return fmt.Errorf("%w: request cannot be nil", ErrValidationFailed)
	}

	// Validate struct tags first
	if err := v.validator.Struct(req); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}

	// Run custom validations
	validators := []func(*CompletionRequest) error{
		v.validateSearchDomainFilter,
		v.validateSearchRecencyFilter,
		v.validateStructuredOutput,
		v.validateImageDomainFilter,
		v.validateImageFormatFilter,
	}

	for _, validator := range validators {
		if err := validator(req); err != nil {
			return err
		}
	}

	return nil
}

// validateSearchDomainFilter validates the search domain filter.
func (v *RequestValidator) validateSearchDomainFilter(req *CompletionRequest) error {
	if len(req.SearchDomainFilter) > MaxLengthOfSearchDomainFilter {
		return ErrSearchDomainFilter
	}
	return nil
}

// validateSearchRecencyFilter validates the search recency filter.
func (v *RequestValidator) validateSearchRecencyFilter(req *CompletionRequest) error {
	if req.ReturnImages && req.SearchRecencyFilter != "" {
		return ErrSearchRecencyFilter
	}
	if req.SearchRecencyFilter != "" {
		switch req.SearchRecencyFilter {
		case "month", "week", "day", "hour":
			return nil
		default:
			return ErrSearchRecencyFilter
		}
	}
	return nil
}

// validateStructuredOutput validates the structured output configuration.
func (v *RequestValidator) validateStructuredOutput(req *CompletionRequest) error {
	if req.ResponseFormat == nil {
		return nil
	}

	if err := v.validateStructuredOutputModel(req); err != nil {
		return err
	}

	if err := v.validateStructuredOutputFormat(req); err != nil {
		return err
	}

	return nil
}

// validateStructuredOutputModel validates that the model supports structured output.
func (v *RequestValidator) validateStructuredOutputModel(req *CompletionRequest) error {
	if req.Model != "sonar" {
		return ErrStructuredOutputModelRequirement
	}
	return nil
}

// validateStructuredOutputFormat validates the response format configuration.
func (v *RequestValidator) validateStructuredOutputFormat(req *CompletionRequest) error {
	switch req.ResponseFormat.Type {
	case "json_schema":
		return v.validateJSONSchemaFormat(req)
	case "regex":
		return v.validateRegexFormat(req)
	default:
		return ErrStructuredOutputFormatMismatch
	}
}

// validateJSONSchemaFormat validates JSON Schema format configuration.
func (v *RequestValidator) validateJSONSchemaFormat(req *CompletionRequest) error {
	if req.ResponseFormat.JSONSchema == nil {
		return ErrStructuredOutputFormatMismatch
	}
	if req.ResponseFormat.Regex != nil {
		return ErrStructuredOutputFormatMismatch
	}
	return nil
}

// validateRegexFormat validates Regex format configuration.
func (v *RequestValidator) validateRegexFormat(req *CompletionRequest) error {
	if req.ResponseFormat.Regex == nil {
		return ErrStructuredOutputFormatMismatch
	}
	if req.ResponseFormat.JSONSchema != nil {
		return ErrStructuredOutputFormatMismatch
	}
	if req.ReturnImages {
		return ErrStructuredOutputRegexAndImages
	}
	return nil
}

// validateImageDomainFilter validates the image domain filter.
func (v *RequestValidator) validateImageDomainFilter(req *CompletionRequest) error {
	if len(req.ImageDomainFilter) == 0 {
		return nil
	}

	// Check list length (≤10 entries)
	maxNumberOfDomains := 10
	if len(req.ImageDomainFilter) > maxNumberOfDomains {
		return ErrImageDomainFilterTooLong
	}

	// Validate each domain
	for _, domain := range req.ImageDomainFilter {
		if err := validateImageDomain(domain); err != nil {
			return err
		}
	}

	return nil
}

// validateImageFormatFilter validates the image format filter.
func (v *RequestValidator) validateImageFormatFilter(req *CompletionRequest) error {
	if len(req.ImageFormatFilter) == 0 {
		return nil
	}

	// Validate each format
	for _, format := range req.ImageFormatFilter {
		if err := validateImageFormat(format); err != nil {
			return err
		}
	}

	return nil
}

// validateImageDomain validates a single image domain filter entry.
func validateImageDomain(domain string) error {
	if domain == "" {
		return ErrImageDomainFilterEmpty
	}

	// Check for protocol prefixes (should not include http://, https://)
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return ErrImageDomainFilterProtocolNotAllowed
	}

	// Check for subdomains (should not include subdomains)
	if strings.Count(domain, ".") > 1 {
		return ErrImageDomainFilterSubdomainNotAllowed
	}

	// Check for valid domain format (simple domain names)
	// Allow exclusion prefix (-) but validate the rest
	cleanDomain := domain
	if strings.HasPrefix(domain, "-") {
		cleanDomain = domain[1:]
	}

	// Basic domain validation (alphanumeric, hyphens, dots)
	matched, err := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`, cleanDomain)
	if err != nil || !matched {
		return ErrImageDomainFilterInvalidFormat
	}

	return nil
}

// validateImageFormat validates a single image format filter entry.
func validateImageFormat(format string) error {
	if format == "" {
		return ErrImageFormatFilterEmpty
	}

	// Check for dot prefix (should not include ie. gif not .gif)
	if strings.HasPrefix(format, ".") {
		return ErrImageFormatFilterDotPrefixNotAllowed
	}

	// Check for uppercase (must be lowercase)
	if format != strings.ToLower(format) {
		return ErrImageFormatFilterMustBeLowercase
	}

	// Check for valid format (alphanumeric only)
	matched, err := regexp.MatchString(`^[a-z0-9]+$`, format)
	if err != nil || !matched {
		return ErrImageFormatFilterInvalidFormat
	}

	return nil
}

// CompletionRequest validators for backward compatibility

// Validate validates the completion request using the default validator.
//
//go:deprecated
func (r *CompletionRequest) Validate() error {
	return NewRequestValidator().ValidateRequest(r)
}

// ValidateSearchDomainFilter validates the search domain filter.
//
//go:deprecated
func (r *CompletionRequest) ValidateSearchDomainFilter() error {
	return NewRequestValidator().validateSearchDomainFilter(r)
}

// ValidateSearchRecencyFilter validates the search recency filter.
//
//go:deprecated
func (r *CompletionRequest) ValidateSearchRecencyFilter() error {
	return NewRequestValidator().validateSearchRecencyFilter(r)
}

// ValidateStructuredOutput validates the structured output configuration.
//
//go:deprecated
func (r *CompletionRequest) ValidateStructuredOutput() error {
	return NewRequestValidator().validateStructuredOutput(r)
}

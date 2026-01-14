package perplexity

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	// MaxLengthOfDomainFilter is the maximum number of domains allowed in the search domain filter.
	MaxLengthOfDomainFilter = 10
)

// Error definitions for CompletionRequest validation.
var (
	// ErrSearchRecencyFilter is returned when the search recency filter is invalid or incompatible.
	ErrSearchRecencyFilter = errors.New("search recency filter must be one of month, week, day, hour and is incompatible with images")

	// ErrRegexAndImages is returned when both regex response format and images are requested, which are incompatible.
	ErrRegexAndImages = errors.New("regex and images are not compatible")

	// ErrRegexAndStream is returned when both regex response format and streaming are requested, which are incompatible.
	ErrRegexAndStream = errors.New("regex and stream are not compatible")

	// ErrValidationFailed is returned when struct validation fails.
	ErrValidationFailed = errors.New("validation failed")

	// ErrStructuredOutputFormatMismatch is returned when the response format configuration doesn't match the type.
	ErrStructuredOutputFormatMismatch = errors.New("response format type must match the provided configuration (json_schema or regex)")

	// ErrStructuredOutputRegexAndImages is returned when regex and images are used together.
	ErrStructuredOutputRegexAndImages = errors.New("regex and images are not compatible")

	// ErrDomainFilterTooLong is returned when the domain filter exceeds the maximum allowed number of domains.
	ErrDomainFilterTooLong = errors.New("domain filter must be less than or equal to 10")

	// ErrDomainFilterEmpty is returned when a domain filter entry is empty.
	ErrDomainFilterEmpty = errors.New("domain filter entry cannot be empty")

	// ErrDomainFilterProtocolNotAllowed is returned when a domain filter entry includes a protocol prefix (http:// or https://).
	ErrDomainFilterProtocolNotAllowed = errors.New("domain filter entry cannot include http:// or https://")

	// ErrDomainFilterWWWNotAllowed is returned when a domain filter entry includes a www. prefix.
	ErrDomainFilterWWWNotAllowed = errors.New("domain filter entry cannot include www. prefix")

	// ErrDomainFilterSubdomainNotAllowed is returned when a domain filter entry includes subdomains.
	ErrDomainFilterSubdomainNotAllowed = errors.New("domain filter entry cannot include subdomains")

	// ErrDomainFilterInvalidFormat is returned when a domain filter entry has an invalid format.
	ErrDomainFilterInvalidFormat = errors.New("domain filter entry must be a valid domain name (e.g., example.com or -gettyimages.com)")

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

	// ErrDateFilterInvalidFormat is returned when a date filter has an invalid format.
	ErrDateFilterInvalidFormat = errors.New("date filter must be in format %m/%d/%Y (e.g., 3/1/2025, 12/31/2024)")

	// ErrReasoningEffortModelRequirement is returned when reasoning_effort is used with an unsupported model.
	ErrReasoningEffortModelRequirement = errors.New("reasoning_effort is only available for the '" + ModelSonarDeepResearch + "' model")
	// ErrSearchDomainInvalid is returned when search_domain is set to an invalid value.
	ErrSearchDomainInvalid = errors.New("search_domain must be 'sec' or empty")

	// Multimodal validation errors.

	// ErrImageAndRegexIncompatible is returned when images are used with regex response format.
	ErrImageAndRegexIncompatible = errors.New("images cannot be used with regex response format")

	// ErrImageAndDeepResearchIncompatible is returned when images are used with sonar-deep-research model.
	ErrImageAndDeepResearchIncompatible = errors.New("images cannot be used with sonar-deep-research model")

	// ErrMultimodalAndRegularMessages is returned when both multimodal and regular messages are present.
	ErrMultimodalAndRegularMessages = errors.New("cannot use both multimodal messages and regular messages")

	// ErrMultimodalContentValidation is returned when multimodal content validation fails.
	ErrMultimodalContentValidation = errors.New("multimodal content validation failed")

	// ErrTextContentEmpty is returned when text content is empty.
	ErrTextContentEmpty = errors.New("text content cannot be empty")

	// ErrImageURLContentNil is returned when image URL content is nil.
	ErrImageURLContentNil = errors.New("image URL content cannot be nil")

	// ErrLanguagePreferenceInvalid is returned when language preference format is invalid.
	ErrLanguagePreferenceInvalid = errors.New("language_preference must be valid ISO 639 format (e.g., 'en', 'en-US')")
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
		v.validateSearchDomain,
		v.validateStructuredOutput,
		v.validateImageDomainFilter,
		v.validateImageFormatFilter,
		v.validateDateFilters,
		v.validateReasoningEffort,
		v.validateLanguagePreference,
		v.validateMultimodalMessages,
		v.validateImageCompatibility,
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
	return validateDomainList(req.SearchDomainFilter)
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

// validateSearchDomain validates the search domain parameter.
func (v *RequestValidator) validateSearchDomain(req *CompletionRequest) error {
	// Empty search domain is valid (omitted from request)
	if req.SearchDomain == "" {
		return nil
	}

	// Only "sec" is currently supported
	if req.SearchDomain != "sec" {
		return ErrSearchDomainInvalid
	}

	return nil
}

// validateStructuredOutput validates the structured output configuration.
func (v *RequestValidator) validateStructuredOutput(req *CompletionRequest) error {
	if req.ResponseFormat == nil {
		return nil
	}

	if err := v.validateStructuredOutputFormat(req); err != nil {
		return err
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
	return validateDomainList(req.ImageDomainFilter)
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

func validateDomainList(domainList []string) error {
	if len(domainList) == 0 {
		return nil
	}

	if len(domainList) > MaxLengthOfDomainFilter {
		return ErrDomainFilterTooLong
	}

	// Validate each domain
	for _, domain := range domainList {
		if err := validateDomain(domain); err != nil {
			return err
		}
	}

	return nil
}

// validateDomain validates a single domain filter entry.
func validateDomain(domain string) error {
	if domain == "" {
		return ErrDomainFilterEmpty
	}

	// Check for valid domain format (simple domain names)
	// Allow exclusion prefix (-) but validate the rest
	domain = strings.TrimPrefix(domain, "-")

	// Check for protocol prefixes (should not include http://, https://)
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return ErrDomainFilterProtocolNotAllowed
	}

	if strings.HasPrefix(domain, "www.") {
		return ErrDomainFilterWWWNotAllowed
	}

	// Check for subdomains (should not include subdomains)
	if strings.Count(domain, ".") > 1 {
		return ErrDomainFilterSubdomainNotAllowed
	}

	// Basic domain validation (alphanumeric, hyphens, dots)
	matched, err := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`, domain)
	if err != nil || !matched {
		return ErrDomainFilterInvalidFormat
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

// validateDateFilters validates all date filter fields have correct format.
func (v *RequestValidator) validateDateFilters(req *CompletionRequest) error {
	dateFields := []string{
		req.SearchAfterDateFilter,
		req.SearchBeforeDateFilter,
		req.LastUpdatedAfterFilter,
		req.LastUpdatedBeforeFilter,
		req.PublishedAfter,
		req.PublishedBefore,
	}

	for _, dateStr := range dateFields {
		if dateStr != "" && !isValidDateFormat(dateStr) {
			return ErrDateFilterInvalidFormat
		}
	}

	return nil
}

// isValidDateFormat checks if a date string matches the format %m/%d/%Y.
func isValidDateFormat(dateStr string) bool {
	// Regex pattern for %m/%d/%Y format
	// Month: 1-12 (without leading zeros for 1-9)
	// Day: 1-31 (without leading zeros for 1-9)
	// Year: 4 digits
	pattern := `^(1[0-2]|[1-9])/(3[01]|[12][0-9]|[1-9])/\d{4}$`
	matched, err := regexp.MatchString(pattern, dateStr)
	return err == nil && matched
}

// validateReasoningEffort validates that reasoning_effort is only used with sonar-deep-research model.
func (v *RequestValidator) validateReasoningEffort(req *CompletionRequest) error {
	if req.ReasoningEffort == "" {
		return nil
	}

	// If reasoning_effort is set, model must be sonar-deep-research
	if req.Model != ModelSonarDeepResearch {
		return ErrReasoningEffortModelRequirement
	}

	return nil
}

// validateLanguagePreference validates language preference format.
func (v *RequestValidator) validateLanguagePreference(req *CompletionRequest) error {
	if req.LanguagePreference == "" {
		return nil
	}

	// Check length (ISO 639-1 codes are 2 characters, with optional country code up to 5 total)
	if len(req.LanguagePreference) < 2 || len(req.LanguagePreference) > 5 {
		return fmt.Errorf("%w: length must be 2-5 characters, got: %s", ErrLanguagePreferenceInvalid, req.LanguagePreference)
	}

	// Validate format: lowercase language code, optional uppercase country code with hyphen
	// Examples: "en", "fr", "en-US", "fr-CA"
	matched, err := regexp.MatchString(`^[a-z]{2}(-[A-Z]{2})?$`, req.LanguagePreference)
	if err != nil {
		return fmt.Errorf("failed to validate language preference: %w", err)
	}
	if !matched {
		return fmt.Errorf("%w, got: %s", ErrLanguagePreferenceInvalid, req.LanguagePreference)
	}

	return nil
}

// validateMultimodalMessages validates multimodal message structure and content.
func (v *RequestValidator) validateMultimodalMessages(req *CompletionRequest) error {
	// Check that both multimodal and regular messages are not present
	if len(req.MultimodalMessages) > 0 && len(req.Messages) > 0 {
		return ErrMultimodalAndRegularMessages
	}

	// If no multimodal messages, skip validation
	if len(req.MultimodalMessages) == 0 {
		return nil
	}

	// Validate each multimodal message
	for _, msg := range req.MultimodalMessages {
		if err := v.validateMultimodalMessage(msg); err != nil {
			return fmt.Errorf("%w: %w", ErrMultimodalContentValidation, err)
		}
	}

	return nil
}

// validateMultimodalMessage validates a single multimodal message.
func (v *RequestValidator) validateMultimodalMessage(msg MultimodalMessage) error {
	// Validate struct tags
	if err := v.validator.Struct(msg); err != nil {
		return fmt.Errorf("multimodal message validation failed: %w", err)
	}

	// Validate each content item
	for _, content := range msg.Content {
		if err := v.validateContent(content); err != nil {
			return err
		}
	}

	return nil
}

// validateContent validates a single content item.
func (v *RequestValidator) validateContent(content Content) error {
	// Validate struct tags
	if err := v.validator.Struct(content); err != nil {
		return fmt.Errorf("content validation failed: %w", err)
	}

	// Validate specific content types
	switch content.Type {
	case ContentTypeText:
		if content.Text == nil || *content.Text == "" {
			return ErrTextContentEmpty
		}
	case ContentTypeImageURL:
		if content.ImageURL == nil {
			return ErrImageURLContentNil
		}
		// Validate image URL
		processor := NewImageProcessor()
		if err := processor.ValidateImageURL(content.ImageURL.URL); err != nil {
			return fmt.Errorf("invalid image URL: %w", err)
		}
	default:
		return ErrInvalidContentType
	}

	return nil
}

// validateImageCompatibility checks image compatibility with other request options.
func (v *RequestValidator) validateImageCompatibility(req *CompletionRequest) error {
	hasImages := req.HasImages()

	// Skip validation if no images
	if !hasImages {
		return nil
	}

	// Check regex incompatibility
	if req.ResponseFormat != nil && req.ResponseFormat.Type == "regex" {
		return ErrImageAndRegexIncompatible
	}

	// Check sonar-deep-research model incompatibility
	if req.Model == ModelSonarDeepResearch {
		return ErrImageAndDeepResearchIncompatible
	}

	// Check search recency filter incompatibility (from existing validation)
	if req.SearchRecencyFilter != "" {
		return ErrSearchRecencyFilter
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

// ValidateSearchDomain validates the search domain parameter.
//
//go:deprecated
func (r *CompletionRequest) ValidateSearchDomain() error {
	return NewRequestValidator().validateSearchDomain(r)
}

package options

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/Azure/kubelogin/pkg/internal/token"
)

// ValidationRule defines a validation rule for an option field
type ValidationRule struct {
	Field       string
	Required    bool
	ValidValues []string
	Validator   func(interface{}) error
}

// getValidationRules returns validation rules based on struct tags and business logic
func (o *UnifiedOptions) getValidationRules() []ValidationRule {
	var rules []ValidationRule

	val := reflect.ValueOf(o).Elem()
	typ := reflect.TypeOf(o).Elem()

	// Parse validation rules from struct tags
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		validateTag := fieldType.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		rule := ValidationRule{
			Field: fieldType.Name,
		}

		// Parse validation tag
		validations := strings.Split(validateTag, ",")
		for _, validation := range validations {
			validation = strings.TrimSpace(validation)

			if validation == "required" {
				rule.Required = true
			} else if strings.HasPrefix(validation, "oneof=") {
				values := strings.TrimPrefix(validation, "oneof=")
				rule.ValidValues = strings.Split(values, " ")
			} else if validation == "omitempty,url" {
				rule.Validator = func(v interface{}) error {
					if str, ok := v.(string); ok && str != "" {
						if _, err := url.ParseRequestURI(str); err != nil {
							return fmt.Errorf("invalid URL format")
						}
					}
					return nil
				}
			}
		}

		rules = append(rules, rule)
	}

	// Add command-specific rules
	if o.command == TokenCommand {
		rules = append(rules, ValidationRule{
			Field:    "ServerID",
			Required: true,
		})
	}

	// Add login method-specific rules
	switch o.LoginMethod {
	case token.ServicePrincipalLogin:
		rules = append(rules, ValidationRule{
			Field:    "ClientID",
			Required: true,
		})
		if o.ClientCert == "" {
			rules = append(rules, ValidationRule{
				Field:    "ClientSecret",
				Required: true,
			})
		}
	case token.DeviceCodeLogin, token.InteractiveLogin:
		rules = append(rules, ValidationRule{
			Field:    "ClientID",
			Required: true,
		})
	case token.ROPCLogin:
		rules = append(rules, []ValidationRule{
			{Field: "ClientID", Required: true},
			{Field: "Username", Required: true},
			{Field: "Password", Required: true},
		}...)
	case token.WorkloadIdentityLogin:
		rules = append(rules, []ValidationRule{
			{Field: "ClientID", Required: true},
			{Field: "FederatedTokenFile", Required: true},
		}...)
	}

	return rules
}

// validateField validates a single field based on the validation rule
func (o *UnifiedOptions) validateField(rule ValidationRule) error {
	val := reflect.ValueOf(o).Elem()
	field := val.FieldByName(rule.Field)

	if !field.IsValid() {
		return fmt.Errorf("field %s not found", rule.Field)
	}

	fieldValue := field.Interface()

	// Check required fields
	if rule.Required {
		switch v := fieldValue.(type) {
		case string:
			if v == "" {
				return fmt.Errorf("%s is required", strings.ToLower(rule.Field))
			}
		case bool:
			// Booleans are considered valid even if false
		default:
			if field.IsZero() {
				return fmt.Errorf("%s is required", strings.ToLower(rule.Field))
			}
		}
	}

	// Check valid values
	if len(rule.ValidValues) > 0 {
		if str, ok := fieldValue.(string); ok && str != "" {
			valid := false
			for _, validValue := range rule.ValidValues {
				if str == validValue {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("%s must be one of: %s", strings.ToLower(rule.Field), strings.Join(rule.ValidValues, ", "))
			}
		}
	}

	// Check custom validator
	if rule.Validator != nil {
		if err := rule.Validator(fieldValue); err != nil {
			return fmt.Errorf("%s validation failed: %v", strings.ToLower(rule.Field), err)
		}
	}

	return nil
}

// validateCustomRules validates business logic rules that can't be expressed in struct tags
func (o *UnifiedOptions) validateCustomRules() error {
	var errors []string

	// Validate AuthorityHost format
	if o.AuthorityHost != "" {
		if u, err := url.ParseRequestURI(o.AuthorityHost); err != nil {
			errors = append(errors, fmt.Sprintf("authority host %q is not valid: %s", o.AuthorityHost, err))
		} else if u.Scheme == "" || u.Host == "" {
			errors = append(errors, fmt.Sprintf("authority host %q is not valid", o.AuthorityHost))
		} else if !strings.HasSuffix(o.AuthorityHost, "/") {
			errors = append(errors, fmt.Sprintf("authority host %q should have a trailing slash", o.AuthorityHost))
		}
	}

	// Validate PoP token configuration
	if o.IsPoPTokenEnabled && o.PoPTokenClaims == "" {
		errors = append(errors, "if enabling pop token mode, please provide the pop-claims flag containing the PoP token claims as a comma-separated string: `u=popClaimHost,key1=val1`")
	}

	if o.PoPTokenClaims != "" && !o.IsPoPTokenEnabled {
		errors = append(errors, "pop-enabled flag is required to use the PoP token feature. Please provide both pop-enabled and pop-claims flags")
	}

	// Validate timeout
	if o.Timeout <= 0 {
		errors = append(errors, "timeout must be greater than 0")
	}

	// Validate login method exists
	supportedLogins := strings.Split(token.GetSupportedLogins(), ", ")
	validLoginMethod := false
	for _, login := range supportedLogins {
		if o.LoginMethod == login {
			validLoginMethod = true
			break
		}
	}
	if !validLoginMethod {
		errors = append(errors, fmt.Sprintf("'%s' is not a supported login method. Supported methods: %s", o.LoginMethod, token.GetSupportedLogins()))
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// ValidateForConversion validates the complete configuration after field extraction during conversion.
// This is the single validation point for convert command - much cleaner than dual validation.
func (o *UnifiedOptions) ValidateForConversion() error {
	var errors []string

	// Basic login method validation
	if o.LoginMethod == "" {
		errors = append(errors, "login method is required")
	} else {
		supportedLogins := strings.Split(token.GetSupportedLogins(), ", ")
		validLoginMethod := false
		for _, login := range supportedLogins {
			if o.LoginMethod == login {
				validLoginMethod = true
				break
			}
		}
		if !validLoginMethod {
			errors = append(errors, fmt.Sprintf("login must be one of: %s", token.GetSupportedLogins()))
		}
	}

	// URL format validation for any provided URLs
	if o.RedirectURL != "" {
		if _, err := url.ParseRequestURI(o.RedirectURL); err != nil {
			errors = append(errors, "invalid redirect URL format")
		}
	}

	if o.AuthorityHost != "" {
		if _, err := url.ParseRequestURI(o.AuthorityHost); err != nil {
			errors = append(errors, "invalid authority host URL format")
		}
	}

	// For conversion, we're more lenient about missing fields since:
	// 1. Some fields can be provided via environment variables during token operations
	// 2. Conversion is just restructuring the kubeconfig, not performing authentication
	// 3. The actual validation happens when get-token is called

	// Only validate truly critical conflicts or malformed data
	if o.LoginMethod == token.ServicePrincipalLogin && o.ClientCert != "" && o.ClientSecret != "" {
		errors = append(errors, "cannot specify both client certificate and client secret for service principal")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// ValidateForTokenExecution validates the options for immediate token execution
// This performs strict validation requiring all values to be present now
func (o *UnifiedOptions) ValidateForTokenExecution() error {
	var errors []string

	// Get validation rules and validate
	rules := o.getValidationRules()
	for _, rule := range rules {
		if err := o.validateField(rule); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// Custom validation logic
	if err := o.validateCustomRules(); err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

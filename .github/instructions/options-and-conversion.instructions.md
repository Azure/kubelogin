# Options and Conversion Instructions

This document provides guidance for developers on how to add new command line options, validation rules, and conversion logic to kubelogin using the unified options system.

## Validation Architecture Overview

Kubelogin has **two distinct validation methods** for different use cases:

### 1. `ValidateForTokenExecution()` - Strict Validation
- **Purpose**: Validates options for immediate token execution (`get-token` command)
- **Requirements**: All required fields must be present NOW  
- **Use case**: When kubelogin needs to authenticate immediately
- **Location**: `pkg/internal/options/validation.go`
- **Method name**: `ValidateForTokenExecution()`

### 2. `ValidateForConversion()` - Lenient Validation  
- **Purpose**: Validates options for kubeconfig conversion (`convert-kubeconfig` command)
- **Requirements**: Allows missing values that can be provided via environment variables at runtime
- **Use case**: When building kubeconfig that will be executed later
- **Location**: `pkg/internal/options/validation.go`
- **Method name**: `ValidateForConversion()`

## Validation Architecture Pattern

The validation follows a **self-contained execution pattern** where each command handles its own validation:

```go
func (o *UnifiedOptions) ExecuteCommand(ctx context.Context, flags *pflag.FlagSet) error {
    // Pure dispatcher - no validation logic here
    switch o.command {
    case ConvertCommand:
        return o.executeConvert()  // Handles its own validation internally
    case TokenCommand:
        return o.executeToken(ctx) // Handles its own validation internally
    }
}

func (o *UnifiedOptions) executeToken(ctx context.Context) error {
    // Self-contained: validate first, then execute
    if err := o.ValidateForTokenExecution(); err != nil {
        return err
    }
    // ... execution logic
}

func (o *UnifiedOptions) executeConvert() error {
    // Self-contained: extract fields, then validate complete config
    // ... field extraction logic
    if err := tempOptions.ValidateForConversion(); err != nil {
        return err
    }
    // ... execution logic
}
```

**Benefits of this pattern:**
- **Consistent**: Both commands follow the same self-contained pattern
- **Clear separation**: `ExecuteCommand()` is purely a dispatcher
- **Single responsibility**: Each execution method owns its validation logic
- **Easy to understand**: Validation happens at the start of each execution method

## Validation Method Naming

**IMPORTANT**: The validation methods have specific, descriptive names to avoid confusion:
- `ValidateForTokenExecution()` - For immediate token execution (strict)
- `ValidateForConversion()` - For kubeconfig conversion (lenient)

**DO NOT** create a generic `Validate()` method, as it's ambiguous which validation behavior is needed.

## When Adding New Fields - Validation Requirements

When adding new fields to `UnifiedOptions`, you need to update validation in **both** methods if applicable:

### Required Fields (Authentication Critical)
For fields that are **always required** for authentication:

1. **Add struct tag validation**: Add `validate:"required"` to the field
2. **Update `ValidateForTokenExecution()`**: Ensure strict validation catches missing values immediately
3. **Update `ValidateForConversion()`**: Add appropriate validation logic that accounts for runtime environment variables

**Example**:
```go
// In unified.go
TenantID string `flag:"tenant-id,t" env:"AZURE_TENANT_ID" validate:"required" description:"AAD tenant ID"`

// In validation.go - ValidateForTokenExecution()
if o.TenantID == "" {
    errors = append(errors, "tenant-id is required")
}

// In validation.go - ValidateForConversion()  
// Allow empty if it can come from env vars or kubeconfig extraction
```

### Optional/Conditional Fields
For fields that are **optional** or **login-method-specific**:

1. **Add conditional validation**: Use login method checks in both validation functions  
2. **Update extraction logic**: Ensure `populateFromExecConfig()` can extract the field from existing kubeconfig
3. **Update execution logic**: Ensure `shouldIncludeArg()` properly includes/excludes the field

**Example**:
```go
// In unified.go  
ClientSecret string `flag:"client-secret" env:"AZURE_CLIENT_SECRET" description:"Client secret for SPN login"`

// In validation.go - ValidateForTokenExecution()
if o.LoginMethod == "spn" && o.ClientSecret == "" && o.ClientCert == "" {
    errors = append(errors, "either client-secret or client-certificate is required for SPN login")
}

// In validation.go - ValidateForConversion()
// More lenient - allow missing if env vars will provide at runtime
```

## Design Principles

### Separation of Concerns  
- **Convert-time validation**: "Will this work when executed later?"
- **Execution-time validation**: "Can I authenticate right now?"

### Environment Variable Support
- Conversion allows missing values that will come from env vars at runtime
- Token execution requires all values to be present immediately

### Field Extraction
- During conversion, fields can be extracted from existing kubeconfig
- Validation happens **after** extraction to ensure complete configuration

## Key File Locations

### Core Validation Files
- `pkg/internal/options/unified.go` - Main options struct and CLI integration
- `pkg/internal/options/validation.go` - Both `ValidateForTokenExecution()` and `ValidateForConversion()` methods, plus validation utilities
- `pkg/internal/options/execution.go` - Command execution and field extraction

### Test Files That Use Validation
- `pkg/internal/options/validation_test.go` - Uses `ValidateForTokenExecution()`
- `pkg/internal/options/unified_test.go` - Uses `ValidateForTokenExecution()`  
- `pkg/internal/options/field_extraction_diagnostic_test.go` - Uses `ValidateForTokenExecution()`

**IMPORTANT**: When adding new validation, update tests to use the correct validation method.

## Common Patterns When Adding Fields

### Adding Authentication Fields
```go
// 1. Add to UnifiedOptions struct with proper tags
ClientNewField string `flag:"new-field" env:"AZURE_NEW_FIELD" validate:"required_if=LoginMethod spn" description:"New authentication field"`

// 2. Add to ValidateForTokenExecution() - strict validation
if o.LoginMethod == "spn" && o.ClientNewField == "" {
    return fmt.Errorf("--new-field is required for SPN authentication")  
}

// 3. Add to ValidateForConversion() - lenient validation
// (May allow empty if env var will provide at runtime)

// 4. Add to populateFromExecConfig() for extraction  
case "--new-field":
    if o.ClientNewField == "" {
        o.ClientNewField = value
    }

// 5. Add to shouldIncludeArg() for inclusion logic
case "spn":
    return fieldName == "ClientNewField" || ...
```

### Adding Optional Fields
```go
// 1. Add to struct without "required" validation
NewOptionalField string `flag:"optional-field" description:"Optional field"`

// 2. Add to extraction and inclusion logic as needed
// 3. No strict validation required for optional fields
```

## Testing Guidelines

### Unit Tests
- Test both validation methods with new fields
- Use `ValidateForTokenExecution()` for token command scenarios
- Use conversion integration tests for `ValidateForConversion()` scenarios
- Test field extraction from kubeconfig
- Test environment variable loading

### Integration Tests
- Add test cases covering the new field in conversion scenarios
- Ensure both legacy and unified modes work correctly
- Test that the field properly extracts from existing kubeconfig

## Migration Notes

### Validation Method Migration
- All test files have been updated to use `ValidateForTokenExecution()` instead of the ambiguous `Validate()`
- New code should always use the descriptive method names
- This prevents confusion about which validation behavior is expected

### Legacy Compatibility  
- New fields should work in both legacy and unified modes
- Environment variable names should follow existing patterns
- Deprecation warnings for old environment variables when needed

### Breaking Changes
- Changes to validation logic may affect existing users
- Document breaking changes in CHANGELOG.md
- Consider feature flags for significant validation changes

The kubelogin project uses a **unified options system** that dramatically simplifies adding new CLI arguments. Instead of modifying 6+ files (as was required before), you now typically only need to modify 1-2 files.

## Quick Reference

### Adding a New CLI Option

**Before (Legacy - 6 touch points):**
```go
// 1. Add constant definitions
argNewOption = "--new-option"
flagNewOption = "new-option"

// 2. Modify getArgValues() function signature
func getArgValues(..., argNewOptionVal string, ...) 

// 3. Add manual processing logic
if o.isSet(flagNewOption) {
    argNewOptionVal = o.TokenOptions.NewOption
}

// 4. Update function call sites
// 5. Add conditional argument building  
// 6. Update validation logic
```

**After (Unified - 1 touch point):**
```go
// Single struct field addition with tags
NewOption string `flag:"new-option" env:"AZURE_NEW_OPTION" description:"Description of new option"`
```

## Detailed Instructions

### 1. Adding a New CLI Option

#### Step 1: Add Field to UnifiedOptions

Edit `pkg/internal/options/unified.go` and add your new field:

```go
type UnifiedOptions struct {
    // ... existing fields ...
    
    // Your new option
    NewOption string `flag:"new-option" env:"AZURE_NEW_OPTION,ALT_NEW_OPTION" description:"Description of the new option" default:"default-value"`
}
```

#### Struct Tag Reference:

| Tag | Purpose | Example | Required |
|-----|---------|---------|----------|
| `flag` | CLI flag name (with optional short form) | `flag:"new-option,n"` | ✅ Yes |
| `env` | Environment variable(s) with fallbacks | `env:"AZURE_NEW_OPTION,ALT_NEW_OPTION"` | ❌ Optional |
| `description` | Help text for the flag | `description:"Description text"` | ✅ Yes |
| `default` | Default value | `default:"default-value"` | ❌ Optional |
| `validate` | Validation rules | `validate:"required,oneof=value1 value2"` | ❌ Optional |
| `sensitive` | Hide value in logs | `sensitive:"true"` | ❌ Optional |
| `commands` | Which commands should include this flag | `commands:"convert,token"` | ❌ Optional |

#### Step 2: Update ToTokenOptions() (if needed)

If the new option needs to be available in legacy token operations, add it to the conversion method in `pkg/internal/options/execution.go`:

```go
func (o *UnifiedOptions) ToTokenOptions() *token.Options {
    return &token.Options{
        // ... existing fields ...
        NewOption: o.NewOption,
    }
}
```

**That's it!** The option is now:
- ✅ Automatically registered as a CLI flag
- ✅ Automatically loaded from environment variables
- ✅ Automatically included in help text
- ✅ Automatically converted to legacy options
- ✅ Automatically included in kubeconfig conversion arguments

### 2. Adding Validation Rules

#### Built-in Validation Tags

The following validation tags are **currently implemented** in the system:

```go
// Required field - IMPLEMENTED
RequiredField string `flag:"required-field" validate:"required"`

// Must be one of specific values - IMPLEMENTED
LoginMethod string `flag:"login" validate:"required,oneof=devicecode interactive spn ropc msi azurecli azd workloadidentity"`

// URL validation (only supports "omitempty,url" format) - IMPLEMENTED
AuthorityHost string `flag:"authority-host" validate:"omitempty,url"`
```

**Note**: The following validation examples are **NOT YET IMPLEMENTED** but could be added in the future:

```go
// Minimum/maximum values - NOT IMPLEMENTED
Timeout time.Duration `flag:"timeout" validate:"min=1s,max=300s"`

// Email format - NOT IMPLEMENTED
Email string `flag:"email" validate:"email"`

// Custom regex - NOT IMPLEMENTED
Pattern string `flag:"pattern" validate:"regexp=^[a-zA-Z0-9]+$"`
```

#### Currently Supported Validation Rules

Based on the implementation in `pkg/internal/options/validation.go`, the system supports:

1. **`required`** - Field must not be empty/zero value
2. **`oneof=value1 value2 value3`** - Field must match one of the specified values
3. **`omitempty,url`** - If field is not empty, it must be a valid URL format

To add new validation types, modify the validation parsing logic in `getValidationRules()`.

#### Custom Validation Rules

For complex validation logic, add rules to `pkg/internal/options/validation.go`:

```go
func (o *UnifiedOptions) validateCustomRules() error {
    var errors []string
    
    // Example: Service principal requires either client secret or certificate
    if o.LoginMethod == "spn" {
        if o.ClientSecret == "" && o.ClientCert == "" {
            errors = append(errors, "service principal login requires either --client-secret or --client-certificate")
        }
        if o.ClientSecret != "" && o.ClientCert != "" {
            errors = append(errors, "service principal login cannot use both --client-secret and --client-certificate")
        }
    }
    
    // Your custom validation here
    if o.NewOption == "special-value" && o.OtherOption == "" {
        errors = append(errors, "when --new-option is 'special-value', --other-option is required")
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed:\n  - %s", strings.Join(errors, "\n  - "))
    }
    return nil
}
```

### 3. Command-Specific Options

Some options only apply to specific commands. Use the `commands` struct tag to control which commands include the option:

```go
// Option only for convert command
Context string `flag:"context" commands:"convert" description:"The name of the kubeconfig context to use"`

// Option only for token command  
ServerID string `flag:"server-id" commands:"token" description:"AAD server application ID"`

// Option for both commands (default behavior if commands tag is omitted)
ClientID string `flag:"client-id" description:"AAD client application ID"`

// Option for multiple specific commands
MultiOption string `flag:"multi-option" commands:"convert,token" description:"Available for both commands"`
```

**Command Values:**
- `"convert"` - Only available in `kubelogin convert-kubeconfig`
- `"token"` - Only available in `kubelogin get-token`  
- `"convert,token"` - Available in both commands
- No `commands` tag - Available in all commands (default)

**Real Example from kubelogin:**
```go
// These fields are only shown in convert-kubeconfig command help
Context        string `flag:"context" commands:"convert" description:"The name of the kubeconfig context to use"`
AzureConfigDir string `flag:"azure-config-dir" commands:"convert" description:"Azure CLI config path"`
```

The system automatically filters flags during registration, so:
- `kubelogin convert-kubeconfig --help` shows `--context` and `--azure-config-dir` flags
- `kubelogin get-token --help` does NOT show these flags

```go
// Option that only applies to service principal login
ClientCert string `flag:"client-certificate" env:"AZURE_CLIENT_CERTIFICATE_PATH" description:"Client certificate for service principal authentication"`
### 4. Login Method-Specific Options

Some options only apply to specific login methods. Add logic to `buildExecConfig()` in `pkg/internal/options/convert.go` for method-specific inclusion:

```go
func (o *UnifiedOptions) shouldIncludeArg(fieldName, value string) bool {
    switch fieldName {
    case "ClientCert", "ClientCertPassword":
        // Only include certificate options for service principal login
        return o.LoginMethod == "spn"
    case "Username", "Password":
        // Only include username/password for ROPC login
        return o.LoginMethod == "ropc"
    case "FederatedTokenFile":
        // Only include federated token for workload identity
        return o.LoginMethod == "workloadidentity"
    case "NewOption":
        // Your custom logic here
        return o.LoginMethod == "your-login-method"
    }
    return true // Include by default
}
```

### 5. Environment Variable Support

The unified system supports multiple environment variable names with fallbacks:

```go
// Single environment variable
ClientID string `flag:"client-id" env:"AZURE_CLIENT_ID"`

// Multiple environment variables (tries in order)
ClientSecret string `flag:"client-secret" env:"AZURE_CLIENT_SECRET,AAD_SERVICE_PRINCIPAL_CLIENT_SECRET"`

// Support for Terraform Azure Provider variables
TenantID string `flag:"tenant-id" env:"AZURE_TENANT_ID,ARM_TENANT_ID"`
```

### 6. Sensitive Options

For options containing secrets, use the `sensitive` tag:

```go
ClientSecret string `flag:"client-secret" env:"AZURE_CLIENT_SECRET" sensitive:"true" description:"Client secret for authentication"`
```

This prevents the value from being logged or displayed in debug output.

### 7. Advanced Conversion Logic

For complex conversion scenarios, you can override the automatic behavior in `buildExecConfig()`:

```go
func (o *UnifiedOptions) buildExecConfig(authInfo *api.AuthInfo) (*api.ExecConfig, error) {
    args := []string{"get-token"}
    
    // ... automatic field processing ...
    
    // Custom argument handling
    if o.NewOption != "" && o.someCondition() {
        args = append(args, "--new-option", o.transformValue(o.NewOption))
    }
    
    // Special environment variables
    var envVars []api.ExecEnvVar
    if o.NewOption == "special" {
        envVars = append(envVars, api.ExecEnvVar{
            Name:  "SPECIAL_ENV_VAR",
            Value: o.deriveSpecialValue(),
        })
    }
    
    return &api.ExecConfig{
        APIVersion: execAPIVersion,
        Command:    execName,
        Args:       args,
        Env:        envVars,
        InstallHint: execInstallHint,
    }, nil
}
```

## Testing Your Changes

### 1. Unit Tests

Add tests to the appropriate test file in `pkg/internal/options/`:

```go
func TestNewOptionValidation(t *testing.T) {
    tests := []struct {
        name        string
        options     *UnifiedOptions
        expectError bool
    }{
        {
            name: "valid new option",
            options: &UnifiedOptions{
                NewOption: "valid-value",
                // ... other required fields
            },
            expectError: false,
        },
        {
            name: "invalid new option",
            options: &UnifiedOptions{
                NewOption: "invalid-value",
                // ... other required fields
            },
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.options.Validate()
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Integration Tests

Add test cases to `test/integration/converter_integration_test.go`:

```go
{
    name:          "convert with new option",
    fixture:       "sample_kubeconfig.yaml",
    args:          []string{"--login", "devicecode", "--new-option", "test-value", "--tenant-id", "test-tenant", "--client-id", "test-client", "--server-id", "test-server"},
    expectSuccess: true,
    validateResult: func(t *testing.T, originalContent, resultContent string) {
        var config Kubeconfig
        err := yaml.Unmarshal([]byte(resultContent), &config)
        require.NoError(t, err)
        
        user := config.Users[0]
        args := user.User.Exec.Args
        assert.Contains(t, args, "--new-option")
        assert.Contains(t, args, "test-value")
    },
},
```

### 3. Manual Testing

Test both modes to ensure compatibility:

```bash
# Legacy mode
./kubelogin convert-kubeconfig --help

# Unified mode  
KUBELOGIN_USE_UNIFIED_OPTIONS=true ./kubelogin convert-kubeconfig --help

# Test your new option
KUBELOGIN_USE_UNIFIED_OPTIONS=true ./kubelogin convert-kubeconfig --new-option test-value --help
```

## Common Patterns

### 1. Boolean Flags

```go
EnableFeature bool `flag:"enable-feature" description:"Enable special feature"`
```

### 2. Duration Options

```go
Timeout time.Duration `flag:"timeout" env:"AZURE_TIMEOUT" default:"60s" description:"Request timeout"`
```

### 3. File Path Options

```go
ConfigFile string `flag:"config-file" env:"AZURE_CONFIG_FILE" description:"Path to configuration file"`
```

### 4. Enum/Choice Options

```go
LogLevel string `flag:"log-level" env:"AZURE_LOG_LEVEL" validate:"oneof=debug info warn error" default:"info" description:"Logging level"`
```

## Migration Notes

### Enabling Legacy Options

The unified options system is the default behavior. To use the legacy options system, set the `KUBELOGIN_USE_LEGACY_OPTIONS` environment variable:

```bash
# Use legacy options
export KUBELOGIN_USE_LEGACY_OPTIONS=true

# Or per-command
KUBELOGIN_USE_LEGACY_OPTIONS=true ./kubelogin convert-kubeconfig --help
```

### Backward Compatibility

The unified system is fully backward compatible and is now the default:
- All existing options work identically
- All existing environment variables are supported
- All existing validation behavior is preserved
- Legacy mode is available via `KUBELOGIN_USE_LEGACY_OPTIONS=true`

## Troubleshooting

### Common Issues

1. **Option not appearing in help**: Check the `flag` struct tag is present and correct
2. **Environment variable not working**: Verify the `env` tag and test with `env | grep AZURE`
3. **Validation not working**: Check the `validate` tag syntax and test with invalid values
4. **Option not in kubeconfig**: Verify `shouldIncludeArg()` returns true for your option

### Debug Commands

```bash
# Check flag registration (unified is now default)
./kubelogin convert-kubeconfig --help | grep "new-option"

# Test validation
./kubelogin convert-kubeconfig --new-option invalid-value

# Test environment variables
AZURE_NEW_OPTION=test-value ./kubelogin convert-kubeconfig --help

# Test legacy mode if needed
KUBELOGIN_USE_LEGACY_OPTIONS=true ./kubelogin convert-kubeconfig --help
```

## Summary

The unified options system makes adding new CLI options significantly easier:

- **1 touch point** instead of 6+ for most options
- **Automatic registration** of flags, environment variables, and help text
- **Command-specific flags** using struct tags for clean separation
- **Centralized validation** with clear error messages
- **Automatic conversion** to kubeconfig arguments
- **Full backward compatibility** with existing code

### Architecture Simplifications

Recent improvements have streamlined the validation architecture:

- **Two focused validation methods**: `ValidateForTokenExecution()` (strict) and `ValidateForConversion()` (lenient)
- **Eliminated redundant validation**: Removed `validateExecConfig()` which duplicated struct-level validation
- **Simplified credential creation**: Direct calls to `token.NewAzIdentityCredential()` instead of complex builder patterns
- **Clear file organization**: All validation logic consolidated in `validation.go`

By following these patterns, you can add new options quickly and safely while maintaining the high quality and consistency of the kubelogin codebase.

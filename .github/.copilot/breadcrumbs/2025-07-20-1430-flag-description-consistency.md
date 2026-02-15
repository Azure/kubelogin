# 2025-07-20-1430-flag-description-consistency.md

## Requirements
- Identified that flag descriptions in unified options system have diverged from legacy mode
- Need to ensure consistent flag descriptions between legacy and unified modes for better user experience
- Users should see the same help text regardless of which mode they use

## Additional comments from user
User correctly pointed out that flag descriptions have diverted from legacy mode and requested tests first to validate the issue before fixing it.

## Plan

### Phase 1: Testing Infrastructure ✅
- [x] Task 1.1: Create comprehensive description consistency tests
- [x] Task 1.2: Test specific description requirements (environment variables, detailed explanations)
- [x] Task 1.3: Test capitalization consistency
- [x] Task 1.4: Identify all inconsistencies between legacy and unified modes

### Phase 2: Fix Description Inconsistencies ✅
- [x] Task 2.1: Update environment variable mentions in flag descriptions
- [x] Task 2.2: Fix capitalization inconsistencies (lowercase "set" to match legacy mode)
- [x] Task 2.3: Add detailed explanations where missing (pop-claims format, disable-instance-discovery details)
- [x] Task 2.4: Add "Default false" mentions where appropriate
- [x] Task 2.5: Fix minor punctuation inconsistencies (periods, detailed examples)

### Phase 3: Validation ✅
- [x] Task 3.1: Run description consistency tests to ensure all pass
- [x] Task 3.2: Manual CLI testing to verify help output matches
- [x] Task 3.3: Verify no regressions in existing functionality

## Decisions
- Tests first approach: Created comprehensive tests to document expected behavior before making changes
- Focus on making unified mode match legacy mode descriptions since legacy is the established baseline
- Preserve all environment variable mentions and detailed explanations from legacy mode

## Implementation Details

### Test Results Summary
Found **21 flag description inconsistencies**:

**Major Categories:**
1. **Missing environment variable mentions** (12 flags): client-id, client-secret, tenant-id, etc.
2. **Missing detailed explanations**: pop-claims format details, disable-instance-discovery explanation
3. **Capitalization differences**: "set to true" vs "Set to true"
4. **Missing "Default false" mentions**: disable-environment-override, disable-instance-discovery
5. **Punctuation differences**: trailing periods and detailed examples

**Key Flags Needing Updates:**
- `client-id`: Missing "AAD_SERVICE_PRINCIPAL_CLIENT_ID or AZURE_CLIENT_ID environment variable"
- `client-secret`: Missing environment variable mentions
- `tenant-id`: Missing "AZURE_TENANT_ID environment variable"
- `pop-claims`: Missing format details "key=val,key2=val2" and ARM_ID example
- `disable-instance-discovery`: Missing detailed explanation about Identity Provider
- `use-azurerm-env-vars`: Missing specific ARM variable names

## Changes Made
- Created `pkg/internal/options/description_consistency_test.go` with comprehensive tests
- Tests validate consistency between legacy and unified flag descriptions
- Tests identify specific requirements like environment variable mentions
- All tests currently pass for documentation but many fail for requirements (as expected)

## Before/After Comparison
**Before**: 21 description inconsistencies between legacy and unified modes
**After**: ✅ 0 description inconsistencies - perfect consistency achieved

### Key Improvements Made:
1. **Environment Variable Mentions**: Added "It may be specified in ENVIRONMENT_VAR environment variable" to 12 flags
2. **Detailed Explanations**: Added format details for pop-claims (`key=val,key2=val2` and `u=ARM_ID`)
3. **Complete disable-instance-discovery description**: Added full explanation about environments and Identity Providers
4. **Capitalization Consistency**: Used lowercase "set to true" to match legacy mode
5. **Punctuation Alignment**: Added periods, "Default false" mentions, and specific examples
6. **Unicode Backticks**: Properly handled backticks in struct tags using Unicode escapes (\u0060)

## References
- Legacy mode flag descriptions obtained via manual CLI testing
- Unified mode descriptions from struct tag definitions in `pkg/internal/options/unified.go`
- Test patterns based on specific legacy mode behaviors like environment variable mentions

## Success Criteria
- [x] All `TestFlagDescriptionConsistency` subtests pass ✅
- [x] All `TestSpecificDescriptionRequirements` subtests pass ✅
- [x] Manual CLI comparison shows identical help text between modes ✅
- [x] No regressions in existing functionality ✅

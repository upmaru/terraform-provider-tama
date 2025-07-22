# JSON Normalization Fix for tama_class schema_json

## Problem

The Terraform provider was producing "inconsistent result after apply" errors when using the `schema_json` attribute of the `tama_class` resource. This occurred because:

1. **During planning**: The `JSONNormalize()` plan modifier normalized JSON formatting to prevent unnecessary diffs
2. **During apply**: The provider received a response from the API and marshaled it back to JSON using `json.Marshal()`, which could produce different key ordering or formatting than what was planned

## Example Error

```
Error: Provider produced inconsistent result after apply

When applying changes to module.global.tama_class.action-call, provider 
"provider[\"registry.terraform.io/upmaru/tama\"]" produced an unexpected 
new value: .schema_json: was cty.StringVal("{\n \"title\": \"action-call\",\n ...") 
but now cty.StringVal("{\"description\":\"An action call...\"}").
```

## Root Cause

The inconsistency occurred in the `Create`, `Read`, and `Update` functions where the provider would:

1. Send schema data to the API
2. Receive a response with the same semantic content
3. Marshal the response using `json.Marshal()` 
4. Set `data.SchemaJSON = types.StringValue(string(schemaJSON))`

The problem was that `json.Marshal()` produces deterministic but potentially differently-ordered JSON compared to what the plan modifier normalized during planning.

## Solution

The fix ensures that after receiving API responses, the JSON is normalized using the same `NormalizeJSON()` function that the plan modifier uses:

### Changes Made

1. **Exported NormalizeJSON function** in `internal/planmodifier/json.go`
   - Changed `normalizeJSON` to `NormalizeJSON` (exported)
   - Updated internal usage and tests

2. **Updated resource operations** in `tama/neural/class/resource.go`
   - Modified `Create()`, `Read()`, and `Update()` functions
   - Added normalization step after marshaling API responses:
   ```go
   // Before (problematic)
   data.SchemaJSON = types.StringValue(string(schemaJSON))
   
   // After (fixed)
   normalizedJSON, err := internalplanmodifier.NormalizeJSON(string(schemaJSON))
   if err != nil {
       resp.Diagnostics.AddError("Schema Error", fmt.Sprintf("Unable to normalize schema JSON: %s", err))
       return
   }
   data.SchemaJSON = types.StringValue(normalizedJSON)
   ```

### Benefits

- **Consistency**: Ensures the same JSON normalization is applied during both planning and apply phases
- **Reliability**: Eliminates "inconsistent result" errors caused by formatting differences
- **Maintainability**: Uses the same normalization logic across the provider

## Testing

Added comprehensive tests in `tama/neural/class/resource_test.go`:

- `TestJSONNormalizationConsistency`: Verifies normalization produces consistent output
- `TestResourceJSONConsistency`: Tests the resource-level JSON handling 
- `TestOriginalErrorScenario`: Reproduces and validates the fix for the original error

## Files Modified

- `internal/planmodifier/json.go` - Exported NormalizeJSON function
- `internal/planmodifier/json_test.go` - Updated test to use exported function
- `tama/neural/class/resource.go` - Added normalization to Create/Read/Update
- `tama/neural/class/resource_test.go` - Added comprehensive tests

## Verification

Run the tests to verify the fix:

```bash
# Test the plan modifier
go test -v ./internal/planmodifier

# Test the resource normalization
go test -v ./tama/neural/class -run TestJSONNormalizationConsistency

# Test the original error scenario  
go test -v ./tama/neural/class -run TestOriginalErrorScenario
```

This fix ensures that `schema_json` fields maintain consistent formatting throughout the Terraform lifecycle, preventing inconsistent result errors while preserving semantic correctness.
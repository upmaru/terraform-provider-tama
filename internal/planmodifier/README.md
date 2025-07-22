# Plan Modifiers

This package contains shared plan modifiers for the Tama Terraform provider.

## JSON Normalization Plan Modifier

The `JSONNormalize()` plan modifier prevents unnecessary updates when JSON strings are semantically equivalent but formatted differently.

### Problem

When working with JSON string attributes in Terraform, users might provide nicely formatted JSON:

```json
{
  "title": "dynamic",
  "description": "A dynamic schema",
  "properties": {
    "entity": {
      "description": "The record of the entity",
      "type": "object"
    }
  },
  "type": "object"
}
```

But the API server returns minified JSON:

```json
{"description":"A dynamic schema","properties":{"entity":{"description":"The record of the entity","type":"object"}},"title":"dynamic","type":"object"}
```

Without the plan modifier, Terraform sees these as different values and tries to "fix" the difference on every plan/apply cycle.

### Solution

The `JSONNormalize()` plan modifier compares JSON strings semantically rather than as raw strings. If the planned value and state value are semantically equivalent JSON, it suppresses the diff and keeps the existing state value.

### Usage

Import the plan modifier package and apply it to JSON string attributes:

```go
import (
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    internalplanmodifier "github.com/upmaru/terraform-provider-tama/internal/planmodifier"
)

// In your resource schema
"schema_json": schema.StringAttribute{
    MarkdownDescription: "JSON schema as a string",
    Optional:            true,
    PlanModifiers: []planmodifier.String{
        internalplanmodifier.JSONNormalize(),
    },
},
```

### Behavior

The plan modifier:

1. **Skips modification** if either value is null/unknown
2. **Returns early** if strings are already identical
3. **Normalizes both values** by parsing and re-marshaling as JSON
4. **Suppresses the diff** if normalized values are semantically equal
5. **Allows the change** if values are semantically different or if JSON parsing fails

### Example

```go
// User input (formatted JSON)
planValue := `{
  "key": "value",
  "number": 123
}`

// Server response (minified JSON)
stateValue := `{"number":123,"key":"value"}`

// Plan modifier will suppress the diff because they're semantically equal
```

### Testing

The plan modifier is thoroughly tested with various scenarios:

- Identical strings
- Semantically equal but differently formatted JSON
- Different JSON content
- Invalid JSON in plan or state
- Null/unknown values
- Empty strings vs empty objects

Run tests with:

```bash
go test ./internal/planmodifier -v
```

### Resources Using This Plan Modifier

- `tama_class.schema_json` - JSON schema definition
- `tama_class.schema.properties` - JSON properties within schema blocks  
- `tama_model.parameters` - Model parameters as JSON

This ensures consistent behavior across all JSON string fields in the provider.
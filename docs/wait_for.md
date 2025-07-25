# Wait For Functionality

The `wait_for` block allows you to specify conditions that must be met before Terraform considers the resource creation or update complete. This is particularly useful when you need to wait for asynchronous processes to complete or for certain states to be reached.

## Basic Usage

```hcl
resource "tama_specification" "example" {
  space_id = "space-123"
  schema   = "..."
  version  = "1.0.0"
  endpoint = "https://api.example.com"

  wait_for {
    field {
      key   = "current_state"
      value = "active"
    }
  }
}
```

## Schema

The `wait_for` block supports the following nested blocks:

### `field` Block

The `field` block defines a condition that must be satisfied. You can specify multiple `field` blocks within a single `wait_for` block, and all conditions must be met.

#### Arguments

- `key` (Required) - The JSON path to the field you want to check in the API response. Uses dot notation for nested fields (e.g., `metadata.status`, `config.deployment.state`).

- `value` (Required) - The expected value that the field should match.

- `value_type` (Optional) - The type of comparison to perform. Defaults to `"eq"`.
  - `"eq"` - Exact string equality (default)
  - `"regex"` - Regular expression matching

## Examples

### Simple State Waiting

Wait for a specification to reach the "completed" state:

```hcl
resource "tama_specification" "example" {
  # ... other configuration ...

  wait_for {
    field {
      key   = "current_state"
      value = "completed"
    }
  }
}
```

### Multiple Conditions

Wait for multiple conditions to be satisfied:

```hcl
resource "tama_specification" "example" {
  # ... other configuration ...

  wait_for {
    field {
      key   = "current_state"
      value = "completed"
    }

    field {
      key   = "provision_state"
      value = "active"
    }
  }
}
```

### Regular Expression Matching

Use regex to match against multiple possible values:

```hcl
resource "tama_specification" "example" {
  # ... other configuration ...

  wait_for {
    field {
      key        = "provision_state"
      value      = "^(active|inactive)$"
      value_type = "regex"
    }
  }
}
```

### Nested Field Access

Access nested fields using dot notation:

```hcl
resource "tama_specification" "example" {
  # ... other configuration ...

  wait_for {
    field {
      key   = "metadata.deployment.status"
      value = "completed"
    }

    field {
      key   = "config.health.status"
      value = "active"
    }
  }
}
```

## Behavior

### Timeout

The wait functionality has a default timeout of 10 minutes. If the conditions are not met within this timeframe, Terraform will fail with a timeout error.

### Polling Interval

The conditions are checked every 5 seconds until either:
- All conditions are satisfied (success)
- The timeout is reached (failure)

### Error Handling

If any of the following occurs, the wait will fail:
- The specified field/key doesn't exist in the API response
- An invalid regex pattern is provided
- The API call to fetch the specification fails
- The timeout is exceeded

### All Conditions Required

When multiple `field` blocks are specified, **all** conditions must be satisfied for the wait to complete successfully. If any single condition fails, the wait continues until timeout.

## JSON Path Examples

The `key` parameter supports standard JSON path notation:

| JSON Response | Key | Description |
|---------------|-----|-------------|
| `{"status": "active"}` | `status` | Top-level field |
| `{"config": {"state": "ready"}}` | `config.state` | Nested field |
| `{"metadata": {"deployment": {"phase": "complete"}}}` | `metadata.deployment.phase` | Deeply nested field |

## Common Use Cases

### Waiting for Deployment Completion

```hcl
wait_for {
  field {
    key   = "current_state"
    value = "completed"
  }

  field {
    key   = "provision_state"
    value = "active"
  }
}
```

### Waiting for Health Checks

```hcl
wait_for {
  field {
    key   = "health.status"
    value = "passing"
  }
}
```

### Waiting for Multiple Acceptable States

```hcl
wait_for {
  field {
    key        = "current_state"
    value      = "^(completed|failed)$"
    value_type = "regex"
  }
}
```

## Best Practices

1. **Use Specific Conditions**: Be as specific as possible with your wait conditions to avoid false positives.

2. **Combine State Checks**: When waiting for complex operations, check multiple related fields to ensure the resource is truly ready.

3. **Use Regex Sparingly**: While powerful, regex patterns can be harder to understand and debug. Use them only when you need to match multiple possible values.

4. **Monitor Timeouts**: If you frequently hit timeouts, consider whether your wait conditions are appropriate or if the underlying service needs more time.

5. **Test Wait Conditions**: Verify that your wait conditions work as expected by checking the actual API responses during testing.

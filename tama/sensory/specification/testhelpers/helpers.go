// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testhelpers

import (
	"encoding/json"
	"fmt"
)

// MustMarshalJSON marshals a Go data structure to JSON string and panics on error.
func MustMarshalJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return string(b)
}

func TestSchema() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Test API",
			"version":     "1.0.0",
			"description": "Test specification",
		},
		"paths": map[string]any{
			"/messages": map[string]any{
				"post": map[string]any{
					"summary": "Create message",
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"message": map[string]any{
											"type":        "string",
											"description": "A message property",
										},
										"count": map[string]any{
											"type":    "integer",
											"minimum": 0,
										},
									},
									"required": []string{"message"},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Success",
						},
					},
				},
			},
		},
	}
}

func TestSchemaUpdated() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Test API Updated",
			"version":     "1.1.0",
			"description": "Updated test specification",
		},
		"paths": map[string]any{
			"/messages": map[string]any{
				"post": map[string]any{
					"summary": "Create message",
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"message": map[string]any{
											"type":        "string",
											"description": "An updated message property",
										},
										"count": map[string]any{
											"type":    "integer",
											"minimum": 0,
										},
										"enabled": map[string]any{
											"type":    "boolean",
											"default": true,
										},
									},
									"required": []string{"message", "enabled"},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Success",
						},
					},
				},
			},
		},
	}
}

func TestComplexSchema() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Complex API",
			"version":     "1.0.0",
			"description": "Complex test specification with nested objects",
		},
		"paths": map[string]any{
			"/resources": map[string]any{
				"post": map[string]any{
					"summary": "Create resource",
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"metadata": map[string]any{
											"type": "object",
											"properties": map[string]any{
												"name": map[string]any{"type": "string"},
												"tags": map[string]any{
													"type":  "array",
													"items": map[string]any{"type": "string"},
												},
											},
										},
										"configuration": map[string]any{
											"type": "object",
											"properties": map[string]any{
												"settings": map[string]any{
													"type":                 "object",
													"additionalProperties": true,
												},
											},
										},
									},
									"required": []string{"metadata"},
								},
							},
						},
					},
					"responses": map[string]any{
						"201": map[string]any{
							"description": "Resource created",
						},
					},
				},
			},
		},
	}
}

func TestSchemaWithWhitespace() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   "Whitespace Test API",
			"version": "1.0.0",
		},
		"paths": map[string]any{
			"/test": map[string]any{
				"post": map[string]any{
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"message": map[string]any{
											"type": "string",
										},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "OK",
						},
					},
				},
			},
		},
	}
}

func TestSchemaCompact() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   "Whitespace Test API",
			"version": "1.0.0",
		},
		"paths": map[string]any{
			"/test": map[string]any{
				"post": map[string]any{
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"message": map[string]any{"type": "string"},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "OK"},
					},
				},
			},
		},
	}
}

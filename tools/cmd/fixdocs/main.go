// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	documentPath      = "../docs/resources/thought_tool_modifier.md"
	sourceDescription = "Required trusted metadata source to inject at the target. Configure this block exactly once. (see [below for nested schema](#nestedblock--source))"

	optionalSourceSection = "### Optional\n\n- `source` (Block, Optional) " + sourceDescription + "\n\n"
	requiredSourceSection = "### Required Blocks\n\n- `source` (Block, Required) " + sourceDescription + "\n\n"
)

func main() {
	contents, err := os.ReadFile(documentPath)
	if err != nil {
		panic(fmt.Errorf("read generated modifier documentation: %w", err))
	}

	document := string(contents)
	if strings.Count(document, optionalSourceSection) != 1 {
		panic("generated modifier documentation no longer contains the expected optional source section")
	}

	document = strings.Replace(document, optionalSourceSection, requiredSourceSection, 1)
	if err := os.WriteFile(documentPath, []byte(document), 0o644); err != nil {
		panic(fmt.Errorf("write corrected modifier documentation: %w", err))
	}
}

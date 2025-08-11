// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package operation_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/upmaru/terraform-provider-tama/internal/acceptance"
)

func TestAccClassOperationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClassOperationResourceConfig(fmt.Sprintf("test-operation-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "class_id"),
					resource.TestCheckResourceAttr("tama_class_operation.test", "node_type", "explicit"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "current_state"),
					resource.TestCheckResourceAttr("tama_class_operation.test", "chain_ids.#", "1"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "chain_ids.0"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "node_ids.#"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "tama_class_operation.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"chain_ids", "node_type", "current_state"}, // chain_ids and node_type might not be returned in read
				ImportStateIdFunc:       testAccClassOperationImportStateIdFunc,
			},
		},
	})
}

func TestAccClassOperationResource_ReactiveNodeType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassOperationResourceConfigReactive(fmt.Sprintf("test-operation-reactive-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "class_id"),
					resource.TestCheckResourceAttr("tama_class_operation.test", "node_type", "reactive"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "current_state"),
					resource.TestCheckResourceAttr("tama_class_operation.test", "chain_ids.#", "1"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "chain_ids.0"),
				),
			},
		},
	})
}

func TestAccClassOperationResource_MultipleChains(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassOperationResourceConfigMultipleChains(fmt.Sprintf("test-operation-multi-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "class_id"),
					resource.TestCheckResourceAttr("tama_class_operation.test", "node_type", "explicit"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "current_state"),
					resource.TestCheckResourceAttr("tama_class_operation.test", "chain_ids.#", "2"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "chain_ids.0"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "chain_ids.1"),
				),
			},
		},
	})
}

func TestAccClassOperationResource_InvalidNodeType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccClassOperationResourceConfigInvalidNodeType(fmt.Sprintf("test-operation-%d", time.Now().UnixNano())),
				ExpectError: regexp.MustCompile("Attribute node_type value must be one of"),
			},
		},
	})
}

func TestAccClassOperationResource_DefaultNodeType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassOperationResourceConfigDefaultNodeType(fmt.Sprintf("test-operation-default-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "id"),
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "class_id"),
					resource.TestCheckResourceAttr("tama_class_operation.test", "node_type", "reactive"), // Should default to reactive
					resource.TestCheckResourceAttrSet("tama_class_operation.test", "current_state"),
					resource.TestCheckResourceAttr("tama_class_operation.test", "chain_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccClassOperationResource_BaseInfrastructure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClassOperationResourceConfigBaseInfrastructure(fmt.Sprintf("test-infra-%d", time.Now().UnixNano())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tama_space.test", "id"),
					resource.TestCheckResourceAttrSet("tama_specification.test", "id"),
					resource.TestCheckResourceAttrSet("data.tama_class.test", "id"),
					resource.TestCheckResourceAttrSet("tama_chain.test", "id"),
					resource.TestCheckResourceAttrSet("tama_modular_thought.test", "id"),
					resource.TestCheckResourceAttrSet("tama_thought_path.test", "id"),
					resource.TestCheckResourceAttrSet("tama_node.handle-extraction", "id"),
				),
			},
		},
	})
}

func testAccClassOperationResourceConfig(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
data "tama_space" "global" {
  id = "global"
}

data "tama_class" "class-proxy" {
  space_id = data.tama_space.global.id
  name = "class-proxy"
}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_specification" "test" {
  space_id = tama_space.test.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

data "tama_class" "test" {
  specification_id = tama_specification.test.id
  name             = "create-index"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  index    = 0
  relation = "extraction"

  module {
    reference = "tama/classes/extraction"
    parameters = jsonencode({
      types = ["array"]
      depth = 1
    })
  }
}

resource "tama_thought_path" "test" {
  thought_id      = tama_modular_thought.test.id
  target_class_id = data.tama_class.test.id
}

resource "tama_node" "handle-extraction" {
  space_id = tama_space.test.id
  class_id = data.tama_class.class-proxy.id
  chain_id = tama_chain.test.id

  type = "explicit"
}

resource "tama_class_operation" "test" {
  class_id  = data.tama_class.test.id
  chain_ids = [tama_chain.test.id]
  node_type = "explicit"

  depends_on = [tama_node.handle-extraction, tama_thought_path.test]
}
`, spaceName)
}

func testAccClassOperationResourceConfigReactive(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
data "tama_space" "global" {
  id = "global"
}

data "tama_class" "class-proxy" {
  space_id = data.tama_space.global.id
  name = "class-proxy"
}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_specification" "test" {
  space_id = tama_space.test.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

data "tama_class" "test" {
  specification_id = tama_specification.test.id
  name             = "create-index"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Reactive Processing Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  index    = 0
  relation = "extraction"

  module {
    reference = "tama/classes/extraction"
    parameters = jsonencode({
      types = ["object"]
      depth = 2
    })
  }
}

resource "tama_thought_path" "test" {
  thought_id      = tama_modular_thought.test.id
  target_class_id = data.tama_class.test.id
}

resource "tama_node" "handle-extraction" {
  space_id = tama_space.test.id
  class_id = data.tama_class.class-proxy.id
  chain_id = tama_chain.test.id

  type = "reactive"
}

resource "tama_class_operation" "test" {
  class_id  = data.tama_class.test.id
  chain_ids = [tama_chain.test.id]
  node_type = "reactive"

  depends_on = [tama_node.handle-extraction]
}
`, spaceName)
}

func testAccClassOperationResourceConfigMultipleChains(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
data "tama_space" "global" {
  id = "global"
}

data "tama_class" "class-proxy" {
  space_id = data.tama_space.global.id
  name = "class-proxy"
}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_specification" "test" {
  space_id = tama_space.test.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

data "tama_class" "test" {
  specification_id = tama_specification.test.id
  name             = "create-index"
}

resource "tama_chain" "test1" {
  space_id = tama_space.test.id
  name     = "First Processing Chain"
}

resource "tama_chain" "test2" {
  space_id = tama_space.test.id
  name     = "Second Processing Chain"
}

resource "tama_modular_thought" "test1" {
  chain_id = tama_chain.test1.id
  index    = 0
  relation = "extraction"

  module {
    reference = "tama/classes/extraction"
    parameters = jsonencode({
      types = ["array"]
      depth = 1
    })
  }
}

resource "tama_modular_thought" "test2" {
  chain_id = tama_chain.test2.id
  index    = 0
  relation = "transformation"

  module {
    reference = "tama/classes/extraction"
    parameters = jsonencode({
      types = ["object"]
      depth = 1
    })
  }
}

resource "tama_thought_path" "test1" {
  thought_id      = tama_modular_thought.test1.id
  target_class_id = data.tama_class.test.id
}

resource "tama_thought_path" "test2" {
  thought_id      = tama_modular_thought.test2.id
  target_class_id = data.tama_class.test.id
}

resource "tama_node" "handle-extraction-1" {
  space_id = tama_space.test.id
  class_id = data.tama_class.class-proxy.id
  chain_id = tama_chain.test1.id

  type = "explicit"
}

resource "tama_node" "handle-extraction-2" {
  space_id = tama_space.test.id
  class_id = data.tama_class.class-proxy.id
  chain_id = tama_chain.test2.id

  type = "explicit"
}

resource "tama_class_operation" "test" {
  class_id  = data.tama_class.test.id
  chain_ids = [tama_chain.test1.id, tama_chain.test2.id]
  node_type = "explicit"

  depends_on = [tama_node.handle-extraction-1, tama_node.handle-extraction-2, tama_thought_path.test1, tama_thought_path.test2]
}
`, spaceName)
}

func testAccClassOperationResourceConfigInvalidNodeType(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
data "tama_space" "global" {
  id = "global"
}

data "tama_class" "class-proxy" {
  space_id = data.tama_space.global.id
  name = "class-proxy"
}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_specification" "test" {
  space_id = tama_space.test.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

data "tama_class" "test" {
  specification_id = tama_specification.test.id
  name             = "create-index"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Test Processing Chain"
}

resource "tama_node" "handle-extraction" {
  space_id = tama_space.test.id
  class_id = data.tama_class.class-proxy.id
  chain_id = tama_chain.test.id

  type = "explicit"
}

resource "tama_class_operation" "test" {
  class_id  = data.tama_class.test.id
  chain_ids = [tama_chain.test.id]
  node_type = "invalid_type"

  depends_on = [tama_node.handle-extraction]
}
`, spaceName)
}

func testAccClassOperationResourceConfigDefaultNodeType(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
data "tama_space" "global" {
  id = "global"
}

data "tama_class" "class-proxy" {
  space_id = data.tama_space.global.id
  name = "class-proxy"
}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_specification" "test" {
  space_id = tama_space.test.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

data "tama_class" "test" {
  specification_id = tama_specification.test.id
  name             = "create-index"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Default Processing Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  index    = 0
  relation = "extraction"

  module {
    reference = "tama/classes/extraction"
    parameters = jsonencode({
      types = ["array"]
      depth = 1
    })
  }
}

resource "tama_thought_path" "test" {
  thought_id      = tama_modular_thought.test.id
  target_class_id = data.tama_class.test.id
}

resource "tama_node" "handle-extraction" {
  space_id = tama_space.test.id
  class_id = data.tama_class.class-proxy.id
  chain_id = tama_chain.test.id

  type = "reactive"
}

resource "tama_class_operation" "test" {
  class_id  = data.tama_class.test.id
  chain_ids = [tama_chain.test.id]
  # node_type is not specified, should default to "reactive"

  depends_on = [tama_node.handle-extraction, tama_thought_path.test]
}
`, spaceName)
}

func testAccClassOperationResourceConfigBaseInfrastructure(spaceName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
data "tama_space" "global" {
  id = "global"
}

data "tama_class" "class-proxy" {
  space_id = data.tama_space.global.id
  name = "class-proxy"
}

resource "tama_space" "test" {
  name = "%s"
  type = "root"
}

resource "tama_specification" "test" {
  space_id = tama_space.test.id
  version  = "1.0.0"
  endpoint = "https://elasticsearch.arrakis.upmaru.network"
  schema   = jsonencode(jsondecode(file("${path.module}/testdata/elasticsearch_schema.json")))

  wait_for {
    field {
      name = "current_state"
      in   = ["completed"]
    }
  }
}

data "tama_class" "test" {
  specification_id = tama_specification.test.id
  name             = "create-index"
}

resource "tama_chain" "test" {
  space_id = tama_space.test.id
  name     = "Default Processing Chain"
}

resource "tama_modular_thought" "test" {
  chain_id = tama_chain.test.id
  index    = 0
  relation = "extraction"

  module {
    reference = "tama/classes/extraction"
    parameters = jsonencode({
      types = ["array"]
      depth = 1
    })
  }
}

resource "tama_thought_path" "test" {
  thought_id      = tama_modular_thought.test.id
  target_class_id = data.tama_class.test.id
}

resource "tama_node" "handle-extraction" {
  space_id = tama_space.test.id
  class_id = data.tama_class.class-proxy.id
  chain_id = tama_chain.test.id

  type = "explicit"
}
`, spaceName)
}

func testAccClassOperationImportStateIdFunc(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["tama_class_operation.test"]
	if !ok {
		return "", fmt.Errorf("not found: %s", "tama_class_operation.test")
	}

	classId := rs.Primary.Attributes["class_id"]
	operationId := rs.Primary.Attributes["id"]

	return fmt.Sprintf("%s:%s", classId, operationId), nil
}

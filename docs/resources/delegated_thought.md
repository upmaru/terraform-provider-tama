---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tama_delegated_thought Resource - tama"
subcategory: ""
description: |-
  Manages a Tama Perception Delegated Thought resource
---

# tama_delegated_thought (Resource)

Manages a Tama Perception Delegated Thought resource



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `chain_id` (String) ID of the chain this thought belongs to

### Optional

- `delegation` (Block, Optional) Delegation configuration for the thought (see [below for nested schema](#nestedblock--delegation))
- `index` (Number) Index position of the thought in the chain
- `output_class_id` (String) ID of the output class for this thought

### Read-Only

- `id` (String) Delegated thought identifier
- `provision_state` (String) Current state of the thought
- `relation` (String) Relation type for the thought

<a id="nestedblock--delegation"></a>
### Nested Schema for `delegation`

Required:

- `target_thought_id` (String) Target thought ID from tama_modular_thought

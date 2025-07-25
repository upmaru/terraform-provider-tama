# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    tama = {
      source  = "upmaru/tama"
      version = "~> 0.1"
    }
  }
}

provider "tama" {}

module "global" {
  source  = "upmaru/base/tama"
  version = "0.1.7"
}

// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package v1

const (
	Schema     = "https://docs.oasis-open.org/openeox/tbd/schema/shell.json"
	CoreSchema = "https://docs.oasis-open.org/openeox/eox-core/v1.0/schema/core.json"

	// CoreSchemaLegacy is the core schema URI used by OpenEoX drafts prior
	// to CSD01 RC3. Documents declaring it are still accepted when parsing.
	CoreSchemaLegacy = "https://docs.oasis-open.org/openeox/v1.0/schema/core.json"

	ProductSoftwareSchema             = "https://docs.oasis-open.org/openeox/tbd/schema/product_software.json"
	ProductHardwareSchema             = "https://docs.oasis-open.org/openeox/tbd/schema/product_hardware.json"
	ProductHardwareWithSoftwareSchema = "https://docs.oasis-open.org/openeox/tbd/schema/product_hardware_with_software.json"
)

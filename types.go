// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import latest "github.com/carabiner-dev/openeox/types/v1"

const (
	Schema     = latest.Schema
	CoreSchema = latest.CoreSchema

	ProductSoftwareSchema             = latest.ProductSoftwareSchema
	ProductHardwareSchema             = latest.ProductHardwareSchema
	ProductHardwareWithSoftwareSchema = latest.ProductHardwareWithSoftwareSchema
)

type (
	Shell                       = latest.Shell
	Core                        = latest.Core
	Statement                   = latest.Statement
	Product                     = latest.Product
	ProductIdentificationHelper = latest.ProductIdentificationHelper
	CryptographicHashes         = latest.CryptographicHashes
	FileHash                    = latest.FileHash
	GenericURI                  = latest.GenericURI
)

func NewShell() *Shell {
	return &Shell{
		Schema:     Schema,
		Statements: []*Statement{},
	}
}

func NewStatement() *Statement {
	return &Statement{
		Core: NewCore(),
	}
}

func NewCore() *Core {
	return &Core{
		Schema: CoreSchema,
	}
}

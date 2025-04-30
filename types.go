// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import latest "github.com/carabiner-dev/openeox/types/tbd"

const (
	Schema     = latest.Schema
	CoreSchema = latest.CoreSchema
)

type (
	Shell     = latest.Shell
	Core      = latest.Core
	Statement = latest.Statement
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

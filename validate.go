// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ErrInvalidDocument is wrapped in the errors returned when a parsed
// document does not conform to the OpenEoX core JSON schema.
var ErrInvalidDocument = errors.New("document does not conform to the OpenEoX core schema")

// The OpenEoX core schema (CSD01 RC3) and the meta schema it declares,
// copied verbatim from oasis-tcs/openeox at commit fd771fe.
var (
	//go:embed schema/core.json
	coreSchemaData []byte

	//go:embed schema/meta.json
	metaSchemaData []byte
)

// metaSchemaURL is the $schema URI declared by the upstream core schema.
const metaSchemaURL = "https://docs.oasis-open.org/openeox/eox-core/v1.0/schema/meta.json"

// coreSchemaAliases lists superseded schema URIs still accepted in the
// $schema field of core documents. Documents declaring them are validated
// as if they declared CoreSchema.
var coreSchemaAliases = map[string]struct{}{
	CoreSchemaLegacy: {},
	"https://docs.oasis-open.org/openeox/tbd/schema/core.json": {},
}

var (
	coreValidator     *jsonschema.Schema
	coreValidatorOnce sync.Once
	errCoreValidator  error
)

// compileCoreValidator compiles the embedded core schema on first use.
func compileCoreValidator() (*jsonschema.Schema, error) {
	coreValidatorOnce.Do(func() {
		compiler := jsonschema.NewCompiler()
		compiler.Draft = jsonschema.Draft2020
		compiler.AssertFormat = true
		if err := compiler.AddResource(metaSchemaURL, bytes.NewReader(metaSchemaData)); err != nil {
			errCoreValidator = fmt.Errorf("loading meta schema: %w", err)
			return
		}
		if err := compiler.AddResource(CoreSchema, bytes.NewReader(coreSchemaData)); err != nil {
			errCoreValidator = fmt.Errorf("loading core schema: %w", err)
			return
		}
		coreValidator, errCoreValidator = compiler.Compile(CoreSchema)
	})
	return coreValidator, errCoreValidator
}

// validateCore checks a raw core document against the upstream OpenEoX
// core JSON schema before it is unmarshaled into the Core type.
func validateCore(data []byte) error {
	schema, err := compileCoreValidator()
	if err != nil {
		return err
	}

	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("decoding core document: %w", err)
	}

	if obj, ok := doc.(map[string]any); ok {
		if s, ok := obj["$schema"].(string); ok {
			if _, isAlias := coreSchemaAliases[s]; isAlias {
				obj["$schema"] = CoreSchema
			}
		}
	}

	if err := schema.Validate(doc); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDocument, err)
	}
	return nil
}

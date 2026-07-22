// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// coreDoc builds a core document declaring the given schema URI with the
// supplied body fields.
func coreDoc(schemaURI, fields string) []byte {
	return []byte(`{"$schema": "` + schemaURI + `", ` + fields + `}`)
}

const validCoreFields = `"end_of_life": "2027-12-31T23:59:59Z",
	"end_of_security_support": "2027-06-30T23:59:59Z",
	"last_updated": "2025-04-30T10:00:00Z"`

func TestParseCoreValidation(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	for _, tc := range []struct {
		name    string
		doc     []byte
		mustErr bool
	}{
		{"valid", coreDoc(CoreSchema, validCoreFields), false},
		{"legacy-schema-uri", coreDoc(CoreSchemaLegacy, validCoreFields), false},
		{"draft-schema-uri", coreDoc("https://docs.oasis-open.org/openeox/tbd/schema/core.json", validCoreFields), false},
		{"tba-values", coreDoc(CoreSchema, `"end_of_life": "tba",
			"end_of_security_support": "tba",
			"last_updated": "2025-04-30T10:00:00Z"`), false},
		{"all-fields", coreDoc(CoreSchema, validCoreFields+`,
			"end_of_sales": "2026-12-31T23:59:59Z",
			"general_availability": "2020-03-15T00:00:00Z"`), false},
		{"missing-last-updated", coreDoc(CoreSchema, `"end_of_life": "tba",
			"end_of_security_support": "tba"`), true},
		{"missing-schema", []byte(`{` + validCoreFields + `}`), true},
		{"unknown-schema-uri", coreDoc("https://example.com/other.json", validCoreFields), true},
		{"tba-last-updated", coreDoc(CoreSchema, `"end_of_life": "tba",
			"end_of_security_support": "tba",
			"last_updated": "tba"`), true},
		{"bad-timestamp", coreDoc(CoreSchema, `"end_of_life": "someday",
			"end_of_security_support": "2027-06-30T23:59:59Z",
			"last_updated": "2025-04-30T10:00:00Z"`), true},
		{"extra-property", coreDoc(CoreSchema, validCoreFields+`, "extra": true`), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core, err := parser.ParseCore(tc.doc)
			if tc.mustErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidDocument)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, core)
		})
	}
}

func TestParseCoreSkipValidation(t *testing.T) {
	parser := &Parser{SkipValidation: true}

	// Missing last_updated fails validation but parses when it is disabled.
	doc := coreDoc(CoreSchema, `"end_of_life": "tba", "end_of_security_support": "tba"`)

	_, err := (&Parser{}).ParseCore(doc)
	require.ErrorIs(t, err, ErrInvalidDocument)

	core, err := parser.ParseCore(doc)
	require.NoError(t, err)
	require.True(t, IsTBA(core.GetEndOfLife()))
}

func TestParseCorePreservesLegacySchema(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	// Validation treats legacy URIs as aliases but the parsed document
	// keeps the schema it declared.
	core, err := parser.ParseCore(coreDoc(CoreSchemaLegacy, validCoreFields))
	require.NoError(t, err)
	require.Equal(t, CoreSchemaLegacy, core.GetSchema())
}

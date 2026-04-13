// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const upstreamRepoEnv = "OPENEOX_UPSTREAM_REPO"

func upstreamRepoPath(t *testing.T) string {
	t.Helper()
	p := os.Getenv(upstreamRepoEnv)
	if p == "" {
		t.Skipf("%s not set, skipping upstream conformance tests", upstreamRepoEnv)
	}
	return p
}

// loadCoreSchemaValidator compiles the upstream OpenEoX core JSON schema.
func loadCoreSchemaValidator(t *testing.T, repoPath string) *jsonschema.Schema {
	t.Helper()
	metaPath := filepath.Join(repoPath, "eox-core-v-1-0", "schema", "meta.json")
	corePath := filepath.Join(repoPath, "eox-core-v-1-0", "schema", "eox-core.json")

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	compiler.AssertFormat = true

	metaURL := "https://docs.oasis-open.org/openeox/eox-core/v1.0/schema/meta.json"
	metaFile, err := os.Open(metaPath)
	require.NoError(t, err)
	defer metaFile.Close() //nolint:errcheck
	require.NoError(t, compiler.AddResource(metaURL, metaFile))

	coreURL := "https://docs.oasis-open.org/openeox/v1.0/schema/core.json"
	coreFile, err := os.Open(corePath)
	require.NoError(t, err)
	defer coreFile.Close() //nolint:errcheck
	require.NoError(t, compiler.AddResource(coreURL, coreFile))

	schema, err := compiler.Compile(coreURL)
	require.NoError(t, err)
	return schema
}

// TestConformanceParseCoreExamples parses the upstream openeox repo's example
// core documents and verifies the values are read correctly.
func TestConformanceParseCoreExamples(t *testing.T) {
	repoPath := upstreamRepoPath(t)
	parser, err := NewParser()
	require.NoError(t, err)

	examplesDir := filepath.Join(repoPath, "eox-core-v-1-0", "examples")
	entries, err := os.ReadDir(examplesDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(examplesDir, entry.Name()))
			require.NoError(t, err)

			core, err := parser.ParseCore(data)
			require.NoError(t, err)

			require.Equal(t, CoreSchema, core.GetSchema())
			require.NotNil(t, core.GetEndOfLife(), "end_of_life must be set")
			require.NotNil(t, core.GetEndOfSecuritySupport(), "end_of_security_support must be set")
			require.NotNil(t, core.GetLastUpdated(), "last_updated must be set")
		})
	}
}

// TestConformanceParseTBA verifies that "tba" values in upstream examples
// are parsed into the sentinel timestamp and correctly identified by IsTBA.
func TestConformanceParseTBA(t *testing.T) {
	repoPath := upstreamRepoPath(t)
	parser, err := NewParser()
	require.NoError(t, err)

	examplesDir := filepath.Join(repoPath, "eox-core-v-1-0", "examples")

	for _, tc := range []struct {
		file     string
		eolTBA   bool
		eossTBA  bool
	}{
		{"oasis_openeox_tc-core-1_0-2025-minimal-tba.json", true, true},
		{"oasis_openeox_tc-core-1_0-2025-minimal-eol-tba.json", true, false},
		{"oasis_openeox_tc-core-1_0-2025-minimal-eoss-tba.json", false, true},
		{"oasis_openeox_tc-core-1_0-2025-minimal.json", false, false},
	} {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(examplesDir, tc.file))
			require.NoError(t, err)

			core, err := parser.ParseCore(data)
			require.NoError(t, err)

			require.Equal(t, tc.eolTBA, IsTBA(core.GetEndOfLife()), "end_of_life TBA mismatch")
			require.Equal(t, tc.eossTBA, IsTBA(core.GetEndOfSecuritySupport()), "end_of_security_support TBA mismatch")
			require.False(t, IsTBA(core.GetLastUpdated()), "last_updated must never be TBA")
		})
	}
}

// TestConformanceMarshalTBA verifies that TBA sentinel timestamps are
// written back as "tba" in the JSON output.
func TestConformanceMarshalTBA(t *testing.T) {
	core := &Core{
		Schema:               CoreSchema,
		EndOfLife:            TBATimestamp(),
		EndOfSecuritySupport: TBATimestamp(),
		LastUpdated:          timestamppb.New(time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)),
	}

	data, err := MarshalCore(core)
	require.NoError(t, err)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	require.Equal(t, "tba", raw["end_of_life"])
	require.Equal(t, "tba", raw["end_of_security_support"])
}

// TestConformanceGenerateValidateCore creates Core documents using our proto
// types, marshals them to JSON, and validates the output against the upstream
// OpenEoX core JSON schema.
func TestConformanceGenerateValidateCore(t *testing.T) {
	repoPath := upstreamRepoPath(t)
	schema := loadCoreSchemaValidator(t, repoPath)

	for _, tc := range []struct {
		name string
		core *Core
	}{
		{
			name: "all-timestamps",
			core: &Core{
				Schema:               CoreSchema,
				EndOfLife:            timestamppb.New(time.Date(2028, 12, 31, 23, 59, 59, 0, time.UTC)),
				EndOfSecuritySupport: timestamppb.New(time.Date(2028, 6, 30, 23, 59, 59, 0, time.UTC)),
				EndOfSales:           timestamppb.New(time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)),
				GeneralAvailability:  timestamppb.New(time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)),
				LastUpdated:          timestamppb.New(time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "minimal-required-fields",
			core: &Core{
				Schema:               CoreSchema,
				EndOfLife:            timestamppb.New(time.Date(2026, 12, 31, 12, 0, 0, 0, time.UTC)),
				EndOfSecuritySupport: timestamppb.New(time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)),
				LastUpdated:          timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "tba-end-of-life",
			core: &Core{
				Schema:               CoreSchema,
				EndOfLife:            TBATimestamp(),
				EndOfSecuritySupport: timestamppb.New(time.Date(2028, 6, 30, 23, 59, 59, 0, time.UTC)),
				LastUpdated:          timestamppb.New(time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "all-tba",
			core: &Core{
				Schema:               CoreSchema,
				EndOfLife:            TBATimestamp(),
				EndOfSecuritySupport: TBATimestamp(),
				LastUpdated:          timestamppb.New(time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := MarshalCore(tc.core)
			require.NoError(t, err)

			var doc interface{}
			require.NoError(t, json.Unmarshal(data, &doc))
			require.NoError(t, schema.Validate(doc), "generated JSON must validate against upstream core schema:\n%s", string(data))
		})
	}
}

// TestConformanceRoundTripShell verifies that Shell documents with various
// product identification helpers survive a marshal → parse round trip.
func TestConformanceRoundTripShell(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	shell := &Shell{
		Schema: Schema,
		Statements: []*Statement{
			{
				Core: &Core{
					Schema:               CoreSchema,
					EndOfLife:            timestamppb.New(time.Date(2029, 12, 31, 23, 59, 59, 0, time.UTC)),
					EndOfSecuritySupport: timestamppb.New(time.Date(2029, 6, 30, 23, 59, 59, 0, time.UTC)),
					LastUpdated:          timestamppb.New(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)),
				},
				Product: &Product{
					Schema:         ProductSoftwareSchema,
					ProductName:    "Test Product",
					ProductVersion: "1.0.0",
					VendorName:     "Test Vendor",
				},
				ProductIdentificationHelper: &ProductIdentificationHelper{
					Cpe:   "cpe:2.3:a:test_vendor:test_product:1.0.0:*:*:*:*:*:*:*",
					Purls: []string{"pkg:generic/test-vendor/test-product@1.0.0"},
					SbomUrls: []string{
						"https://test.example.com/sbom/test-product-1.0.0.spdx.json",
					},
					Hashes: []*CryptographicHashes{
						{
							Filename: "test-product-1.0.0.tar.gz",
							FileHashes: []*FileHash{
								{Algorithm: "sha256", Value: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
							},
						},
					},
					XGenericUris: []*GenericURI{
						{
							Namespace: "https://test.example.com/ns/product-id",
							Uri:       "https://test.example.com/products/test-product/1.0.0",
						},
					},
				},
			},
		},
	}

	data, err := MarshalShell(shell)
	require.NoError(t, err)

	parsed, err := parser.ParseShell(data)
	require.NoError(t, err)

	require.Len(t, parsed.GetStatements(), 1)
	st := parsed.GetStatements()[0]

	// Product
	require.Equal(t, "Test Product", st.GetProduct().GetProductName())
	require.Equal(t, "1.0.0", st.GetProduct().GetProductVersion())
	require.Equal(t, "Test Vendor", st.GetProduct().GetVendorName())

	// Product identification helper
	pih := st.GetProductIdentificationHelper()
	require.Equal(t, "cpe:2.3:a:test_vendor:test_product:1.0.0:*:*:*:*:*:*:*", pih.GetCpe())
	require.Equal(t, []string{"pkg:generic/test-vendor/test-product@1.0.0"}, pih.GetPurls())
	require.Equal(t, []string{"https://test.example.com/sbom/test-product-1.0.0.spdx.json"}, pih.GetSbomUrls())
	require.Len(t, pih.GetHashes(), 1)
	require.Equal(t, "test-product-1.0.0.tar.gz", pih.GetHashes()[0].GetFilename())
	require.Len(t, pih.GetHashes()[0].GetFileHashes(), 1)
	require.Equal(t, "sha256", pih.GetHashes()[0].GetFileHashes()[0].GetAlgorithm())
	require.Len(t, pih.GetXGenericUris(), 1)
	require.Equal(t, "https://test.example.com/ns/product-id", pih.GetXGenericUris()[0].GetNamespace())

	// Core
	require.Equal(t, time.Date(2029, 12, 31, 23, 59, 59, 0, time.UTC), st.GetCore().GetEndOfLife().AsTime())
	require.Equal(t, time.Date(2029, 6, 30, 23, 59, 59, 0, time.UTC), st.GetCore().GetEndOfSecuritySupport().AsTime())
}

// TestConformanceRoundTripTBA verifies that TBA values survive a
// marshal → parse round trip through the Shell format.
func TestConformanceRoundTripTBA(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	shell := &Shell{
		Schema: Schema,
		Statements: []*Statement{
			{
				Core: &Core{
					Schema:               CoreSchema,
					EndOfLife:            TBATimestamp(),
					EndOfSecuritySupport: TBATimestamp(),
					LastUpdated:          timestamppb.New(time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)),
				},
				Product: &Product{
					Schema:         ProductSoftwareSchema,
					ProductName:    "Future Product",
					ProductVersion: "0.1.0",
					VendorName:     "Test Vendor",
				},
				ProductIdentificationHelper: &ProductIdentificationHelper{
					Purls: []string{"pkg:generic/test-vendor/future-product@0.1.0"},
				},
			},
		},
	}

	data, err := MarshalShell(shell)
	require.NoError(t, err)

	parsed, err := parser.ParseShell(data)
	require.NoError(t, err)

	core := parsed.GetStatements()[0].GetCore()
	require.True(t, IsTBA(core.GetEndOfLife()))
	require.True(t, IsTBA(core.GetEndOfSecuritySupport()))
	require.False(t, IsTBA(core.GetLastUpdated()))
}

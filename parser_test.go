// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestParse(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	data, err := os.ReadFile("testdata/latest.eox.json")
	require.NoError(t, err)
	s, err := parser.ParseShell(data)
	require.NoError(t, err)

	require.Len(t, s.GetStatements(), 2)
	require.Equal(t, Schema, s.GetSchema())

	statement := s.GetStatements()[0]

	// Product
	product := statement.GetProduct()
	require.NotNil(t, product)
	require.Equal(t, ProductSoftwareSchema, product.GetSchema())
	require.Equal(t, "Example Technologies", product.GetVendorName())
	require.Equal(t, "Enterprise Server", product.GetProductName())
	require.Equal(t, "5.2", product.GetProductVersion())

	// Product Identification Helper
	pih := statement.GetProductIdentificationHelper()
	require.NotNil(t, pih)
	require.Equal(t, "cpe:2.3:a:example_technologies:enterprise_server:5.2:*:*:*:*:*:*:*", pih.GetCpe())
	require.Len(t, pih.GetPurls(), 1)
	require.Equal(t, "pkg:generic/example-technologies/enterprise-server@5.2", pih.GetPurls()[0])

	// Core
	core := statement.GetCore()
	require.Equal(t, CoreSchema, core.GetSchema())
	require.Equal(t, time.Date(2025, time.April, 30, 10, 0, 0, 0, time.UTC), core.GetLastUpdated().AsTime())
	require.Equal(t, time.Date(2027, time.December, 31, 23, 59, 59, 0, time.UTC), core.GetEndOfLife().AsTime())
	require.Equal(t, time.Date(2026, time.December, 31, 23, 59, 59, 0, time.UTC), core.GetEndOfSales().AsTime())
	require.Equal(t, time.Date(2027, time.June, 30, 23, 59, 59, 0, time.UTC), core.GetEndOfSecuritySupport().AsTime())
	require.Equal(t, time.Date(2020, time.March, 15, 0, 0, 0, 0, time.UTC), core.GetGeneralAvailability().AsTime())
}

func TestParseHashes(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	data, err := os.ReadFile("testdata/hashes.eox.json")
	require.NoError(t, err)
	s, err := parser.ParseShell(data)
	require.NoError(t, err)

	require.Len(t, s.GetStatements(), 1)
	st := s.GetStatements()[0]

	// Product
	require.Equal(t, "SecureLib", st.GetProduct().GetProductName())
	require.Equal(t, "2.1.0", st.GetProduct().GetProductVersion())
	require.Equal(t, "Acme Corp", st.GetProduct().GetVendorName())

	// Hashes
	pih := st.GetProductIdentificationHelper()
	require.Len(t, pih.GetHashes(), 2)

	h0 := pih.GetHashes()[0]
	require.Equal(t, "securelib-2.1.0.tar.gz", h0.GetFilename())
	require.Len(t, h0.GetFileHashes(), 2)
	require.Equal(t, "sha256", h0.GetFileHashes()[0].GetAlgorithm())
	require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", h0.GetFileHashes()[0].GetValue())
	require.Equal(t, "sha512", h0.GetFileHashes()[1].GetAlgorithm())

	h1 := pih.GetHashes()[1]
	require.Equal(t, "securelib-2.1.0.tar.gz.sig", h1.GetFilename())
	require.Len(t, h1.GetFileHashes(), 1)

	// Core
	require.Equal(t, time.Date(2028, time.June, 30, 23, 59, 59, 0, time.UTC), st.GetCore().GetEndOfLife().AsTime())
}

func TestParseSoftwareIdentifiers(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	data, err := os.ReadFile("testdata/software-identifiers.eox.json")
	require.NoError(t, err)
	s, err := parser.ParseShell(data)
	require.NoError(t, err)

	require.Len(t, s.GetStatements(), 2)

	// First statement: all software identifier types
	st0 := s.GetStatements()[0]
	pih := st0.GetProductIdentificationHelper()
	require.Equal(t, "cpe:2.3:a:acme:widget_framework:4.0.0:*:*:*:*:*:*:*", pih.GetCpe())
	require.Equal(t, []string{
		"pkg:npm/%40acme/widget-framework@4.0.0",
		"pkg:pypi/acme-widget-framework@4.0.0",
	}, pih.GetPurls())
	require.Equal(t, []string{
		"https://acme.example.com/sbom/widget-framework-4.0.0.spdx.json",
	}, pih.GetSbomUrls())
	require.Len(t, pih.GetXGenericUris(), 1)
	require.Equal(t, "https://acme.example.com/ns/product-id", pih.GetXGenericUris()[0].GetNamespace())
	require.Equal(t, "https://acme.example.com/products/widget-framework/4.0.0", pih.GetXGenericUris()[0].GetUri())

	// Core with general_availability
	require.Equal(t, time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC), st0.GetCore().GetGeneralAvailability().AsTime())

	// Second statement: purls + hashes combined
	st1 := s.GetStatements()[1]
	pih1 := st1.GetProductIdentificationHelper()
	require.Len(t, pih1.GetPurls(), 1)
	require.Equal(t, "pkg:npm/%40acme/widget-framework@5.0.0", pih1.GetPurls()[0])
	require.Len(t, pih1.GetHashes(), 1)
	require.Equal(t, "widget-framework-5.0.0.tgz", pih1.GetHashes()[0].GetFilename())
}

func TestParseTBA(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	// A core document with "tba" lifecycle dates
	data := []byte(`{
		"$schema": "https://docs.oasis-open.org/openeox/v1.0/schema/core.json",
		"end_of_life": "tba",
		"end_of_security_support": "tba",
		"last_updated": "2025-07-01T12:00:00Z"
	}`)

	core, err := parser.ParseCore(data)
	require.NoError(t, err)

	require.True(t, IsTBA(core.GetEndOfLife()))
	require.True(t, IsTBA(core.GetEndOfSecuritySupport()))
	require.False(t, IsTBA(core.GetLastUpdated()))
	require.Equal(t, time.Date(2025, time.July, 1, 12, 0, 0, 0, time.UTC), core.GetLastUpdated().AsTime())
}

func TestMarshalTBA(t *testing.T) {
	core := &Core{
		Schema:               CoreSchema,
		EndOfLife:            TBATimestamp(),
		EndOfSecuritySupport: TBATimestamp(),
		LastUpdated:          timestamppb.New(time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)),
	}

	data, err := MarshalCore(core)
	require.NoError(t, err)

	// Round-trip: parse the marshaled output and verify TBA is preserved
	parser, err := NewParser()
	require.NoError(t, err)
	parsed, err := parser.ParseCore(data)
	require.NoError(t, err)

	require.True(t, IsTBA(parsed.GetEndOfLife()))
	require.True(t, IsTBA(parsed.GetEndOfSecuritySupport()))
	require.False(t, IsTBA(parsed.GetLastUpdated()))
}

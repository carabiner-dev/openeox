// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	data, err := os.ReadFile("testdata/latest.eox.json")
	require.NoError(t, err)
	s, err := parser.ParseShell(data)
	require.NoError(t, err)

	require.Len(t, s.GetStatements(), 2)
	statement := s.GetStatements()[0]
	require.Equal(t, "https://docs.oasis-open.org/openeox/tbd/schema/shell.json", s.GetSchema())

	// Product:
	product := statement.GetProduct()
	require.NotNil(t, product)
	sw := product.GetSoftware()
	require.NotNil(t, sw)
	require.Equal(t, "Example Technologies", sw.GetVendorName())
	require.Equal(t, "Enterprise Server", sw.GetProductName())
	require.Equal(t, "5.2", sw.GetProductVersion())

	// Product Identification Helper:
	pih := statement.GetProductIdentificationHelper()
	require.NotNil(t, pih)
	require.Equal(t, "cpe:2.3:a:example_technologies:enterprise_server:5.2:*:*:*:*:*:*:*", pih.GetCpe())
	require.Len(t, pih.GetPurls(), 1)
	require.Equal(t, "pkg:generic/example-technologies/enterprise-server@5.2", pih.GetPurls()[0])

	// Core:
	core := statement.GetCore()
	require.Equal(t, "https://docs.oasis-open.org/openeox/v1.0/schema/core.json", core.GetSchema())
	require.Equal(t, time.Date(2025, time.April, 30, 10, 0, 0, 0, time.UTC), core.GetLastUpdated().AsTime())
	require.Equal(t, time.Date(2027, time.December, 31, 23, 59, 59, 0, time.UTC), core.GetEndOfLife().GetTimestamp().AsTime())
	require.Equal(t, time.Date(2026, time.December, 31, 23, 59, 59, 0, time.UTC), core.GetEndOfSales().GetTimestamp().AsTime())
	require.Equal(t, time.Date(2027, time.June, 30, 23, 59, 59, 0, time.UTC), core.GetEndOfSecuritySupport().GetTimestamp().AsTime())
	require.Equal(t, time.Date(2020, time.March, 15, 0, 0, 0, 0, time.UTC), core.GetGeneralAvailability().GetTimestamp().AsTime())
}

// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestConformance(t *testing.T) {
	unmarshaller := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	s := &Shell{}
	data, err := os.ReadFile("testdata/sample.eox.json")
	require.NoError(t, err)
	require.NoError(t, unmarshaller.Unmarshal(data, s))

	// Shell
	require.Equal(t, Schema, s.GetSchema())
	require.Len(t, s.GetStatements(), 2)
	statement := s.GetStatements()[0]

	// Product
	product := statement.GetProduct()
	require.Equal(t, ProductSoftwareSchema, product.GetSchema())
	require.Equal(t, "Enterprise Server", product.GetProductName())
	require.Equal(t, "5.2", product.GetProductVersion())
	require.Equal(t, "Example Technologies", product.GetVendorName())

	// Product Identification Helper
	pih := statement.GetProductIdentificationHelper()
	require.NotNil(t, pih)
	require.Equal(t, "cpe:2.3:a:example_technologies:enterprise_server:5.2:*:*:*:*:*:*:*", pih.GetCpe())
	require.Len(t, pih.GetPurls(), 1)

	// Core
	core := statement.GetCore()
	require.Equal(t, time.Date(2027, time.December, 31, 23, 59, 59, 0, time.UTC), core.GetEndOfLife().AsTime())
	require.Equal(t, time.Date(2027, time.June, 30, 23, 59, 59, 0, time.UTC), core.GetEndOfSecuritySupport().AsTime())
	require.Equal(t, time.Date(2026, time.December, 31, 23, 59, 59, 0, time.UTC), core.GetEndOfSales().AsTime())
	require.Equal(t, time.Date(2025, time.April, 30, 10, 0, 0, 0, time.UTC), core.GetLastUpdated().AsTime())
}

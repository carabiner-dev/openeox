// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package tbd

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

	// Check the data
	require.Equal(t, "https://docs.oasis-open.org/openeox/tbd/schema/shell.json", s.GetSchema())
	require.Len(t, s.GetStatements(), 2)
	statement := s.GetStatements()[0]

	// Statement:
	require.Equal(t, "Enterprise Server", statement.GetProductName())
	require.Equal(t, "5.2", statement.GetProductVersion())
	require.Equal(t, "Example Technologies", statement.GetVendorName())

	// Core
	require.Equal(t, time.Date(2027, time.December, 31, 23, 59, 59, 0, time.UTC), statement.GetCore().GetEndOfLife().AsTime())
	require.Equal(t, time.Date(2027, time.June, 30, 23, 59, 59, 0, time.UTC), statement.GetCore().GetEndOfSecuritySupport().AsTime())
	require.Equal(t, time.Date(2026, time.December, 31, 23, 59, 59, 0, time.UTC), statement.GetCore().GetEndOfSales().AsTime())
	require.Equal(t, time.Date(2025, time.April, 30, 10, 0, 0, 0, time.UTC), statement.GetCore().GetLastUpdated().AsTime())
}

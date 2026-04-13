// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// TBA is the literal string used in OpenEoX JSON documents to indicate
// that a lifecycle date has not yet been announced.
const TBA = "tba"

// tbaSentinel is a far-future RFC 3339 timestamp used internally to
// represent "tba" in google.protobuf.Timestamp fields. It is chosen to
// be greater than any realistic lifecycle date, matching the spec's
// rule that "tba" compares greater than any concrete date-time.
const tbaSentinel = "9999-12-31T23:59:59Z"

// TBATime is the Go time.Time representation of the TBA sentinel.
// Use IsTBA to check whether a parsed timestamp represents "tba".
var TBATime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// TBATimestamp returns a new protobuf Timestamp representing "tba".
// When marshaled to OpenEoX JSON, this value is written as "tba".
func TBATimestamp() *timestamppb.Timestamp {
	return timestamppb.New(TBATime)
}

// IsTBA reports whether ts represents a "to be announced" lifecycle date.
// It returns true when ts holds the sentinel value used to encode "tba"
// in the protobuf representation, and false for nil timestamps.
func IsTBA(ts *timestamppb.Timestamp) bool {
	if ts == nil {
		return false
	}
	return ts.AsTime().Equal(TBATime)
}

// eoxTimestampFields are the Core fields that use the eox_timestamp_t
// type and may contain "tba" instead of an RFC 3339 timestamp.
var eoxTimestampFields = []string{
	"end_of_life", "end_of_security_support", "end_of_sales", "general_availability",
}

// preprocessTBACore replaces "tba" string values in eox_timestamp_t
// fields of a raw Core JSON object with the sentinel timestamp so that
// protojson can unmarshal them into google.protobuf.Timestamp.
func preprocessTBACore(data []byte) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("preprocessing tba: %w", err)
	}
	replaceTBAFields(raw)
	return json.Marshal(raw)
}

// preprocessTBAShell replaces "tba" values in all core objects nested
// inside a Shell document's statements.
func preprocessTBAShell(data []byte) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("preprocessing tba: %w", err)
	}
	stmts, ok := raw["statements"].([]interface{})
	if !ok {
		return data, nil
	}
	for _, stmt := range stmts {
		s, ok := stmt.(map[string]interface{})
		if !ok {
			continue
		}
		core, ok := s["core"].(map[string]interface{})
		if !ok {
			continue
		}
		replaceTBAFields(core)
	}
	return json.Marshal(raw)
}

func replaceTBAFields(obj map[string]interface{}) {
	for _, field := range eoxTimestampFields {
		if v, ok := obj[field]; ok && v == TBA {
			obj[field] = tbaSentinel
		}
	}
}

// postprocessTBA replaces the sentinel timestamp string with "tba" in
// marshaled JSON output, restoring the upstream OpenEoX representation.
func postprocessTBA(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte(`"`+tbaSentinel+`"`), []byte(`"tba"`))
}

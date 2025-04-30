// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package openeox

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/carabiner-dev/openeox/types/tbd"
)

// NewParser creates a new OpenEOX parser
func NewParser() (*Parser, error) {
	return &Parser{}, nil
}

// Parser implements the OpenEOX parser
type Parser struct{}

//nolint:unused
type schemaFinder struct {
	Schema string `json:"$schema"`
}

//nolint:unused
var typeDict = map[string]reflect.Type{
	"https://docs.oasis-open.org/openeox/tbd/schema/shell.json": reflect.TypeOf(&tbd.Shell{}),
	"https://docs.oasis-open.org/openeox/tbd/schema/core.json":  reflect.TypeOf(&tbd.Core{}),
}

//nolint:unused
func detectSchema(data []byte) (reflect.Type, error) {
	sf := &schemaFinder{}
	if err := json.Unmarshal(data, sf); err != nil {
		return nil, fmt.Errorf("unable to detect type")
	}
	if t, ok := typeDict[sf.Schema]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("unknown schema type")
}

func (p *Parser) ParseShell(data []byte) (*Shell, error) {
	s := &Shell{}
	unmarshaller := protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}

	if err := unmarshaller.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}
	return s, nil
}

func (p *Parser) ParseCore(r io.Reader) (*Core, error) {
	c := &Core{}
	unmarshaller := protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	if err := unmarshaller.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}
	return c, err
}

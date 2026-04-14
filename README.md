# OpenEoX Parser, Go Types and Protobuf definitions

This repository contains a parser and types to read and work with
OASIS [OpenEoX](https://openeox.org/) data. This repo also hosts the
[protocol buffer definitions](proto/) of the OpenEoX elements from
which the types are generated.

## What is OpenEoX

From the official website:

> OpenEoX is an initiative aimed at standardizing the way
> End-of-Life (EOL) and End-of-Support (EOS) information
> is exchanged within the software and hardware industries. 

OpenEoX is a simple format to communicate the end of support of software
and hardware. It is designed to complement SBOM, VEX and security advisories.

## Install

To install the module, simply pull it with go get:

```bash
go get github.com/carabiner-dev/openeox
```

## Usage

### Parsing

The parser reads both OpenEoX Shell and Core documents:

```go
parser, _ := openeox.NewParser()

// Parse a shell document (contains product info + lifecycle data)
shell, err := parser.ParseShell(data)

// Parse a standalone core document (lifecycle data only)
core, err := parser.ParseCore(data)
```

### Handling "tba" (To Be Announced)

The OpenEoX spec allows lifecycle date fields (`end_of_life`,
`end_of_security_support`, `end_of_sales`, `general_availability`) to be
the literal string `"tba"` instead of an RFC 3339 timestamp, meaning the
date has not yet been announced.

Since the Go types use `google.protobuf.Timestamp` for these fields (giving
you real `time.Time` values), the parser transparently maps `"tba"` to a
far-future sentinel timestamp (`9999-12-31T23:59:59Z`). Use `IsTBA` to
check whether a parsed timestamp represents "tba":

```go
core, _ := parser.ParseCore(data)

if openeox.IsTBA(core.GetEndOfLife()) {
    fmt.Println("End of life has not been announced yet")
} else {
    fmt.Println("End of life:", core.GetEndOfLife().AsTime())
}
```

To create a TBA timestamp when building documents:

```go
core := &openeox.Core{
    Schema:               openeox.CoreSchema,
    EndOfLife:            openeox.TBATimestamp(),
    EndOfSecuritySupport: openeox.TBATimestamp(),
    LastUpdated:          timestamppb.Now(),
}
```

The marshal functions (`MarshalCore`, `MarshalShell`) automatically convert
sentinel timestamps back to `"tba"` in the JSON output, producing
spec-compliant documents.

### Building a Document

Here is a complete example that builds an OpenEoX shell document with
product identification using file hashes:

```go
shell := &openeox.Shell{
    Schema: openeox.Schema,
    Statements: []*openeox.Statement{
        {
            Core: &openeox.Core{
                Schema:               openeox.CoreSchema,
                EndOfLife:            timestamppb.New(time.Date(2028, 6, 30, 23, 59, 59, 0, time.UTC)),
                EndOfSecuritySupport: timestamppb.New(time.Date(2028, 3, 31, 23, 59, 59, 0, time.UTC)),
                EndOfSales:           timestamppb.New(time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)),
                LastUpdated:          timestamppb.Now(),
            },
            Product: &openeox.Product{
                Schema:         openeox.ProductSoftwareSchema,
                ProductName:    "SecureLib",
                ProductVersion: "2.1.0",
                VendorName:     "Acme Corp",
            },
            ProductIdentificationHelper: &openeox.ProductIdentificationHelper{
                Purls: []string{"pkg:generic/acme/securelib@2.1.0"},
                Hashes: []*openeox.CryptographicHashes{
                    {
                        Filename: "securelib-2.1.0.tar.gz",
                        FileHashes: []*openeox.FileHash{
                            {
                                Algorithm: "sha256",
                                Value:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
                            },
                        },
                    },
                },
            },
        },
    },
}

data, err := openeox.MarshalShell(shell)
```

This produces:

```json
{
  "$schema": "https://docs.oasis-open.org/openeox/tbd/schema/shell.json",
  "statements": [
    {
      "core": {
        "$schema": "https://docs.oasis-open.org/openeox/v1.0/schema/core.json",
        "end_of_life": "2028-06-30T23:59:59Z",
        "end_of_security_support": "2028-03-31T23:59:59Z",
        "end_of_sales": "2027-12-31T23:59:59Z",
        "last_updated": "2025-04-13T12:00:00Z"
      },
      "product": {
        "$schema": "https://docs.oasis-open.org/openeox/tbd/schema/product_software.json",
        "product_name": "SecureLib",
        "product_version": "2.1.0",
        "vendor_name": "Acme Corp"
      },
      "product_identification_helper": {
        "purls": ["pkg:generic/acme/securelib@2.1.0"],
        "hashes": [
          {
            "filename": "securelib-2.1.0.tar.gz",
            "file_hashes": [
              {
                "algorithm": "sha256",
                "value": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
              }
            ]
          }
        ]
      }
    }
  ]
}
```

### Marshaling

Use the provided marshal functions to produce JSON that conforms to the
upstream OpenEoX schema:

```go
data, err := openeox.MarshalCore(core)
data, err := openeox.MarshalShell(shell)
```

## Spec Conformance

The module tracks the OASIS OpenEoX specification:

- **openeox-core v1.0**: Fully supported. The `Core` message maps directly
  to the [upstream JSON schema](https://docs.oasis-open.org/openeox/v1.0/schema/core.json).
  Conformance is validated against the upstream schema and test data in CI.
- **openeox-shell**: Tracks the current draft (CSD01). The `Shell`,
  `Statement`, `Product`, and `ProductIdentificationHelper` messages
  follow the draft shell schema structure.

## Copyright and License

This module is copyright by Carabiner Systems, Inc. It is released under the
Apache 2.0 license, feel free to use it for anything, contribute patches or file
issues with bugs or feature requests.

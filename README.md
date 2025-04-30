# OpenEoX Parser, Go Types and Protobuf definitions

This repository contains a parser and types to read and work with
OASIS [OpenEoX](https://openeox.org/) data. This repo also hosts the
[protocol buffer definitions](proto/) of the OpenEoX elements from
which the types are generated.

## What is OpenEoX:

From the official website:

> OpenEoX is an initiative aimed at standardizing the way
> End-of-Life (EOL) and End-of-Support (EOS) information
> is exchanged within the software and hardware industries. 

OpenEoX is a simple format to communicate the end of support of software
and hardware. It is designed to completement SBOM, VEX and security advisories.

## Install

To install the module, simply pull it with go get:

```bash
go get github.com/carabiner-dev/openeox
```

## Module Status

The module offers a simple parser, we have one version of the schema which is
following the development of the first version and will be solidified once the
fitst official release is out.

## Copyright and License

This modules is copyright by Carabiner Systems, Inc. It is released under the
Apache 2.0 license, feel free to use it for anything, contribute patches or file
issues with bugs or feature requests.

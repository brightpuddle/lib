# Overview

This library is a utility toolkit for specific projects. It wraps and tailors
the behavior of several 3rd-party Go libraries. The modifications are for
specific use cases and not really intended for general consumption. That said,
if one of these use cases overlaps with your project requirements feel free to
fork and reuse.

# Tools

## Logger

The logger wraps [zerolog](https://github.com/rs/zerolog) with a package-level
singleton and common logging standards.

It also includes [fiber](http://fiber.wiki/) logger middleware.

## JSON

The json library wraps [GJSON](https://github.com/tidwall/gjson),
[SJSON](https://github.com/tidwall/sjson), and
[Segmentio's encoding/json](https://github.com/segmentio/encoding/json)
libraries. It removes explicit errors in favor of embeded logging and
initializes nil values to avoid passing null throuh to JSON data.

**Note**: allocating nil values incurs a not-insignificant performance penalty,
and should be tested thoroughly for performance-sensitive use cases.

## Archive

The archive module abstracts `.zip` and `.tar.gz` extraction.

## ACI

Cisco ACI utilities.

### MIT

The MIT module parses ACI JSON data, e.g. from a JSON backup file,
`moquery -o json`, `icurl` or any other ACI MO JSON data source.

The data is parsed into a [BuntDB](https://github.com/tidwall/buntdb) in-memory
database with `class:dn` as the key and the managed object fiels as values. This
is fronted with `Get`, `Find`, and `FindOne` functions for querying the DB.

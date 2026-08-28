// Package oid provides functionality for working with Object Identifiers (OID).
//
// The package includes:
//   - OID parsing and validation
//   - ASN.1 BER/DER encoding and decoding
//   - JSON serialization
//   - Registry for named OIDs
//   - Database/sql integration
//   - Global API for convenient access
//
// Example usage:
//
//	oid := oid.MustParseOID("1.3.6.1.4.1")
//	oid.MustRegister("enterprise", oid)
//	if o, exists := oid.LookupByName("enterprise"); exists {
//	    fmt.Println(o)
//	}
package oid

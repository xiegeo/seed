package seed

// NameCheck verifies that code names in this domain follow certain sensible rules.
//
// - The only allowed characters are "A" to "Z", "a" to "z", "0" to "9", and "_".
// - "_" and "0" to "9" can not be the first character.
// - "_" can not be the last character.
// - "__" is not allowed.
//
// Duplication is checked on simplified versions of code names by removing case and "_".
// "AB", "aB", "A_b" all simplify to "ab". A field can not have a name that is the prefix
// of another under the same parent (except "v2", "v3"... postfixes). All fields of the
// same name must have the same properties (except label and description).
//
// Until NameCheck is fully implemented, domain consumers can just pretend it's never violated.
func (d Domain) NameCheck() error {
	// future
	return nil
}

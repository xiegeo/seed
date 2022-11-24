package sqldb

// AssertParameterAlignment returns true if parameter alignment, such as rows values lining up with cols
// should be asserted.
func AssertParameterAlignment() bool {
	return true // future: settable
}

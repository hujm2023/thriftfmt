package parser

// tabSize specifies the default indentation size.
var tabSize = 4

// SetTabSize sets the global tab size for formatting.
func SetTabSize(size int) {
	tabSize = size
}

// patchRequired indicates whether to add 'required' to fields lacking 'required' or 'optional'.
var patchRequired bool

// SetPatchRequired sets whether to patch required/optional fields.
func SetPatchRequired(required bool) {
	patchRequired = required
}

var defaultDelimiter = ","

var (
	// structFieldDelimiter specifies the default delimiter for struct fields.
	structFieldDelimiter = defaultDelimiter
	// serviceFieldDelimiter specifies the default delimiter for service function fields.
	serviceFieldDelimiter = defaultDelimiter
	// enumFieldDelimiter specifies the default delimiter for enum fields.
	enumFieldDelimiter = defaultDelimiter
)

// SetStructDelimiter sets the delimiter for struct fields.
func SetStructDelimiter(delimiter string) {
	structFieldDelimiter = delimiter
}

// SetServiceDelimiter sets the delimiter for service function fields.
func SetServiceDelimiter(delimiter string) {
	serviceFieldDelimiter = delimiter
}

// SetEnumDelimiter sets the delimiter for enum fields.
func SetEnumDelimiter(delimiter string) {
	enumFieldDelimiter = delimiter
}

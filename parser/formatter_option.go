package parser

// tabSize 默认缩进
var tabSize = 4

func SetTabSize(size int) {
	tabSize = size
}

// patchRequired 是否需要给缺少 required/optional 的字段补充 required
var patchRequired bool

func SetPatchRequired(required bool) {
	patchRequired = required
}

var defaultDelimiter = ","

var (
	// structFieldDelimiter Struct Field 默认分隔符
	structFieldDelimiter = defaultDelimiter
	// defaultFieldDelimiter Service Function 默认分隔符
	serviceFieldDelimiter = defaultDelimiter
	// enumFieldDelimiter Enum Field 默认分隔符
	enumFieldDelimiter = defaultDelimiter
)

func SetStructDelimiter(delimiter string) {
	structFieldDelimiter = delimiter
}

func SetServiceDelimiter(delimiter string) {
	serviceFieldDelimiter = delimiter
}

func SetEnumDelimiter(delimiter string) {
	enumFieldDelimiter = delimiter
}

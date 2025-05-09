package parser

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/valyala/bytebufferpool"
)

const (
	fieldDelimiter  = ","
	commentSplitter = "\u200B"
)

var (
	atMost1Space = regexp.MustCompile(` +`)
	atMost2Slash = regexp.MustCompile(`(\n{3,})`)
)

// Comment represents Thrift inline/unix Comment.
type Comment string

func (c Comment) FormatThrift() string {
	s := string(c)

	var head, content string
	switch {
	case isLineComment(s):
		head = "//"
		content = strings.TrimLeft(s[2:], " ")
	case isUnixComment(s):
		head = "#"
		content = strings.TrimLeft(s[1:], " ")
	default:
		return s
	}

	if len(content) == 0 {
		return head
	}

	content = atMost1Space.ReplaceAllString(content, " ")
	content = atMost2Slash.ReplaceAllString(content, "\n\n")
	content = strings.TrimSuffix(content, " ")

	return head + " " + content
}

// LineV2 represents a single line in Thrift format
// For example `1: string name, // this is name` where:
//
//	define="1: string name" - the field definition
//	comment="// this is name" - the trailing comment
type LineV2 struct {
	Define      string
	TailComment string
	HeadComment string
}

type LinesV2 struct {
	lines            []LineV2
	tabSize          int    // number of leading spaces
	sep              string // line separator '\n' or '' (no separator)
	lastHasSep       bool   // whether last element needs separator
	delimiter        string // delimiter ',' or ';'
	lastHasDelimiter bool   // whether last element needs delimiter

	addEmptyLineEachLine bool // whether to add empty lines between elements
}

func NewLinesV2(tabSize int, sep string, delimiter string, lines ...LineV2) *LinesV2 {
	return &LinesV2{
		lines:     lines,
		tabSize:   tabSize,
		delimiter: delimiter,
		sep:       sep,
	}
}

func (l *LinesV2) maxDefine() int {
	var length int
	for _, line := range l.lines {
		n := utf8.RuneCountInString(line.Define)
		if n > length {
			length = n
		}
	}
	return length + 1
}

func (l *LinesV2) isLast(idx int) bool {
	return idx == len(l.lines)-1
}

func (l *LinesV2) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	mf := l.maxDefine()

	for idx, line := range l.lines {
		// Header comment
		if line.HeadComment != "" {
			if isLongComment(line.HeadComment) {
				b.WriteString(l.sep)
				b.WriteString(strings.Repeat(" ", l.tabSize))
				b.WriteString(line.HeadComment)
			} else {
				// Possible scenarios: standalone lineComment; mixed longComment and lineComment
				temp := strings.Split(line.HeadComment, commentSplitter)
				// For multi-line comments, force adding an extra line at the beginning
				needAddHeadLine := len(temp) > 1 && idx != 0
				if needAddHeadLine {
					b.WriteString("\n")
				}
				// t always ends with \n
				for _, t := range temp {
					// For longComment, force adding an extra line at the beginning
					if isLongComment(t) && !needAddHeadLine && idx != 0 {
						b.WriteString("\n")
					}
					if t == "\n" {
						b.WriteString("\n")
						continue
					}
					// Indentation
					if l.tabSize > 0 {
						b.WriteString(strings.Repeat(" ", l.tabSize))
					}
					// Normalize comments
					b.WriteString(Comment(t).FormatThrift())
				}
			}
		}

		// Write define
		if l.tabSize > 0 {
			b.WriteString(strings.Repeat(" ", l.tabSize))
		}
		b.WriteString(line.Define)

		// Delimiter
		if l.delimiter != "" {
			if (l.isLast(idx) && l.lastHasDelimiter) || !l.isLast(idx) {
				b.WriteString(l.delimiter)
			}
		}

		// Write tailComment
		if line.TailComment != "" {
			b.WriteString(strings.Repeat(" ", mf-len(line.Define)))
			b.WriteString(line.TailComment)
		}

		// Line break
		if l.isLast(idx) {
			if l.lastHasSep {
				b.WriteString(l.sep)
			}
		} else {
			b.WriteString(l.sep)
		}

		// Whether to add empty lines between multiple lines
		if l.addEmptyLineEachLine {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func CommentForBigOne(s string) string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	// Possible scenarios: standalone lineComment; mixed longComment and lineComment
	temp := strings.Split(s, commentSplitter)
	// t always ends with \n
	for _, t := range temp {
		// longComment，强制换行
		// if isLongComment(t) {
		// 	b.WriteString("\n")
		// }
		b.WriteString(t)
	}
	return b.String()
}

// -------

func field2Line(v *Field) LineV2 {
	line := LineV2{
		TailComment: Comment(v.GetReservedEndLineComments()).FormatThrift(),
		HeadComment: Comment(v.GetReservedComments()).FormatThrift(),
	}
	var s strings.Builder
	s.WriteString(fmt.Sprintf("%d: ", v.GetID()))
	switch v.GetRequiredness() {
	case FieldType_Default:
		if patchRequired {
			s.WriteString("required ")
		}
	case FieldType_Required:
		s.WriteString("required ")
	case FieldType_Optional:
		s.WriteString("optional ")
	}
	s.WriteString(v.GetType().String())
	s.WriteString(" ")
	s.WriteString(v.GetName())
	if d := v.GetDefault(); d != nil {
		s.WriteString(" = ")
		s.WriteString(ParseConstValue(d))
	}
	WriteAnnotations(v.GetAnnotations(), &s)

	line.Define = s.String()

	return line
}

func FormatFunctionArgs(fs []*Field, sep string) string {
	if len(fs) == 0 {
		return ""
	}

	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	b.WriteString("(")

	ls := make([]LineV2, 0, len(fs))
	for _, f := range fs {
		ls = append(ls, field2Line(f))
	}
	lns := NewLinesV2(0, "", sep, ls...)
	b.WriteString(lns.FormatThrift())

	b.WriteString(")")
	return b.String()
}

func function2Line(v *Function) LineV2 {
	line := LineV2{
		TailComment: Comment(v.GetReservedEndLineComments()).FormatThrift(),
		HeadComment: Comment(v.GetReservedComments()).FormatThrift(),
	}
	s := bytebufferpool.Get()
	defer bytebufferpool.Put(s)
	if v.GetOneway() {
		s.WriteString("oneway ")
	}
	if v.GetVoid() {
		s.WriteString("void ")
	}
	// resp
	if v.GetFunctionType() != nil {
		s.WriteString(v.GetFunctionType().String())
		s.WriteString(" ")
	}
	// func name
	s.WriteString(v.GetName())
	s.WriteString(" ")
	// arguments
	if len(v.GetArguments()) > 0 {
		s.WriteString(FormatFunctionArgs(v.GetArguments(), " "))
	}

	// throws
	if len(v.GetThrows()) > 0 {
		s.WriteString(" throws ")
		s.WriteString(FormatFunctionArgs(v.GetThrows(), " "))
	}
	// TODO: 注解应该在throws之后？
	WriteAnnotations(v.GetAnnotations(), s)

	line.Define = s.String()
	return line
}

func FormatFunctions(fs []*Function) string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	ls := make([]LineV2, 0, len(fs))
	for _, f := range fs {
		ls = append(ls, function2Line(f))
	}
	lns := NewLinesV2(tabSize, "\n", serviceFieldDelimiter, ls...)
	lns.lastHasDelimiter = true
	lns.lastHasSep = true

	b.WriteString(lns.FormatThrift())
	return b.String()
}

func WriteAnnotations(ans Annotations, s io.StringWriter) {
	if len(ans) > 0 {
		_, _ = s.WriteString(" (")
		for j := range ans {
			an := ans[j]
			for i := 0; i < len(an.GetValues()); i++ {
				_, _ = s.WriteString(an.GetKey())
				_, _ = s.WriteString(" = ")
				_, _ = s.WriteString(fmt.Sprintf("%q", an.GetValues()[i]))
				if j != len(ans)-1 || i != len(an.GetValues())-1 {
					_, _ = s.WriteString(", ")
				}
			}
		}
		_, _ = s.WriteString(")")
	}
}

func ParseConstValue(t *ConstValue) string {
	if t == nil {
		return "<nil>"
	}
	var val string
	switch t.Type {
	case ConstType_ConstDouble:
		val = fmt.Sprintf("%f", *t.TypedValue.Double)
	case ConstType_ConstInt:
		val = fmt.Sprintf("%d", *t.TypedValue.Int)
	case ConstType_ConstLiteral:
		val = fmt.Sprintf("\"%s\"", *t.TypedValue.Literal)
	case ConstType_ConstIdentifier:
		val = *t.TypedValue.Identifier
	case ConstType_ConstList:
		if len(t.TypedValue.List) == 0 {
			return "{}"
		}
		var ss []string
		for _, item := range t.TypedValue.List {
			// ss = append(ss, indent(ParseConstValue(item), strings.Repeat(" ", tabSize)))
			ss = append(ss, ParseConstValue(item))
		}
		// 如果超过 6 个，按列展开，否则平铺一行
		if len(ss) <= 6 {
			val = fmt.Sprintf("[%s]", strings.Join(ss, ", "))
		} else {
			tempS := make([]string, 0, len(ss)+2)
			tempS = append(tempS, "[")
			for i := 0; i < len(ss); i++ {
				tempS = append(tempS, fmt.Sprintf("%s,", indent(ss[i], strings.Repeat(" ", tabSize))))
			}
			tempS = append(tempS, "]")
			val = strings.Join(tempS, "\n")
		}
	case ConstType_ConstMap:
		if len(t.TypedValue.Map) == 0 {
			return "{}"
		}
		var ss []string
		ss = append(ss, "{")
		for _, kv := range t.TypedValue.Map {
			k := ParseConstValue(kv.GetKey())
			v := ParseConstValue(kv.GetValue())
			pair := k + ": " + v
			ss = append(ss, fmt.Sprintf("%s,", indent(pair, strings.Repeat(" ", tabSize))))
		}
		ss = append(ss, "}")
		val = strings.Join(ss, "\n")
	default:
		return fmt.Sprintf("%+v", *t)
	}
	return val
}

func FormatInline(fs []ThriftFormatter) string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	var lastType FormatType
	var tempLines []LineV2

	for idx := range fs {
		v := fs[idx].Type()
		// fmt.Println(fs[idx])
		switch v {
		case FormatTypeInclude, FormatTypeCppIncludeInclude,
			FormatTypeNamespace, FormatTypeTypedef, FormatTypeConstant:
			if lastType != v && len(tempLines) > 0 {
				lns := NewLinesV2(0, "\n", "", tempLines...)
				lns.lastHasDelimiter = true
				lns.lastHasSep = true
				b.WriteString(lns.FormatThrift())
				b.WriteString("\n")
				lastType = v
				tempLines = tempLines[:0]
			}
			tempLines = append(tempLines, LineV2{Define: fs[idx].FormatThrift()})
			lastType = v
		default:
			if len(tempLines) > 0 {
				lns := NewLinesV2(0, "\n", "", tempLines...)
				lns.lastHasDelimiter = true
				lns.lastHasSep = true
				b.WriteString(lns.FormatThrift())
				b.WriteString("\n")
				lastType = v
				tempLines = tempLines[:0]
			}
			b.WriteString(fs[idx].FormatThrift())
			b.WriteString("\n")
		}
	}
	if len(tempLines) > 0 {
		lns := NewLinesV2(0, "\n", "", tempLines...)
		lns.lastHasDelimiter = true
		lns.lastHasSep = true
		b.WriteString(lns.FormatThrift())
		b.WriteString("\n")
		tempLines = tempLines[:0]
	}

	// at most 1 empty line at the end
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func FormatStructLike(f StructLike) string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	// struct comment
	if f.GetReservedComments() != "" {
		b.WriteString(CommentForBigOne(f.GetReservedComments()))
	}

	// header
	b.WriteString(f.GetCategory() + " " + f.Name + " {")
	b.WriteString("\n")

	// body
	ls := make([]LineV2, 0, len(f.GetFields()))
	for _, ff := range f.GetFields() {
		ls = append(ls, field2Line(ff))
	}
	lns := NewLinesV2(tabSize, "\n", structFieldDelimiter, ls...)
	lns.lastHasSep = true
	lns.lastHasDelimiter = true

	b.WriteString(lns.FormatThrift())

	// foot
	b.WriteString("}")
	b.WriteString("\n")

	return b.String()
}

func isComment(s string) bool {
	return isLineComment(s) || isUnixComment(s) || isLongComment(s)
}

func isLongComment(s string) bool {
	ss := strings.Trim(s, "\n")
	return strings.HasPrefix(ss, "/*") && strings.HasSuffix(ss, "*/")
}

func isLineComment(s string) bool {
	return strings.HasPrefix(s, "//")
}

func isUnixComment(s string) bool {
	return strings.HasPrefix(s, "#")
}

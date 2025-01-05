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

// LineV2 表示“一行”
// 比如 `1: string name, // this is name`，其中
//
//	define="1: string name",
//	comment="// this is name"
type LineV2 struct {
	Define      string
	TailComment string
	HeadComment string
}

type LinesV2 struct {
	lines            []LineV2
	tabSize          int    // 前导空格个数
	sep              string // 换行符 '\n' 或 ''(不需要换行)
	lastHasSep       bool   // 最后一个元素是否需要换行符
	delimiter        string // 结尾符 ',' 或 ';'
	lastHasDelimiter bool   // 最后一个元素是否需要结尾符

	addEmptyLineEachLine bool // 多个line之间是否要用空行
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
		// 头部comment
		if line.HeadComment != "" {
			if isLongComment(line.HeadComment) {
				b.WriteString(l.sep)
				b.WriteString(strings.Repeat(" ", l.tabSize))
				b.WriteString(line.HeadComment)
			} else {
				// 可能的场景：单独的lineComment; longComment和lineComment混合
				temp := strings.Split(line.HeadComment, commentSplitter)
				// 如果是多行comment，开头强制多加一行
				needAddHeadLine := len(temp) > 1 && idx != 0
				if needAddHeadLine {
					b.WriteString("\n")
				}
				// t 都是以\n结尾的
				for _, t := range temp {
					// longComment，开头强制多加一行
					if isLongComment(t) && !needAddHeadLine && idx != 0 {
						b.WriteString("\n")
					}
					if t == "\n" {
						b.WriteString("\n")
						continue
					}
					// 缩进
					if l.tabSize > 0 {
						b.WriteString(strings.Repeat(" ", l.tabSize))
					}
					// comment 归一
					b.WriteString(Comment(t).FormatThrift())
				}
			}
		}

		// 写 define
		if l.tabSize > 0 {
			b.WriteString(strings.Repeat(" ", l.tabSize))
		}
		b.WriteString(line.Define)

		// 结尾符
		if l.delimiter != "" {
			if (l.isLast(idx) && l.lastHasDelimiter) || !l.isLast(idx) {
				b.WriteString(l.delimiter)
			}
		}

		// 写 tailComment
		if line.TailComment != "" {
			b.WriteString(strings.Repeat(" ", mf-len(line.Define)))
			b.WriteString(line.TailComment)
		}

		// 换行
		if l.isLast(idx) {
			if l.lastHasSep {
				b.WriteString(l.sep)
			}
		} else {
			b.WriteString(l.sep)
		}

		// 多个line之间是否加空行
		if l.addEmptyLineEachLine {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func CommentForBigOne(s string) string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	// 可能的场景：单独的lineComment; longComment和lineComment混合
	temp := strings.Split(s, commentSplitter)
	// t 都是以\n结尾的
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
				_, _ = s.WriteString("\"" + an.GetValues()[i] + "\"")
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

package parser

import (
	"fmt"
	"strings"

	"github.com/valyala/bytebufferpool"
)

var (
	_ ThriftFormatter = (*EnumFormatter)(nil)
	_ ThriftFormatter = (*TypedefFormatter)(nil)
	_ ThriftFormatter = (*ConstantFormatter)(nil)
	_ ThriftFormatter = (*StructLikeFormatter)(nil)
	_ ThriftFormatter = (*IncludeFormatter)(nil)
	_ ThriftFormatter = (*NamespaceFormatter)(nil)
	_ ThriftFormatter = (*ServiceFormatter)(nil)
)

type FormatType string

const (
	FormatTypeInclude           FormatType = "Include"
	FormatTypeCppIncludeInclude FormatType = "CppInclude"
	FormatTypeNamespace         FormatType = "Namespace"
	FormatTypeService           FormatType = "Service"
	FormatTypeEnum              FormatType = "Enum"
	FormatTypeConstant          FormatType = "Constant"
	FormatTypeTypedef           FormatType = "Typedef"
	FormatTypeStructLike        FormatType = "StructLike"
)

type ThriftFormatter interface {
	FormatThrift() string
	Type() FormatType
}

// -----

type EnumFormatter struct {
	/*
		Enum  <-
			ENUM Identifier LWING
				(ReservedComments Identifier (EQUAL IntConstant)? Annotations? ListSeparator? ReservedEndLineComments SkipLine)*
			RWING
	*/

	Enum
}

func (f *EnumFormatter) value2Line(v *EnumValue) LineV2 {
	line := LineV2{
		HeadComment: Comment(v.GetReservedComments()).FormatThrift(),
		TailComment: Comment(v.GetReservedEndLineComments()).FormatThrift(),
	}
	s := bytebufferpool.Get()
	defer bytebufferpool.Put(s)

	s.WriteString(v.GetName())
	s.WriteString(" = ")
	s.WriteString(fmt.Sprintf("%d", v.Value))
	WriteAnnotations(v.GetAnnotations(), s)

	line.Define = s.String()
	return line
}

func (f *EnumFormatter) FormatThrift() string {
	var b strings.Builder

	// enum comment
	if f.ReservedComments != "" {
		b.WriteString(CommentForBigOne(f.GetReservedComments()))
	}

	// head: enum definition
	b.WriteString("enum " + f.Enum.Name + " {")
	b.WriteString("\n")

	// body
	ls := make([]LineV2, 0, len(f.Values))
	for idx := range f.Values {
		ls = append(ls, f.value2Line(f.Values[idx]))
	}
	lns := NewLinesV2(tabSize, "\n", enumFieldDelimiter, ls...)
	lns.lastHasDelimiter = true
	lns.lastHasSep = true
	b.WriteString(lns.FormatThrift())

	// foot
	b.WriteString("}")
	b.WriteString("\n")

	return b.String()
}

func (f *EnumFormatter) Type() FormatType {
	return FormatTypeEnum
}

type TypedefFormatter struct {
	/*
		TYPEDEF FieldType Identifier
	*/
	Typedef
}

func (f *TypedefFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	// typedef 的注释只能写在定义上面
	if f.ReservedComments != "" {
		b.WriteString(CommentForBigOne(f.GetReservedComments()))
	}
	b.WriteString("typedef " + f.GetType().String() + " " + f.GetName())
	if f.GetReservedEndLineComments() != "" {
		b.WriteString(" // ")
		b.WriteString(f.GetReservedEndLineComments())
	}

	return b.String()
}

func (f *TypedefFormatter) Type() FormatType {
	return FormatTypeTypedef
}

type StructLikeFormatter struct {
	StructLike
}

func (f *StructLikeFormatter) FormatThrift() string {
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

func (f *StructLikeFormatter) Type() FormatType {
	return FormatTypeStructLike
}

type ConstantFormatter struct {
	/*
		CONST FieldType Identifier EQUAL ConstValue ListSeparator?
	*/
	Constant
}

func (f *ConstantFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	// constant 的注释只能写在定义上面
	if f.ReservedComments != "" {
		b.WriteString(CommentForBigOne(f.GetReservedComments()))
	}
	b.WriteString("const " + f.GetType().String() + " " + f.GetName() + " = " + ParseConstValue(f.GetValue()))
	// b.WriteString(fieldDelimiter)

	return b.String()
}

func (f *ConstantFormatter) Type() FormatType {
	return FormatTypeConstant
}

type IncludeFormatter struct {
	/*
		Include <- INCLUDE Literal
	*/
	Include
}

func (f *IncludeFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	b.WriteString("include " + "\"" + f.GetPath() + "\"")
	return b.String()
}

func (f *IncludeFormatter) Type() FormatType {
	return FormatTypeInclude
}

type CppIncludeFormatter struct {
	/*
		CppInclude <- 'cpp_include' Literal
	*/
	cppInclude string
}

func (c CppIncludeFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	b.WriteString("cpp_include " + "\"" + c.cppInclude + "\"")
	return b.String()
}

func (c CppIncludeFormatter) Type() FormatType {
	return FormatTypeCppIncludeInclude
}

type NamespaceFormatter struct {
	/*
		Namespace <- NAMESPACE NamespaceScope Identifier Annotations?
	*/
	Namespace
}

func (f *NamespaceFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	b.WriteString("namespace " + f.GetLanguage() + " " + f.GetName())
	if len(f.GetAnnotations()) > 0 {
		WriteAnnotations(f.GetAnnotations(), b)
	}
	return b.String()
}

func (f *NamespaceFormatter) Type() FormatType {
	return FormatTypeNamespace
}

type ServiceFormatter struct {
	/*
		Service <- SERVICE Identifier ( EXTENDS Identifier )? { Function* } Annotations?
		Function  <- ReservedComments Skip
					ONEWAY? FunctionType Identifier { Field* } Throws? Annotations? ListSeparator? ReservedEndLineComments SkipLine
	*/
	Service
}

func (f *ServiceFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	// comment
	if f.ReservedComments != "" {
		b.WriteString(CommentForBigOne(f.GetReservedComments()))
	}
	// -- header --
	b.WriteString("service " + f.GetName())
	// extends
	if f.GetExtends() != "" {
		b.WriteString(" extends " + f.GetExtends())
	}
	b.WriteString(" {")
	b.WriteString("\n")

	// functions
	b.WriteString(FormatFunctions(f.GetFunctions()))

	// -- footer --
	b.WriteString("}")
	if len(f.GetAnnotations()) > 0 {
		WriteAnnotations(f.GetAnnotations(), b)
	}
	b.WriteString("\n")

	return b.String()
}

func (f *ServiceFormatter) Type() FormatType {
	return FormatTypeService
}

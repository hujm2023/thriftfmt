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
	FormatTypeUnion             FormatType = "Union"
	FormatTypeException         FormatType = "Exception"
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

// FormatThrift formats the Enum to its Thrift representation.
func (f *EnumFormatter) FormatThrift() string {
	var b strings.Builder

	// enum comment (if any)
	if f.ReservedComments != "" {
		b.WriteString(CommentForBigOne(f.GetReservedComments()))
	}

	// head: enum definition
	b.WriteString("enum " + f.Enum.Name + " {")
	b.WriteString("\n")

	// body: enum values
	ls := make([]LineV2, 0, len(f.Values))
	for idx := range f.Values {
		ls = append(ls, f.value2Line(f.Values[idx]))
	}
	lns := NewLinesV2(tabSize, "\n", enumFieldDelimiter, ls...)
	lns.lastHasDelimiter = true
	lns.lastHasSep = true
	b.WriteString(lns.FormatThrift())

	// foot: closing brace
	b.WriteString("}")
	b.WriteString("\n")

	return b.String()
}

// Type returns the FormatType of the EnumFormatter.
func (f *EnumFormatter) Type() FormatType {
	return FormatTypeEnum
}

type TypedefFormatter struct {
	/*
		TYPEDEF FieldType Identifier
	*/
	Typedef
}

// FormatThrift formats the Typedef to its Thrift representation.
func (f *TypedefFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	// typedef comments can only be placed above the definition
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

// Type returns the FormatType of the TypedefFormatter.
func (f *TypedefFormatter) Type() FormatType {
	return FormatTypeTypedef
}

type StructLikeFormatter struct {
	StructLike
}

// FormatThrift formats the StructLike (struct, union, exception) to its Thrift representation.
func (f *StructLikeFormatter) FormatThrift() string {
	return FormatStructLike(f.StructLike)
}

// Type returns the FormatType of the StructLikeFormatter.
func (f *StructLikeFormatter) Type() FormatType {
	return FormatTypeStructLike
}

type ConstantFormatter struct {
	/*
		CONST FieldType Identifier EQUAL ConstValue ListSeparator?
	*/
	Constant
}

// FormatThrift formats the Constant to its Thrift representation.
func (f *ConstantFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	// constant comments can only be placed above the definition
	if f.ReservedComments != "" {
		b.WriteString(CommentForBigOne(f.GetReservedComments()))
	}
	b.WriteString("const " + f.GetType().String() + " " + f.GetName() + " = " + ParseConstValue(f.GetValue()))
	// b.WriteString(fieldDelimiter)

	return b.String()
}

// Type returns the FormatType of the ConstantFormatter.
func (f *ConstantFormatter) Type() FormatType {
	return FormatTypeConstant
}

type IncludeFormatter struct {
	/*
		Include <- INCLUDE Literal
	*/
	Include
}

// FormatThrift formats the Include directive to its Thrift representation.
func (f *IncludeFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	b.WriteString("include " + "\"" + f.GetPath() + "\"")
	return b.String()
}

// Type returns the FormatType of the IncludeFormatter.
func (f *IncludeFormatter) Type() FormatType {
	return FormatTypeInclude
}

type CppIncludeFormatter struct {
	/*
		CppInclude <- 'cpp_include' Literal
	*/
	cppInclude string
}

// FormatThrift formats the CppInclude directive to its Thrift representation.
func (c CppIncludeFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	b.WriteString("cpp_include " + "\"" + c.cppInclude + "\"")
	return b.String()
}

// Type returns the FormatType of the CppIncludeFormatter.
func (c CppIncludeFormatter) Type() FormatType {
	return FormatTypeCppIncludeInclude
}

type NamespaceFormatter struct {
	/*
		Namespace <- NAMESPACE NamespaceScope Identifier Annotations?
	*/
	Namespace
}

// FormatThrift formats the Namespace directive to its Thrift representation.
func (f *NamespaceFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	b.WriteString("namespace " + f.GetLanguage() + " " + f.GetName())
	if len(f.GetAnnotations()) > 0 {
		WriteAnnotations(f.GetAnnotations(), b)
	}
	return b.String()
}

// Type returns the FormatType of the NamespaceFormatter.
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

// FormatThrift formats the Service to its Thrift representation.
func (f *ServiceFormatter) FormatThrift() string {
	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)
	// service comment (if any)
	if f.ReservedComments != "" {
		b.WriteString(CommentForBigOne(f.GetReservedComments()))
	}
	// -- header: service definition --
	b.WriteString("service " + f.GetName())
	// extends (if any)
	if f.GetExtends() != "" {
		b.WriteString(" extends " + f.GetExtends())
	}
	b.WriteString(" {")
	b.WriteString("\n")

	// functions
	b.WriteString(FormatFunctions(f.GetFunctions()))

	// -- footer: closing brace and annotations --
	b.WriteString("}")
	if len(f.GetAnnotations()) > 0 {
		WriteAnnotations(f.GetAnnotations(), b)
	}
	b.WriteString("\n")

	return b.String()
}

// Type returns the FormatType of the ServiceFormatter.
func (f *ServiceFormatter) Type() FormatType {
	return FormatTypeService
}

type UnionFormatter struct {
	/*
		Union <- UNION Identifier { Field* }
	*/
	StructLike
}

// FormatThrift formats the Union to its Thrift representation.
// It utilizes the common FormatStructLike function.
func (f *UnionFormatter) FormatThrift() string {
	return FormatStructLike(f.StructLike)
}

// Type returns the FormatType of the UnionFormatter.
func (f *UnionFormatter) Type() FormatType {
	return FormatTypeUnion
}

type ExceptionFormatter struct {
	/*
		Exception <- EXCEPTION Identifier { Field* }
	*/
	StructLike
}

// FormatThrift formats the Exception to its Thrift representation.
// It utilizes the common FormatStructLike function.
func (e ExceptionFormatter) FormatThrift() string {
	return FormatStructLike(e.StructLike)
}

// Type returns the FormatType of the ExceptionFormatter.
func (e ExceptionFormatter) Type() FormatType {
	return FormatTypeException
}

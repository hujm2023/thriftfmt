// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testAnnotation = `
const string (a = "a") str = "str"

typedef map<i32, string> (cpp.template = "std::map") itoa_map (foo='bar')
typedef list<double (string.presentation = "hex")> float_list

enum Enum {
	E1 (value = "10"),
	E2
	E3 (value = "100")
} (eee = "eee")

struct s {
	1: string f1 ( a = "a" );
	2: string f2 ( a = "a", b = "" );
	3: string (str = "str") f3;
	4: string f4;
	5: string f5 ();
	6: string f6 (  );
} (
	xxx = "",
	yyy = "y",
	zzz = "zzz",
)

exception myerror {
  1: i32 error_code ( range = "<0" )
  2: string error_msg
} (hello = "world")

service test_service {
	i32 (what = "response-annotation") method() (what = "method-annotation")
} (
	what.is.this = "service.annotation",
	empty.annotation = "",
	another.one = "more"
)

`

func TestAnnotation(t *testing.T) {
	ast, err := ParseString("main.thrift", testAnnotation)
	assert.True(t, err == nil, err)

	has := func(m Annotations, k, v string) bool {
		vs := m.Get(k)
		for _, val := range vs {
			if v == val {
				return true
			}
		}
		return false
	}
	assert.True(t, len(ast.Constants) == 1)
	assert.True(t, len(ast.Constants[0].Annotations) == 0)
	assert.True(t, len(ast.Constants[0].Type.Annotations) == 1)
	assert.True(t, has(ast.Constants[0].Type.Annotations, "a", "a"))

	assert.True(t, len(ast.Typedefs) == 2)
	assert.True(t, len(ast.Typedefs[0].Annotations) == 1)
	assert.True(t, has(ast.Typedefs[0].Annotations, "foo", "bar"))
	assert.True(t, len(ast.Typedefs[0].Type.Annotations) == 1)
	assert.True(t, has(ast.Typedefs[0].Type.Annotations, "cpp.template", "std::map"))
	assert.True(t, len(ast.Typedefs[0].Type.KeyType.Annotations) == 0)
	assert.True(t, len(ast.Typedefs[0].Type.ValueType.Annotations) == 0)
	assert.True(t, len(ast.Typedefs[1].Annotations) == 0)
	assert.True(t, len(ast.Typedefs[1].Type.Annotations) == 0)
	assert.True(t, len(ast.Typedefs[1].Type.ValueType.Annotations) == 1)
	assert.True(t, has(ast.Typedefs[1].Type.ValueType.Annotations, "string.presentation", "hex"))

	assert.True(t, len(ast.Enums) == 1)
	assert.True(t, len(ast.Enums[0].Annotations) == 1)
	assert.True(t, has(ast.Enums[0].Annotations, "eee", "eee"))
	assert.True(t, len(ast.Enums[0].Values) == 3)
	assert.True(t, len(ast.Enums[0].Values[0].Annotations) == 1)
	assert.True(t, len(ast.Enums[0].Values[1].Annotations) == 0)
	assert.True(t, len(ast.Enums[0].Values[2].Annotations) == 1)
	assert.True(t, has(ast.Enums[0].Values[0].Annotations, "value", "10"))
	assert.True(t, has(ast.Enums[0].Values[2].Annotations, "value", "100"))

	assert.True(t, len(ast.Structs) == 1)
	assert.True(t, len(ast.Structs[0].Annotations) == 3)
	assert.True(t, has(ast.Structs[0].Annotations, "xxx", ""))
	assert.True(t, has(ast.Structs[0].Annotations, "yyy", "y"))
	assert.True(t, has(ast.Structs[0].Annotations, "zzz", "zzz"))
	assert.True(t, len(ast.Structs[0].Fields) == 6)
	assert.True(t, len(ast.Structs[0].Fields[0].Annotations) == 1)
	assert.True(t, len(ast.Structs[0].Fields[1].Annotations) == 2)
	assert.True(t, len(ast.Structs[0].Fields[2].Annotations) == 0)
	assert.True(t, len(ast.Structs[0].Fields[3].Annotations) == 0)
	assert.True(t, has(ast.Structs[0].Fields[0].Annotations, "a", "a"))
	assert.True(t, has(ast.Structs[0].Fields[1].Annotations, "a", "a"))
	assert.True(t, has(ast.Structs[0].Fields[1].Annotations, "b", ""))
	assert.True(t, has(ast.Structs[0].Fields[2].Type.Annotations, "str", "str"))
	assert.True(t, ast.Structs[0].Fields[4].Annotations == nil)
	assert.True(t, ast.Structs[0].Fields[5].Annotations == nil)

	assert.True(t, len(ast.Exceptions) == 1)
	assert.True(t, len(ast.Exceptions[0].Annotations) == 1)
	assert.True(t, has(ast.Exceptions[0].Annotations, "hello", "world"))
	assert.True(t, len(ast.Exceptions[0].Fields) == 2)
	assert.True(t, len(ast.Exceptions[0].Fields[0].Annotations) == 1)
	assert.True(t, len(ast.Exceptions[0].Fields[1].Annotations) == 0)
	assert.True(t, has(ast.Exceptions[0].Fields[0].Annotations, "range", "<0"))
	assert.True(t, len(ast.Services) == 1)
	assert.True(t, len(ast.Services[0].Annotations) == 3)
	assert.True(t, len(ast.Services[0].Functions) == 1)
	assert.True(t, len(ast.Services[0].Functions[0].Annotations) == 1)
	assert.True(t, len(ast.Services[0].Functions[0].FunctionType.Annotations) == 1)
	assert.True(t, has(ast.Services[0].Annotations, "what.is.this", "service.annotation"))
	assert.True(t, has(ast.Services[0].Annotations, "empty.annotation", ""))
	assert.True(t, has(ast.Services[0].Annotations, "another.one", "more"))
	assert.True(t, has(ast.Services[0].Functions[0].Annotations, "what", "method-annotation"))
	assert.True(t, has(ast.Services[0].Functions[0].FunctionType.Annotations, "what", "response-annotation"))
}

func TestLiteralEscape(t *testing.T) {
	ast, err := ParseString("main.thrift", `
const string str1 = "a\'b\"c\td\ve\nf\rg\\h"
const string str2 = 'a\'b\"c\td\ve\nf\rg\\h'
	`)
	assert.True(t, err == nil, err)
	assert.True(t, len(ast.Constants) == 2)
	assert.True(t, ast.Constants[0].Value.TypedValue.GetLiteral() == `a\'b"c\td\ve\nf\rg\\h`)
	assert.True(t, ast.Constants[1].Value.TypedValue.GetLiteral() == `a'b\"c\td\ve\nf\rg\\h`)
}

const testSpaceSkip = `
namespace
*
test
enum
Numbers
{
ONE
=
1
,
TWO
,
}
const
Numbers
MyNumber
=
ONE
typedef
i8
MyByte
struct
MyStruct
{
1
:
string
str
,
2
:
list
<
string
>
strList
}
service
MyService
{
list
<
string
>
getStrList
(
1
:
i64
id
,
)
}
`

const testCommentSkip = `
namespace /*c*/ * /*c*/test /*c*/ 
enum /*c*/ Numbers /*c*/ { /*c*/ ONE /*c*/ = /*c*/ 1 /*c*/ , /*c*/ TWO /*c*/ , /*c*/ } /*c*/ 
const /*c*/ Numbers /*c*/ MyNumber /*c*/ = /*c*/ ONE /*c*/ 
typedef /*c*/ i8 /*c*/ MyByte /*c*/ 
struct /*c*/ MyStruct /*c*/ { /*c*/ 1 /*c*/ : /*c*/ string /*c*/ str /*c*/ , /*c*/ 2 /*c*/ : /*c*/ list /*c*/ < /*c*/ string /*c*/ > /*c*/ strList /*c*/ } /*c*/ 
service /*c*/ MyService /*c*/ { /*c*/ list /*c*/ < /*c*/ string /*c*/ > /*c*/ getStrList /*c*/ ( /*c*/ 1 /*c*/ : /*c*/ i64 /*c*/ id /*c*/ , /*c*/ ) /*c*/ } /*c*/
`

func TestSkip(t *testing.T) {
	_, err := ParseString("main.thrift", testSpaceSkip)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseString("main.thrift", testCommentSkip)
	if err != nil {
		t.Fatal(err)
	}
}

const testEscape = `
const string str = "hello%s\nworld"

struct s {
	1: string f1 = "\"\'1\a2\\\t3\007本\u12e4456" (a = "vd:\"\'1\a2\\\t3\007本\u12e4456\"")
	2: string f2 = '\"\'1\a2\\\t3\007本\u12e4456' (a = 'vd:\"\'1\a2\\\t3\007本\u12e4456\"')
}
`

func TestEscape(t *testing.T) {
	ast, err := ParseString("main.thrift", testEscape)
	assert.True(t, err == nil, err)

	assert.True(t, len(ast.Constants) == 1)
	assert.True(t, *ast.Constants[0].Value.TypedValue.Literal == `hello%s\nworld`)

	assert.True(t, len(ast.Structs) == 1)
	assert.True(t, *ast.Structs[0].Fields[0].Default.TypedValue.Literal == `"\'1\a2\\\t3\007本\u12e4456`)
	assert.True(t, *ast.Structs[0].Fields[1].Default.TypedValue.Literal == `\"'1\a2\\\t3\007本\u12e4456`)
	assert.True(t, ast.Structs[0].Fields[0].Annotations[0].Values[0] == `vd:"\'1\a2\\\t3\007本\u12e4456"`)
	assert.True(t, ast.Structs[0].Fields[1].Annotations[0].Values[0] == `vd:\"'1\a2\\\t3\007本\u12e4456\"`)
}

const testEnum = `
enum A {}

enum B {
	B1
	B2,B3
}

enum C {
	C1 = 1

	C2 = 10

	C3

	C4 = 1
	C5,C6
}
`

func TestEnumValue(t *testing.T) {
	ast, err := ParseString("main.thrift", testEnum)
	assert.True(t, err == nil)

	assert.True(t, len(ast.Enums) == 3)
	e1, e2, e3 := ast.Enums[0], ast.Enums[1], ast.Enums[2]
	assert.True(t, len(e1.Values) == 0)
	assert.True(t, len(e2.Values) == 3)
	assert.True(t, len(e3.Values) == 6)
	assert.True(t, e2.Values[0].Value == 0)
	assert.True(t, e2.Values[1].Value == 1)
	assert.True(t, e2.Values[2].Value == 2)
	assert.True(t, e3.Values[0].Value == 1)
	assert.True(t, e3.Values[1].Value == 10)
	assert.True(t, e3.Values[2].Value == 11)
	assert.True(t, e3.Values[3].Value == 1)
	assert.True(t, e3.Values[4].Value == 2)
	assert.True(t, e3.Values[5].Value == 3)
}

const testNamespace = `
namespace * whatever
namespace go golang
namespace py python.org
`

func TestNamespace(t *testing.T) {
	ast, err := ParseString("main.thrift", testNamespace)
	assert.True(t, err == nil)
	assert.True(t, len(ast.Namespaces) == 3)
	assert.True(t, ast.Namespaces[0].Language == "*")
	assert.True(t, ast.Namespaces[0].Name == "whatever")
	assert.True(t, ast.Namespaces[1].Language == "go")
	assert.True(t, ast.Namespaces[1].Name == "golang")
	assert.True(t, ast.Namespaces[2].Language == "py")
	assert.True(t, ast.Namespaces[2].Name == "python.org")
}

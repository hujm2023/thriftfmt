package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Example(t *testing.T) {
	res, err := ParseFileAsFormatters("./testdata/demo.thrift", nil, false)
	if err != nil {
		panic(err)
	}
	t.Log("\n", FormatInline(res))
}

func TestEnumFormatter(t *testing.T) {
	_testEnum := `enum StatusCode {
    SUCCESS = 0,
    ERROR = 1,
    ConfigSystemErrorStatus = 1000,      //系统错误
    ConfigDBOperationErrorStatus = 1001, //数据库异常
    ConfigDataMissStatus = 1002,         //关联数据（e.g. 类似外键对应数据）不存在
    ConfigParamErrorStatus = 1003,       //参数非法
    ConfigExistsStatus = 1004,           //数据已存在
    ConfigNotExistStatus = 1005,         //数据不存在
    ConfigBatchErrorStatus = 1006,       //批量操作失败
}
`
	expected := `enum StatusCode {
    SUCCESS = 0,
    ERROR = 1,
    ConfigSystemErrorStatus = 1000,      // 系统错误
    ConfigDBOperationErrorStatus = 1001, // 数据库异常
    ConfigDataMissStatus = 1002,         // 关联数据（e.g. 类似外键对应数据）不存在
    ConfigParamErrorStatus = 1003,       // 参数非法
    ConfigExistsStatus = 1004,           // 数据已存在
    ConfigNotExistStatus = 1005,         // 数据不存在
    ConfigBatchErrorStatus = 1006,       // 批量操作失败
}
`
	p, err := parseFormatters("main.thrift", _testEnum, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(p))
	assert.Equal(t, expected, p[0].FormatThrift())
}

func TestStructFormatter(t *testing.T) {
	_testStruct := `struct Base {
    // 这是headComment
    1: string LogID = "", // 这是inlineContent
    2: string Caller = "",
    3: string Addr = "",
    4: string Client = "",
    5: optional TrafficEnv TrafficEnv, // 另一个comment

    /*
    如果是longComment，需要跟上一个字段空一行
    否则会被识别成上一个的字段的 lineComment
      目前不想解决-_-
    */
    6: optional map<string, string> Extra,
}
`
	expect := `struct Base {
    // 这是headComment
    1: string LogID = "",                 // 这是inlineContent
    2: string Caller = "",
    3: string Addr = "",
    4: string Client = "",
    5: optional TrafficEnv TrafficEnv,    // 另一个comment

    /*
    如果是longComment，需要跟上一个字段空一行
    否则会被识别成上一个的字段的 lineComment
      目前不想解决-_-
    */
    6: optional map<string,string> Extra,
}
`
	p, err := parseFormatters("main.thrift", _testStruct, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(p))
	res := p[0].FormatThrift()
	assert.Equal(t, expect, res)
}

func TestServiceFormatter(t *testing.T) {
	_testService := `service HelloService {
    // Ping says hello to you.
    BaseResp Ping(1: Base req) (api.get='/v1/ping')  // this comment will be ignored
}
`
	expect := `
service HelloService {
    // Ping says hello to you.
    BaseResp Ping (1: Base req) (api.get = "/v1/ping"), // this comment will be ignored
}
`
	p, err := parseFormatters("main.thrift", _testService, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(p))
	res := p[0].FormatThrift()
	assert.Equal(t, expect, "\n"+res)
}

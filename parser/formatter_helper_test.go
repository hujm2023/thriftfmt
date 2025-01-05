package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComment_FormatThrift(t *testing.T) {
	for _, item := range []struct {
		s        string
		expected string
	}{
		{s: "//这是注释", expected: "// 这是注释"},
		{s: "#这是注释", expected: "# 这是注释"},
		{s: "//     这是注释", expected: "// 这是注释"},
		{s: "//     这是注释     后面跟了一堆空格", expected: "// 这是注释 后面跟了一堆空格"},
		{s: "//     这是注释     后面跟了一堆空格   ", expected: "// 这是注释 后面跟了一堆空格"},
	} {
		assert.Equal(t, item.expected, Comment(item.s).FormatThrift())
	}
}

func TestCommentFormat(t *testing.T) {
	s := "\n/* \n---- 下面是新的定义----- \n  comment here\n */ \n"
	ln := LineV2{
		HeadComment: s,
	}
	lns := NewLinesV2(4, "\n", "", ln)
	res := lns.FormatThrift()
	t.Log("\n", res)
}

func TestFormatLines(t *testing.T) {
	t.Run("fields", func(t *testing.T) {
		expected := "(1:a_business.Ping req, 2:Base BB, 3:BaseResp CC)"
		lines := []LineV2{
			{
				Define:      "1:a_business.Ping req",
				TailComment: "",
			},
			{
				Define: "2:Base BB",
			},
			{
				Define: "3:BaseResp CC",
			},
		}
		lv2 := LinesV2{
			lines:                lines,
			tabSize:              0,
			delimiter:            ",",
			sep:                  " ",
			lastHasDelimiter:     false,
			addEmptyLineEachLine: false,
		}

		res := "(" + lv2.FormatThrift() + ")"
		// t.Log(res)
		assert.Equal(t, res, expected)
	})
}

package query

import (
	"net/url"
	"testing"
	"time"
)

func TestValues(t *testing.T) {
	testValues := make(url.Values)
	testValues.Add("index", "1")
	testValues.Add("index", "2")
	testValues.Add("token", "xxxx")

	createTime, _ := time.Parse(time.DateTime, "2011-11-11 11:11:11")

	tests := []struct {
		query interface{}
		want  string
	}{
		// string
		{
			query: "index=1&index=2&token=xxxx",
			want:  testValues.Encode(),
		},
		{
			query: "?index=1&index=2&token=xxxx&scope=&api=oauth",
			// sort by key
			want: "api=oauth&index=1&index=2&scope=&token=xxxx",
		},

		// map
		{
			query: map[string]string{
				"index": "1",
				"token": "xxxx",
			},
			want: "index=1&token=xxxx",
		},
		{
			query: map[string][]string{
				"index": []string{"1", "2"},
				"token": []string{"xxxx"},
			},
			want: "index=1&index=2&token=xxxx",
		},
		{
			query: map[any]any{
				"index": []int{1, 2},
				"token": []string{"xxx"},
			},
			want: "index=1&index=2&token=xxx",
		},

		// slice or array
		{
			query: []string{
				"index", "1",
				"index", "2",
				"token", "xxx",
			},
			want: "index=1&index=2&token=xxx",
		},
		{
			query: [7]any{
				"index", 1,
				"index", 2,
				"token", "xxx",
				"token-1",
			},
			want: "index=1&index=2&token=xxx",
		},

		// struct
		{
			query: struct {
				IsClean             bool
				IsCleanNum          bool      `query:",int"`
				IsCleanNumS         bool      `query:"-"`
				IsCleanNumSs        bool      `query:"-,"`
				Index               []string  `query:"index"`
				IndexByComma        []any     `query:"index_by_comma,del:comma"`
				IndexBySpace        []any     `query:"index_by_space,del:space"`
				IndexBySemicolon    []any     `query:"index_by_semicolon,del:semicolon"`
				IndexByBrackets     []any     `query:"index_by_brackets,del:brackets"`
				IndexByCustom       []any     `query:"index_by_custom,del:-"`
				Token               string    `query:"token,omitempty"`
				CreateTime          time.Time `query:"create_time,omitempty,time_format:2006-01-02 15:04:05"`
				CreateDate          time.Time `query:"create_date,time_format:2006-01-02,omitempty"`
				CreateTimeUnix      time.Time `query:"create_time_unix,unix,omitempty"`
				CreateTimeUnixmilli time.Time `query:"create_time_unixmilli,unixmilli,omitempty"`
				CreateTimeUnixnano  time.Time `query:"create_time_unixnano,unixnano,omitempty"`
				SubStruct           any       `query:"sub_struct,omitempty"`
				SubStructInline     any       `query:",inline,omitempty"`
			}{
				Index:               []string{"1", "2"},
				IndexByComma:        []any{3, 4},
				IndexBySpace:        []any{5, 6},
				IndexBySemicolon:    []any{7, 8},
				IndexByBrackets:     []any{9, 10},
				IndexByCustom:       []any{10, 11},
				Token:               "",
				IsCleanNum:          true,
				CreateTime:          createTime,
				CreateDate:          createTime,
				CreateTimeUnix:      createTime,
				CreateTimeUnixmilli: createTime,
				CreateTimeUnixnano:  createTime,
				SubStruct: struct {
					Name string `query:"name,omitempty"`
				}{
					Name: "ssss",
				},
				SubStructInline: struct {
					InlineName string `query:"inlineName,omitempty"`
				}{
					InlineName: "inlineName",
				},
			},
			want: "-=false&IsClean=false&IsCleanNum=1&create_date=2011-11-11&create_time=2011-11-11+11%3A11%3A11&create_time_unix=1321009871&create_time_unixmilli=1321009871000&create_time_unixnano=1321009871000000000&index=1&index=2&index_by_brackets%5B%5D=9&index_by_brackets%5B%5D=10&index_by_comma=3%2C4&index_by_custom=10-11&index_by_semicolon=7%3B8&index_by_space=5+6&inlineName=inlineName&sub_struct%5Bname%5D=ssss",
		},
	}

	for i, v := range tests {
		target, err := Values(v.query)
		if err == nil {
			en := target.Encode()
			if en != v.want {
				t.Logf("index: %d, Values() failed: target=%s want=%s", i, en, v.want)
			}
		} else {
			t.Logf("index: %d, Values() failed: err=%s", i, err)
		}
	}
}

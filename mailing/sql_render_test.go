package mailing

import (
	"fmt"
	"testing"
)

func TestRenderSQL(t *testing.T) {
	type testCase struct {
		name         string
		sqlKey       string
		relationName string
		dbType       string
		result       string
	}

	testCases := []*testCase{
		{
			name:         "pqsl get_users",
			sqlKey:       "get_users",
			relationName: "shazam.users",
			dbType:       PSQL,
			result:       "SELECT id, lang, advert_channel FROM shazam.users WHERE status = $1 OR status = '' ORDER BY id LIMIT $2;",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			renderQuery := renderSQL(tc.sqlKey, tc.relationName, tc.dbType)

			if renderQuery != tc.result {
				fmt.Println(renderQuery)
				fmt.Println(tc.result)
				t.Fail()
			}
		})
	}
}

func TestRenderSQL2(t *testing.T) {
	for key := range sqlQueries {
		fmt.Println(renderSQL(key, "shazam.users", PSQL))
		fmt.Println(renderSQL(key, "users", MySQL))
		fmt.Println()
	}
}

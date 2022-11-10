package mailing

import (
	"fmt"
	"strings"
)

const (
	MySQL = "mysql"
	PSQL  = "postgres"

	relationKey = "{{relation}}"
	variableKey = "{{variable}}"
)

var (
	sqlQueries = map[string]string{
		"get_users": `
SELECT id, lang, advert_channel
	FROM {{relation}}
WHERE status = {{variable}} OR status = ''
ORDER BY id
	LIMIT {{variable}};`,

		"count_mailing_users": `
SELECT COUNT(id) 
FROM {{relation}}
WHERE status = {{variable}};`,

		"mark_mailing_user": `
UPDATE {{relation}}
	SET status = {{variable}}
WHERE status = {{variable}}
	AND advert_channel = {{variable}};`,

		"mark_init_mailing_user": `
UPDATE {{relation}}
	SET status = {{variable}}
WHERE id = {{variable}};`,

		"get_init_mailing_users": `
SELECT id, lang, advert_channel
	FROM {{relation}}
WHERE status = {{variable}};`,

		"mark_active_user": `
UPDATE {{relation}}
	SET status = {{variable}}
WHERE id = {{variable}};`,
	}
)

func renderSQL(key, relationName, dbType string) string {
	baseSQL := sqlQueries[key]

	baseSQL = strings.Replace(baseSQL, relationKey, relationName, -1)

	return replaceVariables(baseSQL, dbType)
}

func replaceVariables(baseSQL, dbType string) string {
	baseSQL = strings.TrimPrefix(baseSQL, "\n")
	baseSQL = strings.TrimSuffix(baseSQL, ";")
	baseSQL = strings.Replace(baseSQL, "\n", " ", -1)
	baseSQL = strings.Replace(baseSQL, "\t", "", -1)
	splitSQL := strings.Split(baseSQL, " ")

	variableCount := 1
	for i, word := range splitSQL {
		if word != variableKey {
			continue
		}

		switch dbType {
		case MySQL:
			splitSQL[i] = "?"
		case PSQL:
			splitSQL[i] = fmt.Sprintf("$%d", variableCount)
		}

		variableCount++
	}

	return strings.Join(splitSQL, " ") + ";"
}

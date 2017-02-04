package main

import "text/template"

var (
	importTemplate = template.Must(template.New("prolog").Parse(`
		// Generated with genorm. DO NOT EDIT.
		package {{ index . 0 }}

		import (
			"database/sql"
			{{- index . 1 }}
		)
	`))

	bodyTemplate = template.Must(template.New("const").Parse(`
	const (
		{{- range $i, $f := . }}
	        {{- if $f.HasPK }}
		    select{{ $f.Type }}ByIdSql = "SELECT {{ $f.FieldsForSelect  }} FROM {{ $f.SQLObjectName }}{{ $f.WhereId }}"
	        {{- if $f.IsTable }}
		    insert{{ $f.Type }}Sql = "INSERT INTO {{ $f.SQLObjectName }} ( {{ $f.FieldsForInsert }} ) VALUES ( {{ $f.PlaceholdersForInsert }} ){{ $f.Returning }}"
		    update{{ $f.Type }}Sql = "WITH rows AS (UPDATE {{ $f.SQLObjectName }} SET {{ $f.FieldsForUpdate}} RETURNING 1) SELECT count(*) FROM rows"
		    {{- end }}
		    {{- end }}
		{{- end }}
	)

	var (
		{{- range $i, $f := . }}
	        {{- if $f.HasPK }}
		    select{{ $f.Type }}ByIdStmt *sql.Stmt
	        {{- if $f.IsTable }}
		    insert{{ $f.Type }}Stmt *sql.Stmt
		    update{{ $f.Type }}Stmt *sql.Stmt
		    {{- end }}
		    {{- end }}
		{{- end }}
	)

	{{- range $i, $f := . }}
	{{- if $f.HasPK }}
    func get{{ $f.Type }}ById(id {{ $f.ParamTypeSelectById }}) ({{ $f.FullType }}, error) {
	        t := {{ $f.FullType }}{}
			{{ $f.BeforeSelectById }}
            err := select{{ $f.Type }}ByIdStmt.QueryRow(id).Scan(
				{{ $f.ScanFields }})
            if err != nil {
				return t, err
			}
			{{ $f.AfterSelectById }}
	        return t, nil
	}

	{{ if $f.IsTable }}
	func insert{{ $f.Type }}(tx *sql.Tx, t *{{ $f.FullType }}) error {
	        row := tx.Stmt(insert{{ $f.Type }}Stmt).QueryRow(
				{{ $f.QueryRowInsert }})
            return row.Scan({{ $f.ScanInsertedId }})
	}

	func update{{ $f.Type }}(tx *sql.Tx, t *{{ $f.FullType }}) (int, error) {
	        row := tx.Stmt(update{{ $f.Type }}Stmt).QueryRow(
				{{ $f.QueryRowUpdate }})
            var cnt int
			if err := row.Scan(&cnt); err != nil {
				return 0, err
			}
			return cnt, nil
	}
	{{- end }}
	{{- end }}
	{{- end }}
	`))
)

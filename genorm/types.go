package main

import "fmt"

// FieldInfo represents information about struct field.
type FieldInfo struct {
	Name   string // field name as defined in source file, e.g. Name
	PKType string // primary key field type as defined in source file, e.g. string
	Column string // SQL database column name from "genorm:" struct field tag, e.g. name
}

// StructInfo represents information about struct.
// Comment "genorm:" in the following format:
// genorm:table_or_view_name[:view]
type StructInfo struct {
	Type      string      // struct type as defined in source file, e.g. User
	Pack      string      // package of struct type
	SQLSchema string      // SQL database schema name from "genorm:" comment, e.g. public
	SQLName   string      // SQL database view or table name from "genorm:" comment, e.g. users
	IsTable   bool        // This is a table or view from "genorm:" comment, e.g. view
	Fields    []FieldInfo // fields info
	PKIndex   int         // index of primary key field in Fields, -1 if none
}

func (si StructInfo) FullType() string {
	if si.Pack == "" {
		return si.Type
	}
	return si.Pack + "." + si.Type
}

func (si StructInfo) SQLObjectName() string {
	if si.SQLSchema == "" {
		return si.SQLName
	}
	return si.SQLSchema + "." + si.SQLName
}

func (si StructInfo) HasPK() bool {
	return si.PKIndex >= 0
}

func (si StructInfo) FieldsForSelect() string {
	fields := ""
	for _, f := range si.Fields {
		if fields == "" {
			fields = f.Column
		} else {
			fields += ", " + f.Column
		}
	}
	return fields
}

func (si StructInfo) FieldsForInsert() string {
	fields := ""
	for _, f := range si.Fields {
		if f.PKType != "" || f.Column == "updated_at" {
			continue
		}
		if fields == "" {
			fields = f.Column
		} else {
			fields += ", " + f.Column
		}
	}
	return fields
}

func (si StructInfo) FieldsForUpdate() string {
	i := 1
	fields := ""
	for _, f := range si.Fields {
		if f.PKType != "" || f.Column == "created_at" {
			continue
		}
		if fields == "" {
			fields = f.Column + fmt.Sprintf(" = $%d", i)
		} else {
			fields += ", " + f.Column + fmt.Sprintf(" = $%d", i)
		}
		i++
	}
	if si.PKIndex >= 0 {
		fields += " WHERE " + si.Fields[si.PKIndex].Column + fmt.Sprintf(" = $%d", i)
	}
	return fields
}

func (si StructInfo) PlaceholdersForInsert() string {
	i := 1
	placeholders := ""
	for _, f := range si.Fields {
		if f.PKType != "" || f.Column == "updated_at" {
			continue
		}
		if placeholders == "" {
			placeholders += fmt.Sprintf("$%d", i)
		} else {
			placeholders += ", " + fmt.Sprintf("$%d", i)
		}
		i++
	}
	return placeholders
}

func (si StructInfo) WhereId() string {
	if si.PKIndex == -1 {
		panic("WhereId method can only be called for table with primary key")
	}
	return " WHERE " + si.Fields[si.PKIndex].Column + " = $1"
}

func (si StructInfo) Returning() string {
	if si.PKIndex == -1 {
		panic("Returning method can only be called for table with primary key")
	}
	return " RETURNING " + si.Fields[si.PKIndex].Column
}

func (si StructInfo) ParamTypeSelectById() string {
	if si.PKIndex == -1 {
		panic("ParamTypeSelectById method can only be called for table with primary key")
	}
	return si.Fields[si.PKIndex].PKType
}

func (si StructInfo) BeforeSelectById() string {
	res := ""
	for _, f := range si.Fields {
		if f.Column == "created_at" || f.Column == "updated_at" {
			res += fmt.Sprintf("%s sql.NullInt64\n", f.Name)
		}
	}
	if res != "" {
		res = fmt.Sprintf("var (\n%s)", res)
	}
	return res
}

func (si StructInfo) ScanFields() string {
	enrich := func(f FieldInfo) string {
		if f.Column == "created_at" || f.Column == "updated_at" {
			return "&" + f.Name
		}
		return "&t." + f.Name
	}
	fields := ""
	for _, f := range si.Fields {
		if fields == "" {
			fields = enrich(f)
		} else {
			fields += ",\n" + enrich(f)
		}
	}
	return fields
}

func (si StructInfo) AfterSelectById() string {
	res := ""
	for _, f := range si.Fields {
		if f.Column == "created_at" || f.Column == "updated_at" {
			res += fmt.Sprintf("if %s.Valid {\nt.%s = %s.Int64\n}\n", f.Name, f.Name, f.Name)
		}
	}
	return res
}

func (si StructInfo) QueryRowInsert() string {
	enrich := func(f FieldInfo) string {
		if f.Column == "created_at" {
			return "time.Now().Unix()"
		}
		return "t." + f.Name
	}
	fields := ""
	for _, f := range si.Fields {
		if f.PKType != "" || f.Column == "updated_at" {
			continue
		}
		if fields == "" {
			fields = enrich(f)
		} else {
			fields += ",\n" + enrich(f)
		}
	}
	return fields
}

func (si StructInfo) ScanInsertedId() string {
	if si.PKIndex == -1 {
		panic("ScanInsertedId method can only be called for table with primary key")
	}
	return "&t." + si.Fields[si.PKIndex].Name
}

func (si StructInfo) QueryRowUpdate() string {
	enrich := func(f FieldInfo) string {
		if f.Column == "updated_at" {
			return "time.Now().Unix()"
		}
		return "t." + f.Name
	}
	fields := ""
	for _, f := range si.Fields {
		if f.PKType != "" || f.Column == "created_at" {
			continue
		}
		if fields == "" {
			fields = enrich(f)
		} else {
			fields += ",\n" + enrich(f)
		}
	}
	if si.PKIndex == -1 {
		panic("ScanInsertedId method can only be called for table with primary key")
	}
	fields += ",\nt." + si.Fields[si.PKIndex].Name
	return fields
}

func (si StructInfo) NeedImportTime() bool {
	for _, f := range si.Fields {
		if f.Column == "created_at" || f.Column == "updated_at" {
			return true
		}
	}
	return false
}

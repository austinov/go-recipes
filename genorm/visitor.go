package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"regexp"
	"strings"
)

var ormStructComment = regexp.MustCompile(`genorm:([0-9A-Za-z_\.:]+)`)

type visitor struct {
	pack     string
	dstPack  string
	file     *ast.File
	fset     *token.FileSet
	structs  map[string]StructInfo // key is type of struct
	notfound map[string][]string   // key is embedded type name, values are array of struct types where it wasn't found
	errors   []error
}

func (v *visitor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return v
	}
	switch node.(type) {
	case *ast.GenDecl:
		gd := node.(*ast.GenDecl)
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			doc := ts.Doc
			if doc == nil && len(gd.Specs) == 1 {
				doc = gd.Doc
			}
			if doc == nil {
				continue
			}
			sm := ormStructComment.FindStringSubmatch(doc.Text())
			if len(sm) < 2 {
				continue
			}
			parts := strings.SplitN(sm[1], ".", 2)
			var schema string
			if len(parts) == 2 {
				schema = parts[0]
			}
			table, isTable := parts[len(parts)-1], true
			parts = strings.SplitN(table, ":", 2)
			if len(parts) == 2 {
				table = parts[0]
				if parts[1] == "view" {
					isTable = false
				}
			}

			pack := ""
			if v.pack != v.dstPack {
				pack = v.pack
			}

			si := StructInfo{
				Type:      ts.Name.Name,
				Pack:      pack,
				SQLSchema: schema,
				SQLName:   table,
				IsTable:   isTable,
				PKIndex:   -1,
			}

			pkIndex := 0
			var fieldInfos []FieldInfo
			for _, field := range st.Fields.List {
				tag := obtainFieldTag(field.Tag)
				if tag == "" {
					continue
				}

				pkType := ""
				column, pk := parseFieldTag(tag)
				if pk {
					if si.PKIndex >= 0 {
						v.errors = append(v.errors, fmt.Errorf(`%s has field with duplicate "pk" label in tag`, si.Type))
						continue
					}
					switch t := field.Type.(type) {
					case *ast.Ident:
						pkType = t.String()
					}
					si.PKIndex = pkIndex
				}

				if field.Names != nil {
					fieldInfos = append(fieldInfos, FieldInfo{
						Name:   field.Names[0].Name,
						Column: column,
						PKType: pkType,
					})
				} else {
					embedType := fmt.Sprintf("%s", field.Type)
					if tp, ok := v.structs[embedType]; ok {
						if tp.PKIndex >= 0 && si.PKIndex >= 0 {
							v.errors = append(v.errors, fmt.Errorf(`%s has field with duplicate "pk" label in tag, it is not allowed`, embedType))
						}
						si.PKIndex = tp.PKIndex
						fieldInfos = append(fieldInfos, tp.Fields...)
					} else {
						structs, ok := v.notfound[embedType]
						if !ok {
							structs = make([]string, 0)
						}
						structs = append(structs, ts.Name.Name)
						// key is embedded type name (embedType),
						// values are array of types where it wasn't found
						v.notfound[embedType] = structs
					}
				}
				pkIndex++
			}
			si.Fields = fieldInfos
			v.structs[ts.Name.Name] = si
		}
	}
	return v
}

func obtainFieldTag(fieldTag *ast.BasicLit) string {
	// process fields with "genorm:" tag
	if fieldTag == nil {
		return ""
	}
	tag := fieldTag.Value
	if len(tag) < 3 {
		return ""
	}
	tag = reflect.StructTag(tag[1 : len(tag)-1]).Get("genorm")
	if len(tag) == 0 {
		return ""
	}
	return tag
}

func parseFieldTag(tag string) (columnName string, isPK bool) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 || len(parts) > 2 {
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "pk":
			isPK = true
		default:
			return
		}
	}

	columnName = parts[0]
	return
}

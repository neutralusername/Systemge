package Utilities

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
)

func ReflectTypeName(entity interface{}) string {
	reflectType := reflect.TypeOf(entity)
	return reflectType.Name()
}

func ReflectFieldNames(entity interface{}) []string {
	reflectType := reflect.TypeOf(entity)
	fieldNames := make([]string, reflectType.NumField())
	for i := 0; i < reflectType.NumField(); i++ {
		fieldNames[i] = reflectType.Field(i).Name
	}
	return fieldNames
}

func ReflectFieldType(entity interface{}, fieldName string) reflect.Type {
	reflectType := reflect.TypeOf(entity)
	typ, ok := reflectType.FieldByName(fieldName)
	if !ok {
		return nil
	}
	return typ.Type
}

func GetVariableNamesInGoFile(filePath string) []string {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// Extract variable names
	varNames := []string{}
	ast.Inspect(node, func(n ast.Node) bool {
		if decl, ok := n.(*ast.GenDecl); ok && decl.Tok == token.VAR {
			for _, spec := range decl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valueSpec.Names {
						varNames = append(varNames, name.Name)
					}
				}
			}
		}
		return true
	})
	return varNames
}

func GetVariableByFilePathAndName(filePath string, variableName string) (interface{}, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	variable := &ast.ValueSpec{}
	ast.Inspect(node, func(n ast.Node) bool {
		if decl, ok := n.(*ast.GenDecl); ok && decl.Tok == token.VAR {
			for _, spec := range decl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valueSpec.Names {
						if name.Name == variableName {
							variable = valueSpec
							return false
						}
					}
				}
			}
		}
		return true
	})
	return variable, nil
}

package preformBuilder

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

var (
	genericPathRx = regexp.MustCompile(`[\[\,]\*?(?:\[[0-9]*\])?\*?([^,\[\]]+)\.[a-zA-Z0-9_]+`)
)

func parseColType(tType reflect.Type, colName, schemaName string) (typeName string, importPath []string) {
	typeName = tType.String()
	var (
		types    = map[string]string{}
		pkgName  string
		ok       bool
		i        int
		typePath string
	)
	if strings.Contains(typeName, "main.CustomType_") {
		typeName = strings.Replace(strings.Replace(typeName, "main.CustomType_", "", 1), "_", "", -1)
	}
	if strings.Contains(typeName, "main.Enum_") {
		typeName = strings.Replace(strings.Replace(typeName, "main.Enum_", "", 1), "_", "", -1)
	}
	if tType.PkgPath() != "" {
		if strings.HasSuffix(tType.PkgPath(), "preform/types") {
			types[""] = tType.PkgPath()
		} else {
			types[hashPkgName(tType.PkgPath())] = tType.PkgPath()
		}
	}
	if matches := genericPathRx.FindAllStringSubmatch(tType.String(), -1); matches != nil {
		for _, match := range matches {
			if strings.Contains(match[1], ".") {
				if strings.HasSuffix(match[1], "preform/types") {
					types[""] = match[1]
					typeName = strings.Replace(typeName, match[0], strings.Replace(match[0], match[1], "preformTypes", 1), 1)
				} else {
					path := match[1]
					if strings.Contains(match[1], "%") {
						path, _ = url.PathUnescape(path)
					}
					pkgName = hashPkgName(path)
					if typePath, ok = types[pkgName]; ok && typePath != path {
						pkgName += fmt.Sprintf("%d", i)
						i++
					}
					types[pkgName] = path
					typeName = strings.Replace(typeName, match[0], strings.Replace(match[0], match[1], pkgName, 1), 1)
				}
			} else {
				types[match[1]] = match[1]
			}
		}
		for t, path := range types {
			if !strings.Contains(path, ".") {
				if t != "main" {
					importPath = append(importPath, path)
				}
			} else if t == "" {
				importPath = append(importPath, path)
			} else {
				importPath = append(importPath, fmt.Sprintf("%s %s", t, path))
			}
		}
	} else if tType.PkgPath() != "" && tType.PkgPath() != "main" {
		importPath = append(importPath, tType.PkgPath())
	}
	return
}

func hashPkgName(path string) string {
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	if len(name) <= 5 && len(parts) > 1 {
		name = parts[len(parts)-2] + " " + name
	}
	if strings.Contains(name, ".") {
		parts = strings.Split(name, ".")
		return parts[len(parts)-1]
	}
	return strcase.ToCamel(name)
}

package specsynth

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	reRoutePattern = regexp.MustCompile(`(?i)(?:app|router|group|wsGroup)\.(Get|Post|Put|Delete|Patch)\s*\(\s*["']([^"']+)["']`)
	reStructDef    = regexp.MustCompile(`type\s+([A-Za-z0-9_]+)\s+struct\s*\{([^}]+)\}`)
	reStructField  = regexp.MustCompile(`(?m)^\s*([A-Za-z0-9_]+)\s+([A-Za-z0-9_\[\]\*]+)(?:\s+` + "`" + `json:"([^",]+)[^` + "`" + `]*` + "`" + `)?`)
)

// ParseCodeEndpoints extracts HTTP route registrations and Go struct definitions from source code.
func ParseCodeEndpoints(sourceCode string) ([]EndpointSpec, []ModelDefinition) {
	var endpoints []EndpointSpec
	var models []ModelDefinition

	// 1. Extract Endpoints
	routeMatches := reRoutePattern.FindAllStringSubmatch(sourceCode, -1)
	for _, m := range routeMatches {
		if len(m) > 2 {
			method := strings.ToUpper(m[1])
			path := m[2]
			summary := fmt.Sprintf("%s %s", method, path)

			pathParams := make(map[string]string)
			parts := strings.Split(path, "/")
			for _, p := range parts {
				if strings.HasPrefix(p, ":") {
					paramName := strings.TrimPrefix(p, ":")
					pathParams[paramName] = "string"
				} else if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
					paramName := strings.Trim(p, "{}")
					pathParams[paramName] = "string"
				}
			}

			status := 200
			if method == "POST" {
				status = 201
			}

			endpoints = append(endpoints, EndpointSpec{
				Method:       method,
				Path:         path,
				Summary:      summary,
				PathParams:   pathParams,
				ResponseType: "StandardResponse",
				StatusCode:   status,
			})
		}
	}

	// 2. Extract Struct Models
	structMatches := reStructDef.FindAllStringSubmatch(sourceCode, -1)
	for _, sm := range structMatches {
		if len(sm) > 2 {
			structName := sm[1]
			body := sm[2]

			fields := make(map[string]string)
			var required []string

			fieldMatches := reStructField.FindAllStringSubmatch(body, -1)
			for _, fm := range fieldMatches {
				if len(fm) > 2 {
					goFieldName := fm[1]
					goType := fm[2]
					jsonName := goFieldName
					if len(fm) > 3 && fm[3] != "" {
						jsonName = fm[3]
					}
					fields[jsonName] = mapGoTypeToSchemaType(goType)
					required = append(required, jsonName)
				}
			}

			models = append(models, ModelDefinition{
				Name:   structName,
				Fields: fields,
				Required: required,
			})
		}
	}

	return endpoints, models
}

// InferSchemaFromJSON extracts property models and types from a sample JSON payload.
func InferSchemaFromJSON(payload string) (map[string]string, error) {
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(payload), &parsed)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]string)
	for k, v := range parsed {
		fields[k] = inferTypeFromValue(v)
	}
	return fields, nil
}

func mapGoTypeToSchemaType(goType string) string {
	clean := strings.TrimPrefix(goType, "*")
	switch clean {
	case "string":
		return "string"
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "time.Time":
		return "string (date-time)"
	default:
		if strings.HasPrefix(clean, "[]") {
			return "array"
		}
		return "object"
	}
}

func inferTypeFromValue(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case string:
		return "string"
	case float64, int, int64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "string"
	}
}


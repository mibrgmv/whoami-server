package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type OpenAPISpec struct {
	OpenAPI      string                   `json:"openapi,omitempty"`
	Swagger      string                   `json:"swagger,omitempty"`
	Info         map[string]interface{}   `json:"info"`
	Host         string                   `json:"host,omitempty"`
	BasePath     string                   `json:"basePath,omitempty"`
	Schemes      []string                 `json:"schemes,omitempty"`
	Consumes     []string                 `json:"consumes,omitempty"`
	Produces     []string                 `json:"produces,omitempty"`
	Paths        map[string]interface{}   `json:"paths"`
	Definitions  map[string]interface{}   `json:"definitions,omitempty"`
	Components   map[string]interface{}   `json:"components,omitempty"`
	Security     []map[string][]string    `json:"security,omitempty"`
	SecurityDefs map[string]interface{}   `json:"securityDefinitions,omitempty"`
	Tags         []map[string]interface{} `json:"tags,omitempty"`
}

func main() {
	docsDir := "./api/swagger"

	customSpec, err := readSwaggerFile(filepath.Join(docsDir, "custom", "swagger.json"))
	if err != nil {
		log.Fatalf("Failed to read custom swagger spec: %v", err)
	}

	err = filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".swagger.json") {
			grpcSpec, err := readSwaggerFile(path)
			if err != nil {
				log.Printf("Failed to read %s: %v", path, err)
				return nil
			}

			mergeSpecs(customSpec, grpcSpec)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to walk docs directory: %v", err)
	}

	mergedPath := filepath.Join(docsDir, "swagger.json")
	if err := writeSwaggerFile(mergedPath, customSpec); err != nil {
		log.Fatalf("Failed to write merged spec: %v", err)
	}

	fmt.Printf("Successfully merged specs into %s\n", mergedPath)
}

func readSwaggerFile(path string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

func writeSwaggerFile(path string, spec *OpenAPISpec) error {
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func mergeSpecs(base *OpenAPISpec, toMerge *OpenAPISpec) {
	if base.Paths == nil {
		base.Paths = make(map[string]interface{})
	}

	for path, pathSpec := range toMerge.Paths {
		isGRPCPath := strings.Contains(path, "/history") ||
			strings.Contains(path, "/quizzes") ||
			strings.Contains(path, "/questions")

		if isGRPCPath {
			if pathMap, ok := pathSpec.(map[string]interface{}); ok {
				for _, methodSpec := range pathMap {
					if methodMap, ok := methodSpec.(map[string]interface{}); ok {
						methodMap["security"] = []map[string][]string{
							{"BearerAuth": {}},
						}
					}
				}
			}
		}

		base.Paths[path] = pathSpec
	}

	if toMerge.Definitions != nil {
		if base.Definitions == nil {
			base.Definitions = make(map[string]interface{})
		}
		for def, defSpec := range toMerge.Definitions {
			base.Definitions[def] = defSpec
		}
	}

	if toMerge.Components != nil {
		if base.Components == nil {
			base.Components = make(map[string]interface{})
		}
		for comp, compSpec := range toMerge.Components {
			base.Components[comp] = compSpec
		}
	}

	if toMerge.Tags != nil {
		base.Tags = append(base.Tags, toMerge.Tags...)
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Impisigmatus/service_core/log"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func main() {
	// YAML записывается в stdout, поэтому сообщения логгера направляем в stderr.
	logger := log.New(log.LevelInfo).Output(os.Stderr)
	if len(os.Args) != 3 {
		logger.Error().Msg("использование: oapi-converter swagger-v2.yaml README.md")
		os.Exit(1)
	}

	output, err := convert(os.Args[1], os.Args[2])
	if err != nil {
		logger.Error().Err(err).Msg("Не удалось преобразовать Swagger в OpenAPI")
		os.Exit(1)
	}

	if _, err := os.Stdout.Write(output); err != nil {
		logger.Error().Err(err).Msg("Не удалось записать OpenAPI")
		os.Exit(1)
	}
}

func convert(swaggerPath, readmePath string) ([]byte, error) {
	source, err := readSwagger(swaggerPath)
	if err != nil {
		return nil, err
	}

	result, err := openapi2conv.ToV3(source)
	if err != nil {
		return nil, err
	}

	description, err := os.ReadFile(readmePath)
	if err != nil {
		return nil, fmt.Errorf("чтение README: %w", err)
	}

	result.Info.Description = string(description)
	setTextErrorMediaTypes(result)
	if err := result.Validate(context.Background()); err != nil {
		return nil, err
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var resultYAML map[string]any
	if err := yaml.Unmarshal(resultJSON, &resultYAML); err != nil {
		return nil, err
	}

	return yaml.Marshal(resultYAML)
}

// Swagger 2 задаёт produces для всей операции, а OpenAPI 3 — для каждого ответа.
func setTextErrorMediaTypes(doc *openapi3.T) {
	for _, path := range doc.Paths.Map() {
		for _, operation := range path.Operations() {
			for status, responseRef := range operation.Responses.Map() {
				code, err := strconv.Atoi(status)
				if err != nil || code < 400 || responseRef.Value == nil {
					continue
				}
				content := responseRef.Value.Content
				media := content["application/json"]
				if media == nil || media.Schema == nil || media.Schema.Value == nil || !media.Schema.Value.Type.Is("string") {
					continue
				}
				delete(content, "application/json")
				content["text/plain"] = media
			}
		}
	}
}

func readSwagger(swaggerPath string) (*openapi2.T, error) {
	input, err := os.ReadFile(swaggerPath)
	if err != nil {
		return nil, fmt.Errorf("чтение Swagger: %w", err)
	}

	var sourceYAML map[string]any
	if err := yaml.Unmarshal(input, &sourceYAML); err != nil {
		return nil, fmt.Errorf("разбор Swagger YAML: %w", err)
	}

	sourceJSON, err := json.Marshal(sourceYAML)
	if err != nil {
		return nil, err
	}

	var source openapi2.T
	if err := json.Unmarshal(sourceJSON, &source); err != nil {
		return nil, err
	}

	if source.Swagger != "2.0" {
		return nil, fmt.Errorf("ожидался Swagger 2.0, получен %q", source.Swagger)
	}
	return &source, nil
}

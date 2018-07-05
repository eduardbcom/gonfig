package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	ucl "github.com/cesanta/ucl"
	validateJSONSchema "github.com/cesanta/validate-json/schema"
)

type Schema struct {
	validator *validateJSONSchema.Validator
}

func (s *Schema) Validate(config *Config) error {
	var err error

	var rawConfig []byte
	var data ucl.Value

	rawConfig, err = json.Marshal(config.data)
	if err != nil {
		return err
	}

	data, err = ucl.Parse(bytes.NewReader(rawConfig))
	if err != nil {
		return err
	}

	return s.validator.Validate(data)
}

func NewSchema(schemaPath string) (*Schema, error) {
	files := make(map[string]string)

	if _files, err := ioutil.ReadDir(schemaPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	} else {
		for _, _file := range _files {
			files[_file.Name()] = path.Join(schemaPath, _file.Name())
		}
	}

	if _, ok := files["index.json"]; ok == false {
		return nil, errors.New("schema folder does not contain index.json file")
	}

	loader := validateJSONSchema.NewLoader()

	if schema, err := readSchema(files["index.json"]); err != nil {
		return nil, err
	} else {
		delete(files, "index.json")

		for _, schemaRefPath := range files {
			if schemaRef, err := readSchema(schemaRefPath); err != nil {
				return nil, err
			} else {
				loader.Add(schemaRef)
			}
		}

		if validator, err := validateJSONSchema.NewValidator(schema, loader); err != nil {
			return nil, err
		} else {
			return &Schema{validator: validator}, nil
		}
	}
}

func readSchema(filePath string) (parsedFileContent ucl.Value, err error) {
	var file *os.File

	file, err = os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %s", filePath, err)
	}

	defer file.Close()

	parsedFileContent, err = ucl.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema: %s", err)
	}

	return
}

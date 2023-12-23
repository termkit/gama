package workflow

import (
	py "github.com/termkit/gama/pkg/yaml"
)

type Workflow struct {
	// Content is a map of key and value designed for workflow_dispatch.inputs
	Content map[string]Content
}

type Content struct {
	Description string
	Type        string
	Required    bool

	// KeyValue is a map of key and value designed for JSONContent
	KeyValue *[]KeyValue

	// Choice is a map of key and value designed for Options
	Choice *Choice

	// Value is a map of string and value designed for string
	Value *Value
}

type KeyValue struct {
	Default string
	Key     string
	Value   string
}

type Value struct {
	Default string
	Value   string
}

type Choice struct {
	Default string
	Options []string
	Value   string
}

func ParseWorkflow(content py.WorkflowContent) (*Workflow, error) {
	w := &Workflow{
		Content: make(map[string]Content),
	}

	for key, value := range content.On.WorkflowDispatch.Inputs {
		if value.JSONContent != nil {
			var keyValue []KeyValue
			for k, v := range value.JSONContent {
				keyValue = append(keyValue, KeyValue{
					Key:     k,
					Value:   "",
					Default: v,
				})
			}

			w.Content[key] = Content{
				Description: value.Description,
				Type:        "json",
				Required:    value.Required,
				KeyValue:    &keyValue,
			}
		}

		if value.Type == "choice" {
			w.Content[key] = Content{
				Description: value.Description,
				Type:        "choice",
				Required:    value.Required,
				Choice: &Choice{
					Default: value.Default.(string),
					Options: value.Options,
					Value:   "",
				},
			}
		}

		if value.Type == "string" {
			w.Content[key] = Content{
				Description: value.Description,
				Type:        "string",
				Required:    value.Required,
				Value: &Value{
					Default: value.Default.(string),
					Value:   "",
				},
			}
		}
	}

	return w, nil
}

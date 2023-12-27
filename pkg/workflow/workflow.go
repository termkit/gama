package workflow

import (
	"encoding/json"

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
	Default any
	Value   any
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
		if value.JSONContent != nil && len(value.JSONContent) > 0 {
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
			continue // Skip the rest of the loop
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

		if value.Type == "string" || value.Type == "boolean" || value.Type == "number" || value.Type == "" {
			w.Content[key] = Content{
				Description: value.Description,
				Type:        "input",
				Required:    value.Required,
				Value: &Value{
					Default: value.Default,
					Value:   "",
				},
			}
		}
	}

	return w, nil
}

func (w *Workflow) ToPretty() Pretty {
	var pretty Pretty
	var id int
	for parent, data := range w.Content {
		if data.KeyValue != nil {
			for _, v := range *data.KeyValue {
				pretty.KeyVals = append(pretty.KeyVals, PrettyKeyValue{
					ID:      id,
					Parent:  stringPtr(parent),
					Key:     v.Key,
					Value:   "",
					Default: v.Default,
				})
				id++
			}
		}
		if data.Choice != nil {
			pretty.Choices = append(pretty.Choices, PrettyChoice{
				ID:      id,
				Key:     parent,
				Value:   "",
				Values:  data.Choice.Options,
				Default: data.Choice.Default,
			})
			id++
		}
		if data.Value != nil {
			pretty.Inputs = append(pretty.Inputs, PrettyInput{
				ID:      id,
				Key:     parent,
				Value:   "",
				Default: data.Value.Default,
			})
			id++
		}
	}

	return pretty
}

func (p *Pretty) ToJson() (string, error) {
	// Create a map to hold the aggregated data
	result := make(map[string]interface{})

	// Process KeyVals
	for _, kv := range p.KeyVals {
		if kv.Parent != nil {
			parent := *kv.Parent
			if _, ok := result[parent]; !ok {
				result[parent] = make(map[string]any)
			}
			result[parent].(map[string]any)[kv.Key] = kv.Value
		} else {
			result[kv.Key] = kv.Value
		}
	}

	// Process Choices
	for _, c := range p.Choices {
		result[c.Key] = c.Value
	}

	// Process Inputs
	for _, i := range p.Inputs {
		result[i.Key] = i.Value
	}

	// Marshal the aggregated data into JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", err // Return an empty string and the error
	}
	return string(jsonData), nil // Return the JSON string
}

type Pretty struct {
	Choices []PrettyChoice
	Inputs  []PrettyInput
	KeyVals []PrettyKeyValue
}

type PrettyChoice struct {
	ID      int
	Key     string
	Value   any
	Values  []string
	Default string
}

func (c *PrettyChoice) SetValue(value any) {
	c.Value = value
}

type PrettyInput struct {
	ID      int
	Key     string
	Value   any
	Default any
}

func (i *PrettyInput) SetValue(value any) {
	i.Value = value
}

type PrettyKeyValue struct {
	ID      int
	Parent  *string
	Key     string
	Value   any
	Default any
}

func (kv *PrettyKeyValue) SetValue(value any) {
	kv.Value = value
}

func stringPtr(s string) *string {
	return &s
}

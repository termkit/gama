package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	py "github.com/termkit/gama/pkg/yaml"
	"gopkg.in/yaml.v3"
)

func TestParseWorkflow(t *testing.T) {
	var data = []byte(`
name: SystemUpdateTrigger
on:
  workflow_dispatch:
    inputs:
      components:
        description: "JSON configuration for component versions"
        required: true        
        default: '{
          "main-engine-ref": "stable",
          "ui-layer-ref": "3",
          "data-handler-ref": "stable",
          "event-logger-ref": "main",
          "network-api-ref": "main",
          "analytics-service-ref": "main"
          }'           
      deployment_zone:
        description: 'Deployment Zone'
        type: choice
        required: true
        options:
          - 'alpha'
          - 'beta'
          - 'gamma'
          - 'delta'
          - 'epsilon'
          - 'trial'
        default: 'trial'
      industry_category:
        description: 'Industry Category'
        type: string
        required: true
        default: 'general'
    secrets: inherit
`)

	var workflow py.WorkflowContent
	err := yaml.Unmarshal(data, &workflow)

	assert.NoError(t, err)

	w, err := ParseWorkflow(workflow)
	assert.NoError(t, err)

	t.Log(w)
}

package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestWorkflowInput_UnmarshalYAML(t *testing.T) {
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
          - 'zeta'
        default: 'trial'
      industry_category:
        description: 'Industry Category'
        type: string
        required: true
        default: 'general'
    secrets: inherit
`)

	var workflow WorkflowContent
	err := yaml.Unmarshal(data, &workflow)

	assert.NoError(t, err)
	assert.Equal(t, "SystemUpdateTrigger", workflow.Name)
	assert.Equal(t, "JSON configuration for component versions", workflow.On.WorkflowDispatch.Inputs["components"].Description)
	assert.Equal(t, true, workflow.On.WorkflowDispatch.Inputs["components"].Required)
	assert.Equal(t, "trial", workflow.On.WorkflowDispatch.Inputs["deployment_zone"].Default)
	assert.Equal(t, "choice", workflow.On.WorkflowDispatch.Inputs["deployment_zone"].Type)
}

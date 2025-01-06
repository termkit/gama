package handler

import "github.com/charmbracelet/bubbles/table"

var tableColumnsGithubRepository = []table.Column{
	{Title: "Repository", Width: 24},
	{Title: "Default Branch", Width: 16},
	{Title: "Stars", Width: 6},
	{Title: "Workflows", Width: 9},
}

// ---------------------------------------------------------------------------

var tableColumnsTrigger = []table.Column{
	{Title: "ID", Width: 2},
	{Title: "Type", Width: 6},
	{Title: "Key", Width: 24},
	{Title: "Default", Width: 16},
	{Title: "Value", Width: 44},
}

// ---------------------------------------------------------------------------

var tableColumnsWorkflow = []table.Column{
	{Title: "Workflow", Width: 32},
	{Title: "File", Width: 48},
}

// ---------------------------------------------------------------------------

var tableColumnsWorkflowHistory = []table.Column{
	{Title: "Workflow", Width: 12},
	{Title: "Commit Message", Width: 16},
	{Title: "Triggered", Width: 12},
	{Title: "Started At", Width: 19},
	{Title: "Status", Width: 9},
	{Title: "Duration", Width: 8},
}

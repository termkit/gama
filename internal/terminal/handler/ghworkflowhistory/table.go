package ghworkflowhistory

import (
	"github.com/charmbracelet/bubbles/table"
)

var tableColumnsWorkflowHistory = []table.Column{
	{Title: "Workflow", Width: 12},
	{Title: "Commit Message", Width: 16},
	{Title: "Triggered", Width: 12},
	{Title: "Started At", Width: 19},
	{Title: "Status", Width: 9},
	{Title: "Duration", Width: 8},
}

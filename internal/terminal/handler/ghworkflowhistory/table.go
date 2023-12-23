package ghworkflowhistory

import (
	"github.com/charmbracelet/bubbles/table"
)

var tableColumnsWorkflowHistory = []table.Column{
	{Title: "Workflow", Width: 12},
	{Title: "Triggered", Width: 12},
	{Title: "Started At", Width: 22},
	{Title: "Status", Width: 9},
	{Title: "Duration", Width: 8},
}

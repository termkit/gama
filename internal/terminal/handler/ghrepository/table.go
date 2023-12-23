package ghrepository

import (
	"github.com/charmbracelet/bubbles/table"
)

var tableColumnsGithubRepository = []table.Column{
	{Title: "Repository", Width: 24},
	{Title: "Stars", Width: 6},
	{Title: "Workflows", Width: 9},
}

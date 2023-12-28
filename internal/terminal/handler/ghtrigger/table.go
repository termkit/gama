package ghtrigger

import (
	"github.com/charmbracelet/bubbles/table"
)

var tableColumnsTrigger = []table.Column{
	{Title: "ID", Width: 2},
	{Title: "Type", Width: 6},
	{Title: "Key", Width: 24},
	{Title: "Default", Width: 28},
	//{Title: "Description", Width: 64},
	{Title: "Value", Width: 28},
}

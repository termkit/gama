package repository

import "github.com/termkit/gama/pkg/pagination"

var listRepoSortValues = []string{
	"created",
	"updated",
	"pushed",
	"full_name",
}

func prepareListRepoPagination(input pagination.FindOpts) pagination.FindOpts {
	var limit = input.Limit
	if limit <= 0 || limit >= 100 {
		limit = 100
	}

	var page = uint(1)
	if input.Skip > 0 {
		var pageInt = input.Skip / limit
		if pageInt > 0 {
			page = uint(pageInt)
		}
	}

	if input.Sort == "" {
		input.Sort = "updated"
	}

	return pagination.FindOpts{
		Limit: limit,
		Skip:  int(page),
		Sort:  input.Sort,
	}
}

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
	if input.Limit == 0 {
		limit = 200
	}

	var page = uint(1)
	if input.Skip > 0 {
		var pageInt = input.Skip / limit
		if pageInt > 0 {
			page = uint(pageInt)
		}
	}

	return pagination.FindOpts{
		Limit: limit,
		Skip:  int(page),
		Sort:  input.Sort,
	}
}

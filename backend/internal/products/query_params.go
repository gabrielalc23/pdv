package products

import (
	"net/http"
	"strconv"
)

func parseListQuery(r *http.Request) (ListProductsInput, error) {
	query := r.URL.Query()

	input := ListProductsInput{
		Search:     query.Get("search"),
		CategoryID: query.Get("categoryId"),
	}

	if raw, ok := query["page"]; ok && len(raw) > 0 {
		page, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListProductsInput{}, newValidationError(
				"page",
				"must be a valid integer",
			)
		}

		input.Page = &page
	}

	if raw, ok := query["pageSize"]; ok && len(raw) > 0 {
		pageSize, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListProductsInput{}, newValidationError(
				"pageSize",
				"must be a valid integer",
			)
		}

		input.PageSize = &pageSize
	}

	if raw, ok := query["activeOnly"]; ok && len(raw) > 0 {
		activeOnly, err := strconv.ParseBool(raw[0])
		if err != nil {
			return ListProductsInput{}, newValidationError(
				"activeOnly",
				"must be true or false",
			)
		}

		input.ActiveOnly = activeOnly
	}

	return input, nil
}

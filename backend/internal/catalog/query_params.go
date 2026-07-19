package catalog

import (
	"net/http"
	"strconv"
)

func parseListQuery(r *http.Request) (ListCatalogInput, error) {
	query := r.URL.Query()

	input := ListCatalogInput{
		Search:     query.Get("search"),
		CategoryID: query.Get("categoryId"),
	}

	if raw, ok := query["page"]; ok && len(raw) > 0 {
		page, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListCatalogInput{}, newValidationError("page", "must be a valid integer")
		}
		input.Page = &page
	}

	if raw, ok := query["pageSize"]; ok && len(raw) > 0 {
		pageSize, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListCatalogInput{}, newValidationError("pageSize", "must be a valid integer")
		}
		input.PageSize = &pageSize
	}

	if raw, ok := query["activeOnly"]; ok && len(raw) > 0 {
		activeOnly, err := strconv.ParseBool(raw[0])
		if err != nil {
			return ListCatalogInput{}, newValidationError("activeOnly", "must be true or false")
		}
		input.ActiveOnly = activeOnly
		input.ActiveOnlySet = true
	}

	if raw, ok := query["inStockOnly"]; ok && len(raw) > 0 {
		inStockOnly, err := strconv.ParseBool(raw[0])
		if err != nil {
			return ListCatalogInput{}, newValidationError("inStockOnly", "must be true or false")
		}
		input.InStockOnly = inStockOnly
	}

	return input, nil
}

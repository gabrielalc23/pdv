package sales

import (
	"net/http"
	"strconv"
)

func parseListQuery(r *http.Request) (ListSalesInput, error) {
	query := r.URL.Query()

	input := ListSalesInput{
		Status: query.Get("status"),
	}

	if raw, ok := query["page"]; ok && len(raw) > 0 {
		page, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListSalesInput{}, newValidationError("page", "must be a valid integer")
		}

		input.Page = &page
	}

	if raw, ok := query["pageSize"]; ok && len(raw) > 0 {
		pageSize, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListSalesInput{}, newValidationError("pageSize", "must be a valid integer")
		}

		input.PageSize = &pageSize
	}

	return input, nil
}

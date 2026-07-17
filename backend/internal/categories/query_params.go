package categories

import (
	"net/http"
	"strconv"
)

func parseListQuery(r *http.Request) (ListCategoriesInput, error) {
	query := r.URL.Query()
	input := ListCategoriesInput{Search: query.Get("search"), ActiveOnly: true}
	if raw, ok := query["activeOnly"]; ok && len(raw) > 0 {
		activeOnly, err := strconv.ParseBool(raw[0])
		if err != nil {
			return ListCategoriesInput{}, newValidationError("activeOnly", "must be true or false")
		}
		input.ActiveOnly = activeOnly
	}
	return input, nil
}

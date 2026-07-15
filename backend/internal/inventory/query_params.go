package inventory

import (
	"net/http"
	"strconv"
)

func parseListInventoryQuery(r *http.Request) (ListInventoryInput, error) {
	query := r.URL.Query()
	input := ListInventoryInput{
		Search: query.Get("search"),
	}

	if raw, ok := query["page"]; ok && len(raw) > 0 {
		page, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListInventoryInput{}, newValidationError("page", "must be a valid integer")
		}
		input.Page = &page
	}

	if raw, ok := query["pageSize"]; ok && len(raw) > 0 {
		pageSize, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListInventoryInput{}, newValidationError("pageSize", "must be a valid integer")
		}
		input.PageSize = &pageSize
	}

	if raw, ok := query["activeOnly"]; ok && len(raw) > 0 {
		activeOnly, err := strconv.ParseBool(raw[0])
		if err != nil {
			return ListInventoryInput{}, newValidationError("activeOnly", "must be true or false")
		}
		input.ActiveOnly = activeOnly
	}

	return input, nil
}

func parseListMovementsQuery(r *http.Request) (ListInventoryMovementsInput, error) {
	query := r.URL.Query()
	input := ListInventoryMovementsInput{}

	if raw, ok := query["page"]; ok && len(raw) > 0 {
		page, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListInventoryMovementsInput{}, newValidationError("page", "must be a valid integer")
		}
		input.Page = &page
	}

	if raw, ok := query["pageSize"]; ok && len(raw) > 0 {
		pageSize, err := strconv.Atoi(raw[0])
		if err != nil {
			return ListInventoryMovementsInput{}, newValidationError("pageSize", "must be a valid integer")
		}
		input.PageSize = &pageSize
	}

	if typeValue := query.Get("type"); typeValue != "" {
		input.Type = typeValue
	}

	return input, nil
}

package models

import "encoding/json"

// ShortenRequest represents POST /api/shorten request body.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse represents POST /api/shorten response body.
type ShortenResponse struct {
	Result string `json:"result"`
}

// ShortenBatchRequest represents POST /api/shorten/batch request body.
type ShortenBatchRequest struct {
	URLs []ShortenBatchRequestItem
}

// UnmarshalJSON is an implementation of json.Unmarshaler interface.
func (r *ShortenBatchRequest) UnmarshalJSON(data []byte) error {
	tmp := make([]json.RawMessage, 0)
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	items := make([]ShortenBatchRequestItem, 0)
	for _, rawItem := range tmp {
		var requestItem ShortenBatchRequestItem
		if err := json.Unmarshal(rawItem, &requestItem); err != nil {
			return err
		}
		items = append(items, requestItem)
	}

	r.URLs = items

	return nil
}

// ShortenBatchRequestItem represents a single item in POST /api/shorten/batch request body.
type ShortenBatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ShortenBatchResponse represents POST /api/shorten/batch response body.
type ShortenBatchResponse struct {
	URLs []ShortenBatchResponseItem
}

// MarshalJSON is an implementation of json.Marshaler interface.
func (r *ShortenBatchResponse) MarshalJSON() ([]byte, error) {
	list := make([]json.RawMessage, 0, len(r.URLs))
	for _, url := range r.URLs {
		encoded, err := json.Marshal(url)
		if err != nil {
			return nil, err
		}
		list = append(list, encoded)
	}

	return json.Marshal(list)
}

// UnmarshalJSON is an implementation of json.Unmarshaler interface.
func (r *ShortenBatchResponse) UnmarshalJSON(data []byte) error {
	tmp := make([]json.RawMessage, 0)
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	items := make([]ShortenBatchResponseItem, 0)
	for _, rawItem := range tmp {
		var responseItem ShortenBatchResponseItem
		if err := json.Unmarshal(rawItem, &responseItem); err != nil {
			return err
		}
		items = append(items, responseItem)
	}

	r.URLs = items

	return nil
}

// ShortenBatchResponseItem represents a single item in POST /api/shorten/batch response body.
type ShortenBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ListByUserIDResponse represents GET /api/user/urls response body.
type ListByUserIDResponse struct {
	URLs []UserURLItem
}

// MarshalJSON is an implementation of json.Marshaler interface.
func (r *ListByUserIDResponse) MarshalJSON() ([]byte, error) {
	list := make([]json.RawMessage, 0, len(r.URLs))
	for _, url := range r.URLs {
		encoded, err := json.Marshal(url)
		if err != nil {
			return nil, err
		}
		list = append(list, encoded)
	}

	return json.Marshal(list)
}

// UnmarshalJSON is an implementation of json.Unmarshaler interface.
func (r *ListByUserIDResponse) UnmarshalJSON(data []byte) error {
	tmp := make([]json.RawMessage, 0)
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	items := make([]UserURLItem, 0)
	for _, rawItem := range tmp {
		var responseItem UserURLItem
		if err := json.Unmarshal(rawItem, &responseItem); err != nil {
			return err
		}
		items = append(items, responseItem)
	}

	r.URLs = items

	return nil
}

// UserURLItem represents a single item in GET /api/user/urls response body.
type UserURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// DeleteByUserIDRequest represents DELETE /api/user/urls request body.
type DeleteByUserIDRequest struct {
	Slugs []string
}

// UnmarshalJSON is an implementation of json.Unmarshaler interface.
func (r *DeleteByUserIDRequest) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Slugs)
}

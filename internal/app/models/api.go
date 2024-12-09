package models

import "encoding/json"

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenBatchRequest struct {
	URLs []ShortenBatchRequestItem
}

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

type ShortenBatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchResponse struct {
	URLs []ShortenBatchResponseItem
}

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

type ShortenBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ListByUserIDResponse struct {
	URLs []UserURLItem
}

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

type UserURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type DeleteByUserIDRequest struct {
	Slugs []string
}

func (r *DeleteByUserIDRequest) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Slugs)
}

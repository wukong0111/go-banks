package models

type APIResponse[T any] struct {
	Success    bool        `json:"success"`
	Data       T           `json:"data,omitempty"`
	Error      *string     `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

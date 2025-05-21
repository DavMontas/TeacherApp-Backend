package main

type PaginatedXQuery struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Sort   int `json:"sort"`
}


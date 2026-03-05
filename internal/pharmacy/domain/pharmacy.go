package domain

import "github.com/google/uuid"

// Pharmacy is a pure domain entity and must not carry transport or DB tags.
type Pharmacy struct {
	ID           uuid.UUID
	CNPJ         string
	Name         string
	Address      string
	Neighborhood string
	City         string
	State        string
}

type PharmacyFilters struct {
	State        string
	City         string
	Neighborhood string
	Page         int
	Limit        int
}

type Pagination struct {
	Page        int
	Limit       int
	Total       int
	TotalPages  int
	HasNextPage bool
	HasPrevPage bool
}

type PharmacyPage struct {
	Data       []Pharmacy
	Pagination Pagination
}

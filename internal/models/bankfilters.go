package models

// CountryFilter representa un filtro de país con conteo
type CountryFilter struct {
	Code  string `json:"code" example:"ES"`
	Name  string `json:"name" example:"España"`
	Count int    `json:"count" example:"45"`
}

// APIFilter representa un filtro de API con conteo
type APIFilter struct {
	Type  string `json:"type" example:"berlin_group"`
	Count int    `json:"count" example:"120"`
}

// BankGroupFilter representa un filtro de grupo bancario con conteo
type BankGroupFilter struct {
	GroupID string `json:"group_id" example:"santander_group"`
	Name    string `json:"name" example:"Grupo Santander"`
	Count   int    `json:"count" example:"8"`
}

// BankFilters contiene todos los filtros disponibles para bancos
type BankFilters struct {
	Countries    []CountryFilter   `json:"countries"`
	APIs         []APIFilter       `json:"apis"`
	Environments []string          `json:"environments"`
	BankGroups   []BankGroupFilter `json:"bankGroups"`
}

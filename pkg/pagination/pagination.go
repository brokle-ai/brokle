package pagination

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Pagination represents pagination information
type Pagination struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
	HasNext   bool  `json:"has_next"`
	HasPrev   bool  `json:"has_prev"`
	Offset    int   `json:"offset,omitempty"`
}

// PaginationParams represents pagination parameters from request
type PaginationParams struct {
	Page     int    `json:"page" form:"page" query:"page"`
	PageSize int    `json:"page_size" form:"page_size" query:"page_size"`
	SortBy   string `json:"sort_by" form:"sort_by" query:"sort_by"`
	SortDir  string `json:"sort_dir" form:"sort_dir" query:"sort_dir"`
}

// SortDirection represents sort direction
type SortDirection string

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

// Constants for pagination limits
const (
	DefaultPage         = 1
	DefaultPageSize     = 10
	MaxPageSize         = 100
	MinPageSize         = 1
	DefaultSortBy       = "created_at"
	DefaultSortDir      = "DESC"
)

// New creates a new Pagination instance
func New(page, pageSize int, total int64) *Pagination {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < MinPageSize {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	totalPage := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPage < 1 {
		totalPage = 1
	}

	offset := (page - 1) * pageSize

	return &Pagination{
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: totalPage,
		HasNext:   page < totalPage,
		HasPrev:   page > 1,
		Offset:    offset,
	}
}

// NewFromParams creates pagination from request parameters
func NewFromParams(params PaginationParams, total int64) *Pagination {
	return New(params.Page, params.PageSize, total)
}

// ParseParams parses pagination parameters from query string values
func ParseParams(page, pageSize, sortBy, sortDir string) PaginationParams {
	params := PaginationParams{
		Page:     DefaultPage,
		PageSize: DefaultPageSize,
		SortBy:   DefaultSortBy,
		SortDir:  DefaultSortDir,
	}

	// Parse page
	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	// Parse page size
	if pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			params.PageSize = ps
		}
	}

	// Parse sort by
	if sortBy != "" && IsValidSortField(sortBy) {
		params.SortBy = sortBy
	}

	// Parse sort direction
	if sortDir != "" {
		sortDir = strings.ToUpper(sortDir)
		if sortDir == "ASC" || sortDir == "DESC" {
			params.SortDir = sortDir
		}
	}

	return params
}

// Validate validates pagination parameters
func (p *PaginationParams) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}
	
	if p.PageSize < MinPageSize {
		return fmt.Errorf("page_size must be at least %d", MinPageSize)
	}
	
	if p.PageSize > MaxPageSize {
		return fmt.Errorf("page_size cannot exceed %d", MaxPageSize)
	}
	
	if p.SortBy != "" && !IsValidSortField(p.SortBy) {
		return fmt.Errorf("invalid sort field: %s", p.SortBy)
	}
	
	if p.SortDir != "" && p.SortDir != "ASC" && p.SortDir != "DESC" {
		return fmt.Errorf("sort direction must be ASC or DESC")
	}
	
	return nil
}

// Normalize normalizes pagination parameters to default values if invalid
func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	
	if p.PageSize < MinPageSize || p.PageSize > MaxPageSize {
		p.PageSize = DefaultPageSize
	}
	
	if p.SortBy == "" || !IsValidSortField(p.SortBy) {
		p.SortBy = DefaultSortBy
	}
	
	if p.SortDir != "ASC" && p.SortDir != "DESC" {
		p.SortDir = DefaultSortDir
	}
}

// GetOffset returns the offset for database queries
func (p *Pagination) GetOffset() int {
	return p.Offset
}

// GetLimit returns the limit for database queries
func (p *Pagination) GetLimit() int {
	return p.PageSize
}

// GetOrderBy returns the ORDER BY clause for SQL queries
func (p *PaginationParams) GetOrderBy() string {
	return fmt.Sprintf("%s %s", p.SortBy, p.SortDir)
}

// GetSQLClause returns the complete SQL pagination clause
func (p *PaginationParams) GetSQLClause() string {
	offset := (p.Page - 1) * p.PageSize
	return fmt.Sprintf("ORDER BY %s LIMIT %d OFFSET %d", 
		p.GetOrderBy(), p.PageSize, offset)
}

// NextPage returns parameters for the next page
func (p *Pagination) NextPage() *PaginationParams {
	if !p.HasNext {
		return nil
	}
	
	return &PaginationParams{
		Page:     p.Page + 1,
		PageSize: p.PageSize,
	}
}

// PrevPage returns parameters for the previous page
func (p *Pagination) PrevPage() *PaginationParams {
	if !p.HasPrev {
		return nil
	}
	
	return &PaginationParams{
		Page:     p.Page - 1,
		PageSize: p.PageSize,
	}
}

// FirstPage returns parameters for the first page
func (p *Pagination) FirstPage() *PaginationParams {
	return &PaginationParams{
		Page:     1,
		PageSize: p.PageSize,
	}
}

// LastPage returns parameters for the last page
func (p *Pagination) LastPage() *PaginationParams {
	return &PaginationParams{
		Page:     p.TotalPage,
		PageSize: p.PageSize,
	}
}

// GetPageRange returns a range of page numbers for pagination UI
func (p *Pagination) GetPageRange(rangeSize int) []int {
	if rangeSize <= 0 {
		rangeSize = 5
	}
	
	start := p.Page - rangeSize/2
	if start < 1 {
		start = 1
	}
	
	end := start + rangeSize - 1
	if end > p.TotalPage {
		end = p.TotalPage
		start = end - rangeSize + 1
		if start < 1 {
			start = 1
		}
	}
	
	var pages []int
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	
	return pages
}

// ToMap converts pagination to map for JSON serialization
func (p *Pagination) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"page":       p.Page,
		"page_size":  p.PageSize,
		"total":      p.Total,
		"total_page": p.TotalPage,
		"has_next":   p.HasNext,
		"has_prev":   p.HasPrev,
		"offset":     p.Offset,
	}
}

// Links represents pagination links
type Links struct {
	First *string `json:"first,omitempty"`
	Prev  *string `json:"prev,omitempty"`
	Next  *string `json:"next,omitempty"`
	Last  *string `json:"last,omitempty"`
}

// GenerateLinks generates pagination links for API responses
func (p *Pagination) GenerateLinks(baseURL string, params PaginationParams) *Links {
	links := &Links{}
	
	// First page link
	if p.Page > 1 {
		firstURL := fmt.Sprintf("%s?page=1&page_size=%d", baseURL, p.PageSize)
		if params.SortBy != "" {
			firstURL += fmt.Sprintf("&sort_by=%s&sort_dir=%s", params.SortBy, params.SortDir)
		}
		links.First = &firstURL
	}
	
	// Previous page link
	if p.HasPrev {
		prevURL := fmt.Sprintf("%s?page=%d&page_size=%d", baseURL, p.Page-1, p.PageSize)
		if params.SortBy != "" {
			prevURL += fmt.Sprintf("&sort_by=%s&sort_dir=%s", params.SortBy, params.SortDir)
		}
		links.Prev = &prevURL
	}
	
	// Next page link
	if p.HasNext {
		nextURL := fmt.Sprintf("%s?page=%d&page_size=%d", baseURL, p.Page+1, p.PageSize)
		if params.SortBy != "" {
			nextURL += fmt.Sprintf("&sort_by=%s&sort_dir=%s", params.SortBy, params.SortDir)
		}
		links.Next = &nextURL
	}
	
	// Last page link
	if p.Page < p.TotalPage {
		lastURL := fmt.Sprintf("%s?page=%d&page_size=%d", baseURL, p.TotalPage, p.PageSize)
		if params.SortBy != "" {
			lastURL += fmt.Sprintf("&sort_by=%s&sort_dir=%s", params.SortBy, params.SortDir)
		}
		links.Last = &lastURL
	}
	
	return links
}

// CursorPagination represents cursor-based pagination
type CursorPagination struct {
	Cursor   string `json:"cursor,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasNext  bool   `json:"has_next"`
	HasPrev  bool   `json:"has_prev"`
	PageSize int    `json:"page_size"`
}

// NewCursorPagination creates a new cursor-based pagination
func NewCursorPagination(cursor, nextCursor, prevCursor string, hasNext, hasPrev bool, pageSize int) *CursorPagination {
	if pageSize < MinPageSize {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	
	return &CursorPagination{
		Cursor:     cursor,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		PageSize:   pageSize,
	}
}

// Valid sort fields for the Brokle platform
var validSortFields = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"name":         true,
	"email":        true,
	"status":       true,
	"type":         true,
	"provider":     true,
	"model":        true,
	"cost":         true,
	"tokens":       true,
	"latency":      true,
	"usage":        true,
	"requests":     true,
	"errors":       true,
	"success_rate": true,
	"timestamp":    true,
}

// IsValidSortField checks if a sort field is valid
func IsValidSortField(field string) bool {
	return validSortFields[field]
}

// AddValidSortField adds a new valid sort field
func AddValidSortField(field string) {
	validSortFields[field] = true
}

// RemoveValidSortField removes a valid sort field
func RemoveValidSortField(field string) {
	delete(validSortFields, field)
}

// GetValidSortFields returns all valid sort fields
func GetValidSortFields() []string {
	var fields []string
	for field := range validSortFields {
		fields = append(fields, field)
	}
	return fields
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination"`
	Links      *Links      `json:"links,omitempty"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, pagination *Pagination, links *Links) *PaginatedResponse {
	return &PaginatedResponse{
		Data:       data,
		Pagination: pagination,
		Links:      links,
	}
}

// Quick helper functions

// Quick creates pagination with minimal parameters
func Quick(page, pageSize int, total int64) *Pagination {
	return New(page, pageSize, total)
}

// QuickParams creates pagination parameters with defaults
func QuickParams(page, pageSize int) PaginationParams {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < MinPageSize || pageSize > MaxPageSize {
		pageSize = DefaultPageSize
	}
	
	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		SortBy:   DefaultSortBy,
		SortDir:  DefaultSortDir,
	}
}

// IsLastPage checks if current page is the last page
func (p *Pagination) IsLastPage() bool {
	return p.Page >= p.TotalPage
}

// IsFirstPage checks if current page is the first page
func (p *Pagination) IsFirstPage() bool {
	return p.Page == 1
}

// GetStartItem returns the starting item number for current page
func (p *Pagination) GetStartItem() int {
	if p.Total == 0 {
		return 0
	}
	return p.Offset + 1
}

// GetEndItem returns the ending item number for current page
func (p *Pagination) GetEndItem() int {
	end := p.Offset + p.PageSize
	if end > int(p.Total) {
		end = int(p.Total)
	}
	return end
}

// GetDisplayText returns pagination display text like "1-10 of 50 results"
func (p *Pagination) GetDisplayText() string {
	if p.Total == 0 {
		return "No results"
	}
	
	start := p.GetStartItem()
	end := p.GetEndItem()
	
	if p.Total == 1 {
		return "1 result"
	}
	
	return fmt.Sprintf("%d-%d of %d results", start, end, p.Total)
}
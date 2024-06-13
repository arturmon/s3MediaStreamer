package trackhandler

import (
	"fmt"
	"strings"
)

func generatePaginationLinks(baseURL, basePath string, currentPage, totalPages int, pageSize string) string {
	var links []string

	if currentPage > 1 {
		prevLink := fmt.Sprintf("%s%s?page=%d&page_size=%s", baseURL, basePath, currentPage-1, pageSize)
		links = append(links, fmt.Sprintf("<%s>; rel=\"prev\"", prevLink))
	}

	if currentPage < totalPages {
		nextLink := fmt.Sprintf("%s%s?page=%d&page_size=%s", baseURL, basePath, currentPage+1, pageSize)
		links = append(links, fmt.Sprintf("<%s>; rel=\"next\"", nextLink))
	}

	if totalPages > 0 {
		firstLink := fmt.Sprintf("%s%s?page=%d&page_size=%s", baseURL, basePath, 1, pageSize)
		lastLink := fmt.Sprintf("%s%s?page=%d&page_size=%s", baseURL, basePath, totalPages, pageSize)
		links = append(links, fmt.Sprintf("<%s>; rel=\"first\"", firstLink))
		links = append(links, fmt.Sprintf("<%s>; rel=\"last\"", lastLink))
	}

	return strings.Join(links, ", ")
}

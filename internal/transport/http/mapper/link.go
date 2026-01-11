package mapper

import (
	"fmt"

	"code/internal/domain/link"
	"code/internal/transport/http/dto"
)

func ToLinkResponse(l link.Link, baseURL string) dto.LinkResponse {
	return dto.LinkResponse{
		ID:          l.ID,
		OriginalURL: l.OriginalURL,
		ShortName:   l.ShortName,
		ShortURL:    fmt.Sprintf("%s/r/%s", baseURL, l.ShortName),
		CreatedAt:   l.CreatedAt,
	}
}

func ToLinkResponseList(links []link.Link, baseURL string) []dto.LinkResponse {
	result := make([]dto.LinkResponse, 0, len(links))
	for _, l := range links {
		result = append(result, ToLinkResponse(l, baseURL))
	}
	return result
}

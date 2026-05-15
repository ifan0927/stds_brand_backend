package http

import (
	"context"
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ifan0927/stds_brand_backend/internal/application"
)

const faqTimeout = 2 * time.Second

// FAQService provides approved FAQ items to HTTP handlers.
type FAQService interface {
	// ListFAQItems returns approved FAQ items.
	ListFAQItems(ctx context.Context) ([]application.FAQItem, error)
}

type faqResponse struct {
	Items []faqItemPayload `json:"items"`
}

type faqItemPayload struct {
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	SortOrder int    `json:"sort_order"`
}

func handleFAQs(service FAQService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if service == nil {
			writeServiceUnavailableError(c)
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), faqTimeout)
		defer cancel()

		items, err := service.ListFAQItems(ctx)
		if err != nil {
			writeServiceUnavailableError(c)
			return
		}

		c.JSON(nethttp.StatusOK, faqResponse{Items: faqPayloadsFromApplication(items)})
	}
}

func faqPayloadsFromApplication(items []application.FAQItem) []faqItemPayload {
	payloads := make([]faqItemPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, faqItemPayload{
			Question:  item.Question,
			Answer:    item.Answer,
			SortOrder: item.SortOrder,
		})
	}
	return payloads
}

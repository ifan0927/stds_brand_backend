package http

import (
	"context"
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ifan0927/stds_brand_backend/internal/application"
)

const propertyAvailabilityTimeout = 2 * time.Second

// PropertyAvailabilityService provides approved property availability to HTTP handlers.
type PropertyAvailabilityService interface {
	// ListPropertyAvailability returns approved property availability.
	ListPropertyAvailability(ctx context.Context) ([]application.PropertyAvailability, error)
}

type propertyAvailabilityResponse struct {
	Items []propertyAvailabilityPayload `json:"items"`
}

type propertyAvailabilityPayload struct {
	PropertyID         string `json:"property_id"`
	PropertyPublicName string `json:"property_public_name"`
	Address            string `json:"address"`
	HasVacantRoom      bool   `json:"has_vacant_room"`
}

func handlePropertyAvailability(service PropertyAvailabilityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if service == nil {
			writeServiceUnavailableError(c)
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), propertyAvailabilityTimeout)
		defer cancel()

		items, err := service.ListPropertyAvailability(ctx)
		if err != nil {
			writeServiceUnavailableError(c)
			return
		}

		c.JSON(nethttp.StatusOK, propertyAvailabilityResponse{Items: propertyAvailabilityPayloadsFromApplication(items)})
	}
}

func propertyAvailabilityPayloadsFromApplication(items []application.PropertyAvailability) []propertyAvailabilityPayload {
	payloads := make([]propertyAvailabilityPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, propertyAvailabilityPayload{
			PropertyID:         item.PropertyID,
			PropertyPublicName: item.PropertyPublicName,
			Address:            item.Address,
			HasVacantRoom:      item.HasVacantRoom,
		})
	}
	return payloads
}

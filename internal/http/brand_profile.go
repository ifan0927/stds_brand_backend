package http

import (
	"context"
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ifan0927/stds_brand_backend/internal/application"
)

const brandProfileTimeout = 2 * time.Second

// BrandProfileService provides the approved brand profile to HTTP handlers.
type BrandProfileService interface {
	// GetBrandProfile returns the approved brand profile.
	GetBrandProfile(ctx context.Context) (*application.BrandProfile, error)
}

type brandProfileResponse struct {
	Profile *brandProfilePayload `json:"profile"`
}

type brandProfilePayload struct {
	BrandName      string  `json:"brand_name"`
	ContactPhone   *string `json:"contact_phone"`
	ContactEmail   *string `json:"contact_email"`
	ContactAddress *string `json:"contact_address"`
	UpdatedAt      string  `json:"updated_at"`
}

func handleBrandProfile(service BrandProfileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if service == nil {
			writeServiceUnavailableError(c)
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), brandProfileTimeout)
		defer cancel()

		profile, err := service.GetBrandProfile(ctx)
		if err != nil {
			writeServiceUnavailableError(c)
			return
		}

		c.JSON(nethttp.StatusOK, brandProfileResponse{Profile: brandProfilePayloadFromApplication(profile)})
	}
}

func brandProfilePayloadFromApplication(profile *application.BrandProfile) *brandProfilePayload {
	if profile == nil {
		return nil
	}

	return &brandProfilePayload{
		BrandName:      profile.BrandName,
		ContactPhone:   profile.ContactPhone,
		ContactEmail:   profile.ContactEmail,
		ContactAddress: profile.ContactAddress,
		UpdatedAt:      profile.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

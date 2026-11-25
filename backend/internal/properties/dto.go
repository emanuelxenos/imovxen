package properties

import (
	"github.com/google/uuid"
)

type CreatePropertyDTO struct {
	Title         string   `json:"title" binding:"required"`
	Description   *string  `json:"description"`
	Type          string   `json:"type" binding:"required,oneof=HOUSE APARTMENT LAND COMMERCIAL"`
	Purpose       string   `json:"purpose" binding:"required,oneof=SALE RENT"`
	Price         float64  `json:"price" binding:"required,gt=0"`
	CondoFee      *float64 `json:"condo_fee"`
	IPTU          *float64 `json:"iptu"`
	AreaM2        *float64 `json:"area_m2"`
	Bedrooms      *int16   `json:"bedrooms"`
	Bathrooms     *int16   `json:"bathrooms"`
	ParkingSpots  *int16   `json:"parking_spots"`
	ZipCode       *string  `json:"zip_code"`
	Street        *string  `json:"street"`
	Number        *string  `json:"number"`
	Complement    *string  `json:"complement"`
	Neighborhood  *string  `json:"neighborhood"`
	City          string   `json:"city" binding:"required"`
	State         string   `json:"state" binding:"required,len=2"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	CoverImageURL *string  `json:"cover_image_url"`
}

type PropertyResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	Type          string    `json:"type"`
	Purpose       string    `json:"purpose"`
	Status        string    `json:"status"`
	Price         float64   `json:"price"`
	CondoFee      *float64  `json:"condo_fee,omitempty"`
	IPTU          *float64  `json:"iptu,omitempty"`
	AreaM2        *float64  `json:"area_m2,omitempty"`
	Bedrooms      *int16    `json:"bedrooms,omitempty"`
	Bathrooms     *int16    `json:"bathrooms,omitempty"`
	ParkingSpots  *int16    `json:"parking_spots,omitempty"`
	ZipCode       *string   `json:"zip_code,omitempty"`
	Street        *string   `json:"street,omitempty"`
	Number        *string   `json:"number,omitempty"`
	Neighborhood  *string   `json:"neighborhood,omitempty"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	Latitude      *float64  `json:"latitude,omitempty"`
	Longitude     *float64  `json:"longitude,omitempty"`
	CoverImageURL *string   `json:"cover_image,omitempty"`
	ViewsCount    int32     `json:"views_count"`
}

type PresignImageDTO struct {
	Filename string `json:"filename" binding:"required"`
	MimeType string `json:"mime_type" binding:"required"`
}

type ConfirmImageDTO struct {
	URL      string `json:"url" binding:"required"`
	Position int16  `json:"position"`
	IsCover  bool   `json:"is_cover"`
}

type CreateLeadDTO struct {
	Name    string  `json:"name" binding:"required"`
	Email   *string `json:"email"`
	Phone   *string `json:"phone"`
	Message *string `json:"message"`
}

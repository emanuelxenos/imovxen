package properties

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type CreatePropertyDTO struct {
	Title         string   `json:"title" binding:"required,min=5,max=150"`
	Description   *string  `json:"description" binding:"omitempty,max=2000"`
	Type          string   `json:"type" binding:"required,oneof=HOUSE APARTMENT LAND COMMERCIAL"`
	Purpose       string   `json:"purpose" binding:"required,oneof=SALE RENT"`
	Price         float64  `json:"price" binding:"required,gt=0"`
	CondoFee      *float64 `json:"condo_fee" binding:"omitempty,gte=0"`
	IPTU          *float64 `json:"iptu" binding:"omitempty,gte=0"`
	AreaM2        *float64 `json:"area_m2" binding:"omitempty,gt=0"`
	Bedrooms      *int16   `json:"bedrooms" binding:"omitempty,gte=0"`
	Bathrooms     *int16   `json:"bathrooms" binding:"omitempty,gte=0"`
	ParkingSpots  *int16   `json:"parking_spots" binding:"omitempty,gte=0"`
	ZipCode       *string  `json:"zip_code" binding:"omitempty,len=8"`
	Street        *string  `json:"street" binding:"omitempty,max=200"`
	Number        *string  `json:"number" binding:"omitempty,max=20"`
	Complement    *string  `json:"complement" binding:"omitempty,max=50"`
	Neighborhood  *string  `json:"neighborhood" binding:"omitempty,max=100"`
	City          string   `json:"city" binding:"required,min=2,max=100"`
	State         string   `json:"state" binding:"required,len=2"`
	Latitude      *float64 `json:"latitude" binding:"omitempty,latitude"`
	Longitude     *float64 `json:"longitude" binding:"omitempty,longitude"`
	CoverImageURL *string  `json:"cover_image_url"`
}

func (dto *CreatePropertyDTO) Sanitize() error {
	dto.Title = strings.TrimSpace(dto.Title)
	dto.City = strings.TrimSpace(dto.City)
	dto.State = strings.ToUpper(strings.TrimSpace(dto.State))
	dto.Type = strings.ToUpper(strings.TrimSpace(dto.Type))
	dto.Purpose = strings.ToUpper(strings.TrimSpace(dto.Purpose))

	if dto.Price <= 0 {
		return errors.New("o preço do imóvel deve ser maior que zero")
	}
	if dto.AreaM2 != nil && *dto.AreaM2 <= 0 {
		return errors.New("a área do imóvel deve ser maior que zero")
	}
	return nil
}

type PresignImageDTO struct {
	Filename string `json:"filename" binding:"required"`
	MimeType string `json:"mime_type" binding:"required"`
}

type ConfirmImageDTO struct {
	URL      string `json:"url" binding:"required,url"`
	Position int16  `json:"position" binding:"gte=0"`
	IsCover  bool   `json:"is_cover"`
}

type CreateLeadDTO struct {
	Name    string  `json:"name" binding:"required,min=2,max=100"`
	Email   *string `json:"email" binding:"omitempty,email"`
	Phone   *string `json:"phone" binding:"omitempty,min=8,max=20"`
	Message *string `json:"message" binding:"omitempty,max=1000"`
}

func (dto *CreateLeadDTO) Sanitize() {
	dto.Name = strings.TrimSpace(dto.Name)
	if dto.Email != nil {
		e := strings.ToLower(strings.TrimSpace(*dto.Email))
		dto.Email = &e
	}
	if dto.Phone != nil {
		p := strings.TrimSpace(*dto.Phone)
		dto.Phone = &p
	}
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
	City          string    `json:"city"`
	State         string    `json:"state"`
	CoverImageURL *string   `json:"cover_image_url,omitempty"`
	DistanceKm    float64   `json:"distance_km,omitempty"`
}

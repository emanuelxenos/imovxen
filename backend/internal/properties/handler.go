package properties

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imovxen/backend/internal/auth"
	"github.com/imovxen/backend/internal/db"
	"github.com/imovxen/backend/internal/storage"
)

type Handler struct {
	db           *db.Database
	tokenService *auth.TokenService
}

func NewHandler(database *db.Database, tokenService *auth.TokenService) *Handler {
	return &Handler{
		db:           database,
		tokenService: tokenService,
	}
}

// UploadPhotos trata upload com compressão automática Lanczos3
func (h *Handler) UploadPhotos(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar formulário de upload"})
		return
	}

	files := form.File["photos"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum arquivo enviado"})
		return
	}

	var fileURLs []string
	for _, file := range files {
		fileName := fmt.Sprintf("%s.jpg", uuid.New().String())
		dstPath := filepath.Join("./uploads", fileName)

		src, err := file.Open()
		if err != nil {
			continue
		}

		err = storage.ProcessAndOptimizeImage(src, dstPath)
		src.Close()

		if err != nil {
			srcFallback, _ := file.Open()
			dstFallback, _ := os.Create(dstPath)
			_, _ = io.Copy(dstFallback, srcFallback)
			srcFallback.Close()
			dstFallback.Close()
		}

		fileURLs = append(fileURLs, "http://localhost:8080/uploads/"+fileName)
	}

	c.JSON(http.StatusOK, gin.H{"urls": fileURLs})
}

// ListProperties lista com filtro geo-espacial e filtros de finalidade/tipo/preço
func (h *Handler) ListProperties(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		q = c.Query("city")
	}
	propType := c.Query("type")
	purpose := c.Query("purpose")
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)
	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
	radius, _ := strconv.ParseFloat(c.Query("radius_km"), 64)

	limit := int32(20)
	offset := int32(0)
	if p, _ := strconv.Atoi(c.Query("page")); p > 1 {
		offset = int32((p - 1) * 20)
	}

	items, err := h.db.Queries.SearchPropertiesGeo(c.Request.Context(), db.SearchPropertiesGeoParams{
		Query:    q,
		Type:     propType,
		Purpose:  purpose,
		MaxPrice: maxPrice,
		Limit:    limit,
		Offset:   offset,
		Lat:      lat,
		Lng:      lng,
		RadiusKm: radius,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao consultar imóveis"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      items,
		"page":      offset/limit + 1,
		"page_size": limit,
		"total":     len(items),
	})
}

// GetProperty busca imóvel por ID
func (h *Handler) GetProperty(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	prop, err := h.db.Queries.GetPropertyByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Imóvel não encontrado"})
		return
	}

	c.JSON(http.StatusOK, prop)
}

// CreateProperty cria imóvel respeitando cota
func (h *Handler) CreateProperty(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)
	user, err := h.db.Queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	if user.ListingsUsed >= user.ListingQuota {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "quota_exceeded",
			"message": "Você atingiu o limite de anúncios ativos do seu plano. Faça upgrade para continuar.",
		})
		return
	}

	var req CreatePropertyDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Sanitize()

	prop, err := h.db.Queries.CreateProperty(c.Request.Context(), db.CreatePropertyParams{
		UserID:        userID,
		Title:         req.Title,
		Description:   req.Description,
		Type:          req.Type,
		Purpose:       req.Purpose,
		Status:        "ACTIVE",
		Price:         req.Price,
		CondoFee:      req.CondoFee,
		Iptu:          req.IPTU,
		AreaM2:        req.AreaM2,
		Bedrooms:      req.Bedrooms,
		Bathrooms:     req.Bathrooms,
		ParkingSpots:  req.ParkingSpots,
		ZipCode:       req.ZipCode,
		Street:        req.Street,
		Number:        req.Number,
		Complement:    req.Complement,
		Neighborhood:  req.Neighborhood,
		City:          req.City,
		State:         req.State,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		CoverImageUrl: req.CoverImageURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar imóvel no banco"})
		return
	}

	c.JSON(http.StatusCreated, prop)
}

// ListMyProperties lista anúncios do usuário autenticado
func (h *Handler) ListMyProperties(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)
	props, err := h.db.Queries.ListPropertiesByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar anúncios"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": props})
}

// UpdateProperty edita anúncio
func (h *Handler) UpdateProperty(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)
	propID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de imóvel inválido"})
		return
	}

	var req CreatePropertyDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Sanitize()

	updated, err := h.db.Queries.UpdateProperty(c.Request.Context(), db.UpdatePropertyParams{
		ID:            propID,
		UserID:        userID,
		Title:         req.Title,
		Description:   req.Description,
		Type:          req.Type,
		Purpose:       req.Purpose,
		Price:         req.Price,
		CondoFee:      req.CondoFee,
		Iptu:          req.IPTU,
		AreaM2:        req.AreaM2,
		Bedrooms:      req.Bedrooms,
		Bathrooms:     req.Bathrooms,
		ParkingSpots:  req.ParkingSpots,
		City:          req.City,
		State:         req.State,
		CoverImageUrl: req.CoverImageURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar anúncio"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteProperty exclui anúncio com decremento automático de cota via trigger
func (h *Handler) DeleteProperty(c *gin.Context) {
	userID, _ := auth.GetContextUserID(c)
	propID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de imóvel inválido"})
		return
	}

	err = h.db.Queries.DeleteProperty(c.Request.Context(), db.DeletePropertyParams{
		ID:     propID,
		UserID: userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir anúncio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Anúncio excluído com sucesso"})
}

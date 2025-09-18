package handlers

import (
	"net/http"
	"strconv"
	"time"

	"sca/sca/internal/clients/thecatapi"
	"sca/sca/internal/models"

	// "sca/sca/internal/validators"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	v      *validator.Validate
	breeds thecatapi.Client
}

func New(db *gorm.DB, opts ...Option) *Handler {
	h := &Handler{db: db, v: validator.New(), breeds: thecatapi.NewHTTP(10 * time.Minute)}
	for _, o := range opts {
		o(h)
	}
	return h
}

type Option func(*Handler)

func WithBreedClient(c thecatapi.Client) Option { return func(h *Handler) { h.breeds = c } }

type createCatReq struct {
	Name              string `json:"name" validate:"required,min=2"`
	YearsOfExperience int    `json:"years_of_experience" validate:"gte=0"`
	Breed             string `json:"breed" validate:"required"`
	SalaryCents       int64  `json:"salary_cents" validate:"gte=0"`
}

// CreateCat godoc
// @Summary Create a spy cat
// @Tags cats
// @Accept json
// @Produce json
// @Param payload body createCatReq true "Cat payload"
// @Success 201 {object} models.Cat
// @Failure 400 {object} map[string]any
// @Router /cats [post]
func (h *Handler) CreateCat(c *gin.Context) {
	var req createCatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.v.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ok, err := h.breeds.ValidateBreed(req.Breed)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "breed validation upstream unavailable"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid breed: " + req.Breed})
		return
	}

	cat := models.Cat{Name: req.Name, YearsOfExperience: req.YearsOfExperience, Breed: req.Breed, SalaryCents: req.SalaryCents}
	if err := h.db.Create(&cat).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// ListCats godoc
// @Summary List spy cats
// @Tags cats
// @Produce json
// @Success 200 {array} models.Cat
// @Failure 500 {object} map[string]any
// @Router /cats [get]
func (h *Handler) ListCats(c *gin.Context) {
	var cats []models.Cat
	if err := h.db.Find(&cats).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cats)
}

// GetCat godoc
// @Summary Get a spy cat by ID
// @Tags cats
// @Produce json
// @Param id path int true "Cat ID"
// @Success 200 {object} models.Cat
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /cats/{id} [get]
func (h *Handler) GetCat(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var cat models.Cat
	if err := h.db.First(&cat, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, cat)
}

// ListBreeds godoc
// @Summary Get all cat breeds from external API
// @Tags cats
// @Produce json
// @Success 200 {array} thecatapi.Breed
// @Failure 502 {object} map[string]any
// @Router /breeds [get]
func (h *Handler) ListBreeds(c *gin.Context) {
	list, err := h.breeds.ListBreeds()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "breed service unavailable"})
		return
	}
	c.JSON(http.StatusOK, list)
}

type updateCatReq struct {
	SalaryCents *int64 `json:"salary_cents" validate:"omitempty,gte=0"`
}

// UpdateCat godoc
// @Summary Update a spy cat
// @Tags cats
// @Accept json
// @Produce json
// @Param id path int true "Cat ID"
// @Param payload body updateCatReq true "Update cat payload"
// @Success 200 {object} models.Cat
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /cats/{id} [put]
func (h *Handler) UpdateCat(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req updateCatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.v.Struct(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]any{}
	if req.SalaryCents != nil {
		updates["salary_cents"] = *req.SalaryCents
	}
	if len(updates) == 0 {
		c.JSON(400, gin.H{"error": "no fields"})
		return
	}
	if err := h.db.Model(&models.Cat{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	var cat models.Cat
	h.db.First(&cat, id)
	c.JSON(200, cat)
}

// DeleteCat godoc
// @Summary Delete a spy cat
// @Tags cats
// @Produce json
// @Param id path int true "Cat ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /cats/{id} [delete]
func (h *Handler) DeleteCat(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	res := h.db.Delete(&models.Cat{}, id)
	if res.Error != nil {
		c.JSON(500, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.Status(204)
}

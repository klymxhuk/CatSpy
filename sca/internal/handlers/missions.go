package handlers

import (
	"strconv"

	"sca/sca/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type createMissionReq struct {
	AssignedCatID *uint           `json:"assigned_cat_id"`
	Completed     *bool           `json:"completed"`
	Targets       []targetPayload `json:"targets" validate:"required,min=1,max=3,dive"`
}
type targetPayload struct {
	Name      string `json:"name" validate:"required,min=2"`
	Country   string `json:"country" validate:"required"`
	Notes     string `json:"notes"`
	Completed bool   `json:"completed"`
}

// @Summary Create mission with targets
// @Tags missions
// @Accept json
// @Produce json
// @Param payload body createMissionReq true "Mission payload"
// @Success 201 {object} models.Mission
// @Router /missions [post]
func (h *Handler) CreateMission(c *gin.Context) {
	var req createMissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.v.Struct(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// build mission
	completed := false
	if req.Completed != nil {
		completed = *req.Completed
	}
	m := models.Mission{Completed: completed}
	if req.AssignedCatID != nil {
		m.AssignedCatID = req.AssignedCatID
	}

	// ensure unique target names within payload and build targets
	seen := map[string]struct{}{}
	for _, t := range req.Targets {
		if _, ok := seen[t.Name]; ok {
			c.JSON(400, gin.H{"error": "duplicate target name in request: " + t.Name})
			return
		}
		seen[t.Name] = struct{}{}
		m.Targets = append(m.Targets, models.Target{Name: t.Name, Country: t.Country, Notes: t.Notes, Completed: t.Completed})
	}

	if err := h.db.Create(&m).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, m)
}

// ListMissions godoc
// @Summary List missions with targets
// @Tags missions
// @Produce json
// @Success 200 {array} models.Mission
// @Failure 500 {object} map[string]any
// @Router /missions [get]
func (h *Handler) ListMissions(c *gin.Context) {
	var m []models.Mission
	if err := h.db.Preload("Targets").Find(&m).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, m)
}

// GetMission godoc
// @Summary Get a mission by ID with targets
// @Tags missions
// @Produce json
// @Param id path int true "Mission ID"
// @Success 200 {object} models.Mission
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /missions/{id} [get]
func (h *Handler) GetMission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	var m models.Mission
	if err := h.db.Preload("Targets").First(&m, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, m)
}

type updateMissionReq struct {
	Completed *bool `json:"completed"`
}

// UpdateMission godoc
// @Summary Update a mission (mark as completed)
// @Tags missions
// @Accept json
// @Produce json
// @Param id path int true "Mission ID"
// @Param payload body updateMissionReq true "Update mission payload"
// @Success 200 {object} models.Mission
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /missions/{id} [patch]
func (h *Handler) UpdateMission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	var m models.Mission
	if err := h.db.Preload("Targets").First(&m, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if m.Completed {
		c.JSON(400, gin.H{"error": "mission already completed"})
		return
	}
	var req updateMissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Completed != nil && *req.Completed {
		m.Completed = true
		if err := h.db.Save(&m).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(200, m)
}

// DeleteMission godoc
// @Summary Delete a mission
// @Tags missions
// @Produce json
// @Param id path int true "Mission ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /missions/{id} [delete]
func (h *Handler) DeleteMission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	var m models.Mission
	if err := h.db.First(&m, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if m.AssignedCatID != nil {
		c.JSON(400, gin.H{"error": "cannot delete: assigned to a cat"})
		return
	}
	if err := h.db.Delete(&m).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Status(204)
}

type assignCatReq struct {
	CatID uint `json:"cat_id" validate:"required"`
}

// AssignCat godoc
// @Summary Assign a cat to a mission
// @Tags missions
// @Accept json
// @Produce json
// @Param id path int true "Mission ID"
// @Param payload body assignCatReq true "Cat assignment payload"
// @Success 200 {object} models.Mission
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /missions/{id}/assign_cat [post]
func (h *Handler) AssignCat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	var m models.Mission
	if err := h.db.Preload("Targets").First(&m, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if m.Completed {
		c.JSON(400, gin.H{"error": "mission completed"})
		return
	}

	var req assignCatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.v.Struct(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	m.AssignedCatID = &req.CatID
	res := h.db.Model(&models.Mission{}).
		Where("id = ? AND completed = false AND assigned_cat_id IS NULL", id).
		Update("assigned_cat_id", req.CatID)
	if res.Error != nil {
		c.JSON(500, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(409, gin.H{"error": "cat already assigned to an active mission or mission not editable"})
		return
	}

	if err := h.db.Preload("Targets").First(&m, id).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, m)
}

type addTargetsReq struct {
	Targets []targetPayload `json:"targets" validate:"required,min=1,max=3,dive"`
}

// AddTargets godoc
// @Summary Add new targets to a mission
// @Tags missions
// @Accept json
// @Produce json
// @Param id path int true "Mission ID"
// @Param payload body addTargetsReq true "Targets payload (1–3 targets)"
// @Success 200 {object} models.Mission
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /missions/{id}/targets [post]
func (h *Handler) AddTargets(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	var m models.Mission
	if err := h.db.Preload("Targets").First(&m, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if m.Completed {
		c.JSON(400, gin.H{"error": "mission completed"})
		return
	}

	var req addTargetsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.v.Struct(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if len(m.Targets)+len(req.Targets) > 3 {
		c.JSON(400, gin.H{"error": "exceeds max 3 targets"})
		return
	}

	existing := map[string]struct{}{}
	for _, et := range m.Targets {
		existing[et.Name] = struct{}{}
	}
	reqSeen := map[string]struct{}{}
	for _, t := range req.Targets {
		if verr := validator.New().Var(t.Name, "required,min=2"); verr != nil {
			c.JSON(400, gin.H{"error": "invalid target"})
			return
		}
		if _, ok := existing[t.Name]; ok {
			c.JSON(400, gin.H{"error": "target with this name already exists in mission: " + t.Name})
			return
		}
		if _, ok := reqSeen[t.Name]; ok {
			c.JSON(400, gin.H{"error": "duplicate target name in request: " + t.Name})
			return
		}
		reqSeen[t.Name] = struct{}{}
		m.Targets = append(m.Targets, models.Target{Name: t.Name, Country: t.Country, Notes: t.Notes, Completed: t.Completed})
	}
	if err := h.db.Save(&m).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, m)
}

type updateTargetReq struct {
	Notes     *string `json:"notes"`
	Completed *bool   `json:"completed"`
}

// UpdateTarget godoc
// @Summary Update a target in a mission
// @Tags missions
// @Accept json
// @Produce json
// @Param id path int true "Mission ID"
// @Param tid path int true "Target ID"
// @Param payload body updateTargetReq true "Update target payload"
// @Success 200 {object} models.Target
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /missions/{id}/targets/{tid} [patch]
func (h *Handler) UpdateTarget(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	var m models.Mission
	if err := h.db.First(&m, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "mission not found"})
		return
	}
	if m.Completed {
		c.JSON(400, gin.H{"error": "mission completed"})
		return
	}

	tid, _ := strconv.Atoi(c.Param("tid"))
	var t models.Target
	if err := h.db.First(&t, tid).Error; err != nil {
		c.JSON(404, gin.H{"error": "target not found"})
		return
	}
	if t.MissionID != uint(id) {
		c.JSON(400, gin.H{"error": "target not in mission"})
		return
	}
	if t.Completed {
		c.JSON(400, gin.H{"error": "target completed; notes frozen"})
		return
	}

	var req updateTargetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Completed != nil && *req.Completed && req.Notes != nil {
		c.JSON(400, gin.H{"error": "cannot update notes when marking target completed"})
		return
	}

	if t.Completed {
		c.JSON(400, gin.H{"error": "target completed; notes frozen"})
		return
	}

	if req.Completed != nil && *req.Completed {
		t.Completed = true
	}
	if req.Notes != nil {
		t.Notes = *req.Notes
	}
	if err := h.db.Save(&t).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, t)
}

// DeleteTarget godoc
// @Summary Delete a target from a mission
// @Tags missions
// @Produce json
// @Param id path int true "Mission ID"
// @Param tid path int true "Target ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /missions/{id}/targets/{tid} [delete]
func (h *Handler) DeleteTarget(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	tid, _ := strconv.Atoi(c.Param("tid"))
	var t models.Target
	if err := h.db.First(&t, tid).Error; err != nil {
		c.JSON(404, gin.H{"error": "target not found"})
		return
	}
	if t.MissionID != uint(id) {
		c.JSON(400, gin.H{"error": "target not in mission"})
		return
	}
	if t.Completed {
		c.JSON(400, gin.H{"error": "cannot delete completed target"})
		return
	}
	if err := h.db.Delete(&t).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Status(204)
}

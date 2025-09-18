package server

import (
	"fmt"
	"os"
	"time"

	"sca/sca/internal/handlers"
	"sca/sca/internal/storage"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// swagger docs
	_ "sca/docs" // swagger docs

	"gorm.io/gorm"
)

func Router(db *gorm.DB) *gin.Engine {
	storage.MustRunMigrations(db)

	r := gin.New()
	if os.Getenv("APP_ENV") == "dev" {
		r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("{\"time\":\"%s\",\"status\":%d,\"method\":\"%s\",\"path\":\"%s\",\"latency\":\"%s\"}\n",
				param.TimeStamp.Format(time.RFC3339),
				param.StatusCode,
				param.Method,
				param.Path,
				param.Latency,
			)
		}))
		r.Use(ResponseLogger())
	}
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	v1 := r.Group("/api/v1")
	{
		h := handlers.New(db)

		// Cats
		v1.POST("/cats", h.CreateCat)
		v1.GET("/cats", h.ListCats)
		v1.GET("/cats/:id", h.GetCat)
		v1.PATCH("/cats/:id", h.UpdateCat) // salary only
		v1.DELETE("/cats/:id", h.DeleteCat)
		v1.GET("/breeds", h.ListBreeds)

		// Missions
		v1.POST("/missions", h.CreateMission)
		v1.GET("/missions", h.ListMissions)
		v1.GET("/missions/:id", h.GetMission)
		v1.PATCH("/missions/:id", h.UpdateMission)
		v1.DELETE("/missions/:id", h.DeleteMission)
		v1.POST("/missions/:id/assign_cat", h.AssignCat)

		// Targets
		v1.POST("/missions/:id/targets", h.AddTargets)
		v1.PATCH("/missions/:id/targets/:tid", h.UpdateTarget)
		v1.DELETE("/missions/:id/targets/:tid", h.DeleteTarget)
	}

	// Swagger
	// import side-effects in cmd to register docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

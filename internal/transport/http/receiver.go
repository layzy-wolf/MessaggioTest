package http

import (
	"MessagioTest/internal/service/receiver"
	"MessagioTest/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupApi(db *gorm.DB, messages chan<- *models.Message) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.POST("/message", receiver.SetupAPIReceiver(db, messages))
	r.GET("/statistic", receiver.GetStatistic(db))

	return r
}

package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/wso2/identity-customer-data-service/pkg/errors"
	"github.com/wso2/identity-customer-data-service/pkg/models"
	"github.com/wso2/identity-customer-data-service/pkg/service"
	"github.com/wso2/identity-customer-data-service/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

func AddEventSchema(c *gin.Context) {
	var schema models.EventSchema
	if err := c.ShouldBindJSON(&schema); err != nil {
		badReq := errors.NewClientError(errors.ErrorMessage{
			Code:        errors.ErrBadRequest.Code,
			Message:     errors.ErrBadRequest.Message,
			Description: err.Error(),
		}, http.StatusBadRequest)

		utils.HandleError(c, badReq)
		return
	}
	err := service.AddEventSchema(schema)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, schema)
}

func GetEventSchemas(c *gin.Context) {

	schemas, err := service.GetEventSchemas()
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, schemas)
}

func GetEventSchema(c *gin.Context) {
	id := c.Param("event_schema_id")
	schema, err := service.GetEventSchema(id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, schema)
}

func PatchEventSchema(c *gin.Context) {
	id := c.Param("event_schema_id")
	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		badReq := errors.NewClientError(errors.ErrorMessage{
			Code:        errors.ErrBadRequest.Code,
			Message:     errors.ErrBadRequest.Message,
			Description: err.Error(),
		}, http.StatusBadRequest)

		utils.HandleError(c, badReq)
		return
	}
	if err := service.PatchEventSchema(id, updates); err != nil {
		utils.HandleError(c, err)
		return
	}
	eventSchema, _ := service.GetEventSchema(id)
	c.JSON(http.StatusOK, eventSchema)
}

func DeleteEventSchema(c *gin.Context) {
	id := c.Param("event_schema_id")
	if err := service.DeleteEventSchema(id); err != nil {
		utils.HandleError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, "Event schema deleted successfully")
}

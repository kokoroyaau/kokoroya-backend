package response

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func OK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"success": true, "data": data})
}

func Err(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "error": message})
}

func NoContent(c *gin.Context) {
	c.Status(204)
}

func AbortErr(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"success": false, "error": message})
}

var duplicateFieldByConstraint = map[string]string{
	"users_email_key":  "email",
	"users_pin_unique": "PIN",
}

func DBErr(c *gin.Context, err error) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		field, ok := duplicateFieldByConstraint[pqErr.Constraint]
		if !ok {
			field = "value"
		}
		Err(c, 409, field+" already in use")
		return
	}
	Err(c, 500, "internal server error")
}

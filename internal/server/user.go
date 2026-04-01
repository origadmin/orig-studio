package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"origadmin/application/origcms/internal/data/entity"
)

func RegisterUserRoutes(group *gin.RouterGroup, client *entity.Client) {
	users := group.Group("/users")
	{
		// List users
		users.GET("", func(c *gin.Context) {
			limit := 20
			offset := 0
			if l := c.Query("limit"); l != "" {
				if n, err := strconv.Atoi(l); err == nil {
					limit = n
				}
			}
			if o := c.Query("offset"); o != "" {
				if n, err := strconv.Atoi(o); err == nil {
					offset = n
				}
			}

			users, err := client.User.Query().Limit(limit).Offset(offset).All(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"list":  users,
				"total": len(users),
			})
		})

		// Get user by ID
		users.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			u, err := client.User.Get(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusOK, u)
		})

		// Create user
		users.POST("", func(c *gin.Context) {
			var input struct {
				Username string `json:"username"`
				Email    string `json:"email"`
				Password string `json:"password"`
				Name     string `json:"name"`
			}
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			u, err := client.User.Create().
				SetUsername(input.Username).
				SetEmail(input.Email).
				SetName(input.Name).
				SetPassword(input.Password).
				Save(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, u)
		})

		// Delete user
		users.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			err = client.User.DeleteOneID(id).Exec(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})
	}
}

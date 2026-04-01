package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"origadmin/application/origcms/internal/data/entity"
)

// UploadDir is the directory where uploaded media files are stored.
const UploadDir = "data/uploads"

func RegisterMediaRoutes(group *gin.RouterGroup, client *entity.Client) {
	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create upload directory: %v\n", err)
	}

	media := group.Group("/media")
	{
		// List media
		media.GET("", func(c *gin.Context) {
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

			query := client.Media.Query()

			// Count total
			total, _ := query.Count(c.Request.Context())

			// Get page
			offset := (page - 1) * pageSize
			items, err := query.Limit(pageSize).Offset(offset).All(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"list":      items,
				"total":     total,
				"page":      page,
				"page_size": pageSize,
			})
		})

		// Get media by ID
		media.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			m, err := client.Media.Get(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
				return
			}

			// Increment views (simple add)
			client.Media.UpdateOneID(id).
				AddViewCount(1).
				Exec(c.Request.Context())

			c.JSON(http.StatusOK, m)
		})

		// Upload media file
		media.POST("/upload", func(c *gin.Context) {
			file, header, err := c.Request.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request"})
				return
			}
			defer file.Close()

			title := c.PostForm("title")
			description := c.PostForm("description")
			userIDStr := c.PostForm("user_id")
			userID, _ := strconv.Atoi(userIDStr)

			if title == "" {
				title = header.Filename
			}

			// Generate unique filename
			ext := filepath.Ext(header.Filename)
			newFilename := uuid.New().String() + ext
			filePath := filepath.Join(UploadDir, newFilename)

			// Save file to disk
			out, err := os.Create(filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file to disk"})
				return
			}
			defer out.Close()

			if _, err := io.Copy(out, file); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file content"})
				return
			}

			// Create database record
			// We store the relative URL for frontend access
			fileURL := "/uploads/" + newFilename

			m, err := client.Media.Create().
				SetTitle(title).
				SetDescription(description).
				SetType("video"). // Default to video for now
				SetURL(fileURL).
				SetUserID(userID).
				SetState("active").
				SetEncodingStatus("pending").
				Save(c.Request.Context())
			if err != nil {
				// Clean up file if DB save fails
				_ = os.Remove(filePath)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save media record: " + err.Error()})
				return
			}

			c.JSON(http.StatusCreated, m)
		})

		// Delete media
		media.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			// Get media first to find file path
			m, err := client.Media.Get(c.Request.Context(), id)
			if err == nil {
				// Attempt to delete file
				filename := filepath.Base(m.URL)
				_ = os.Remove(filepath.Join(UploadDir, filename))
			}

			err = client.Media.DeleteOneID(id).Exec(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})
	}
}

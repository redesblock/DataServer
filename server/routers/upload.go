package routers

//import (
//	"github.com/gin-gonic/gin"
//	"github.com/redesblock/gateway/dataservice"
//	"net/http"
//	"path/filepath"
//)
//
//// @Summary upload
//// @Description
//// @Security ApiKeyAuth
//// @Tags file
//// @Accept multipart/form-data
//// @Param files formData file true "files"
//// @Produce  json
//// @Success 200 string token
//// @Router /upload [post]
//func UploadHandler(db *dataservice.DataService) func(c *gin.Context) {
//	return func(c *gin.Context) {
//		form, err := c.MultipartForm()
//		if err != nil {
//			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
//			return
//		}
//		files := form.File["files"]
//
//		for _, file := range files {
//			filename := filepath.Base(file.Filename)
//			if err := c.SaveUploadedFile(file, filename); err != nil {
//				c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
//				return
//			}
//		}
//
//		c.String(http.StatusOK, "Uploaded successfully %d files.", len(files))
//	}
//}

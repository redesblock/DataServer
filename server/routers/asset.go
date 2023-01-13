package routers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redesblock/dataserver/dataservice"
	log "github.com/sirupsen/logrus"
)

// @Summary asset
// @Schemes
// @Description asset
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Param   id     path    int     true        "bucket id"
// @Param   fid     query    int     false     "folder id"
// @Success 200 string {}
// @Router /asset/{id} [get]
func GetAssetHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid id"))
			return
		}

		fid, err := strconv.ParseUint(c.DefaultQuery("fid", "0"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, "invalid fid"))
			return
		}

		userID, _ := c.Get("id")
		var user dataservice.User
		if err := db.Find(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		//if user.UsedStorage >= user.TotalStorage {
		//	c.JSON(http.StatusOK, NewResponse(ExecuteCode, "storage usage has reached the maximum"))
		//	return
		//}

		os.Mkdir("assets", os.ModePerm)

		var assetID string
		exists := true
		for exists {
			assetID = uuid.New().String()
			_, err := os.Stat(filepath.Join("assets", assetID))
			if err == nil {
				continue
			}
			if os.IsNotExist(err) {
				exists = false
			}
		}

		if err := os.Mkdir(filepath.Join("assets", assetID), os.ModePerm); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err))
			return
		}

		if err := db.Save(&dataservice.BucketObject{
			Name:     c.Query("name"),
			BucketID: uint(id),
			ParentID: uint(fid),
			AssetID:  assetID,
			Status:   dataservice.STATUS_WAIT,
		}).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		data := map[string]interface{}{
			"asset_id": assetID,
			"url":      "/upload/" + assetID,
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, data))
	}
}

// @Summary asset finish
// @Schemes
// @Description asset finish
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Success 200 string {}
// @Router /finish/{asset_id} [post]
func FinishFileUploadHandler(db *dataservice.DataService, uploadChan chan<- string) func(c *gin.Context) {
	return func(c *gin.Context) {
		assetID := c.Param("asset_id")
		var item *dataservice.BucketObject
		if ret := db.Find(&item, "asset_id = ?", assetID); ret.RowsAffected == 0 {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "asset not found"))
			return
		}
		item.Status = dataservice.STATUS_UPLOADED
		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		readLine := func(fileName string, handler func(string) error) error {
			f, err := os.Open(fileName)
			if err != nil {
				return err
			}
			defer f.Close()

			br := bufio.NewReader(f)
			for {
				line, _, err := br.ReadLine()
				if err != nil {
					// file read complete
					if err == io.EOF {
						return nil
					}
					return err
				}
				if err := handler(string(line)); err != nil {
					return err
				}
			}
		}

		handler := func(line string) error {
			var ret map[string]interface{}
			if err := json.Unmarshal([]byte(line), &ret); err != nil {
				return err
			}
			return nil
		}

		tempFolder := "./assets/" + assetID + "/metadata.json"
		if err := readLine(tempFolder, handler); err != nil {
			fmt.Printf("======= error %s\n", err)
		} else {
			uploadChan <- assetID
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, ""))
	}
}

func GetFileUploadHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		assetID := c.Param("asset_id")
		tempFolder := "./assets/" + assetID
		resumableIdentifier := c.Request.URL.Query()["resumableIdentifier"]
		resumableChunkNumber := c.Request.URL.Query()["resumableChunkNumber"]
		path := fmt.Sprintf("%s/%s", tempFolder, resumableIdentifier[0])
		relativeChunk := fmt.Sprintf("%s%s%s%s", path, "/", "part", resumableChunkNumber[0])
		if _, err := os.Stat(relativeChunk); os.IsNotExist(err) {
			c.JSON(http.StatusMethodNotAllowed, http.StatusText(http.StatusNotFound))
			return
		} else {
			c.JSON(http.StatusCreated, "Chunk already exist")
			return
		}
	}
}

func FileUploadHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")

		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "get form err: %s", err.Error())
			return
		}
		defer file.Close()

		assetID := c.Param("asset_id")
		var item *dataservice.BucketObject
		if ret := db.Find(&item, "asset_id = ?", assetID); ret.RowsAffected == 0 {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "asset not found"))
			return
		}
		tempFolder := "./assets/" + assetID
		resumableIdentifier := c.Request.URL.Query()["resumableIdentifier"]
		resumableChunkNumber := c.Request.URL.Query()["resumableChunkNumber"]
		resumableTotalChunks := c.Request.URL.Query()["resumableTotalChunks"]
		resumableRelativePath := c.Request.URL.Query()["resumableRelativePath"]
		chunkSizeInBytesStr := c.Request.URL.Query()["resumableChunkSize"]
		chunkSizeInBytes, _ := strconv.Atoi(chunkSizeInBytesStr[0])
		path := fmt.Sprintf("%s/%s", tempFolder, resumableIdentifier[0])
		relativeChunk := fmt.Sprintf("%s%s%s%s", path, "/", "part", resumableChunkNumber[0])
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, os.ModePerm)
		}
		f, err := os.OpenFile(relativeChunk, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Errorf("open file %s error ", relativeChunk, err)
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "open file error"))
			return
		}
		defer f.Close()
		if _, err := io.Copy(f, file); err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "copy file error"))
			return
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if resumableChunkNumber[0] == "1" {
				f, err := os.OpenFile(fmt.Sprintf("%s%s%s", tempFolder, "/", "metadata.json"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
				if err != nil {
					return fmt.Errorf("open file metadata.json error %s", err)
				}
				defer f.Close()
				if _, err := f.WriteString(fmt.Sprintf(`{"identifier":"%s", "path":"%s", "chunks":%s}`, resumableIdentifier[0], resumableRelativePath[0], resumableTotalChunks[0]) + "\r\n"); err != nil {
					return fmt.Errorf("write file metadata.json error %s", err)
				}
			}
			var item *dataservice.BucketObject
			if ret := tx.Find(&item, "asset_id = ?", assetID); ret.RowsAffected == 0 {
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, fmt.Errorf("asset %s not found", assetID)))
				return nil
			}
			item.Size += uint64(chunkSizeInBytes)
			item.Name = strings.Split(resumableRelativePath[0], "/")[0]
			item.Status = dataservice.STATUS_UPLOAD
			if err := tx.Save(item).Error; err != nil {
				return err
			}

			time := time.Now().Format("2006-01-02")
			var item2 *dataservice.UsedStorage
			if ret := tx.Model(&dataservice.UsedStorage{}).Where("user_id = ?", userID).Where("time = ?", time).Find(&item2); ret.Error != nil {
				return ret.Error
			} else if ret.RowsAffected == 0 {
				item2 = &dataservice.UsedStorage{
					Time:   time,
					UserID: userID.(uint),
				}
			}
			item2.Num += uint64(chunkSizeInBytes)
			return tx.Save(&item2).Error
		}); err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

// @Summary asset
// @Schemes
// @Description asset
// @Security ApiKeyAuth
// @Tags bucket object
// @Accept json
// @Produce json
// @Param   cid     path    int     true        "cid"
// @Param   path     query  string  false     "path"
// @Success 200 string {}
// @Router /download/{cid}/{path} [get]
func GetFileDownloadHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, _ := c.Get("id")
		var user dataservice.User
		if err := db.Find(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		//if user.UsedTraffic >= user.TotalTraffic {
		//	c.JSON(http.StatusOK, NewResponse(ExecuteCode, "storage usage has reached the maximum"))
		//	return
		//}

		cid := c.Param("cid")
		path := c.Param("path")
		u, _ := url.Parse(viper.GetString("gateway"))
		proxy := httputil.NewSingleHostReverseProxy(u)
		c.Request.URL.Path = "mop/" + cid + "/" + path
		proxy.ModifyResponse = func(response *http.Response) error {
			size, _ := strconv.ParseUint(response.Header.Get("Decompressed-Content-Length"), 10, 64)

			db.Transaction(func(tx *gorm.DB) error {
				time := time.Now().Format("2006-01-02")
				var item2 *dataservice.UsedTraffic
				if ret := db.Model(&dataservice.UsedTraffic{}).Where("user_id = ?", userID).Where("time = ?", time).Find(&item2); ret.Error != nil {
					return ret.Error
				} else if ret.RowsAffected == 0 {
					item2 = &dataservice.UsedTraffic{
						Time:   time,
						UserID: userID.(uint),
					}
				}
				item2.Num += size
				return db.Save(item2).Error
			})

			return nil
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

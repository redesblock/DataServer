package routers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
		os.Mkdir("assets", os.ModePerm)
		uuid := uuid.New()
		if err := os.Mkdir(filepath.Join("assets", uuid.String()), os.ModePerm); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err))
			return
		}

		if err := db.Save(&dataservice.BucketObject{
			Name:     c.Query("name"),
			BucketID: uint(id),
			ParentID: uint(fid),
			AssetID:  uuid.String(),
			Status:   dataservice.STATUS_WAIT,
		}).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}
		data := map[string]interface{}{
			"asset_id": uuid.String(),
			"url":      "/upload/" + uuid.String(),
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
		uploadChan <- assetID
	}
}

func GetFileUploadHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		assetID := c.Param("asset_id")
		tempFolder := "./assets/" + assetID
		resumableIdentifier, _ := c.Request.URL.Query()["resumableIdentifier"]
		resumableChunkNumber, _ := c.Request.URL.Query()["resumableChunkNumber"]
		path := fmt.Sprintf("%s/%s", tempFolder, resumableIdentifier[0])
		relativeChunk := fmt.Sprintf("%s%s%s%s", path, "/", "part", resumableChunkNumber[0])
		if _, err := os.Stat(relativeChunk); os.IsNotExist(err) {
			c.JSON(http.StatusMethodNotAllowed, http.StatusText(http.StatusNotFound))
			return
		} else {
			c.JSON(http.StatusCreated, "Chunk already exist")
			return
		}
		c.JSON(http.StatusOK, "ok")
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
		resumableIdentifier, _ := c.Request.URL.Query()["resumableIdentifier"]
		resumableChunkNumber, _ := c.Request.URL.Query()["resumableChunkNumber"]
		resumableTotalChunks, _ := c.Request.URL.Query()["resumableTotalChunks"]
		resumableRelativePath, _ := c.Request.URL.Query()["resumableRelativePath"]
		resumableFilename, _ := c.Request.URL.Query()["resumableFilename"]
		path := fmt.Sprintf("%s/%s", tempFolder, resumableIdentifier[0])
		relativeChunk := fmt.Sprintf("%s%s%s%s", path, "/", "part", resumableChunkNumber[0])
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, os.ModePerm)
		}
		f, err := os.OpenFile(relativeChunk, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Errorf("open file error ", err)
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, "open file error"))
			return
		}
		defer f.Close()
		io.Copy(f, file)

		uploaded := false
		var videoFileSize int64
		currentChunk, err := strconv.Atoi(resumableChunkNumber[0])
		totalChunks, err := strconv.Atoi(resumableTotalChunks[0])
		// If it is the last chunk, trigger the recombination of chunks
		if currentChunk == totalChunks {
			// log.Info("Combining chunks into one file")
			resumableTotalSize, _ := c.Request.URL.Query()["resumableTotalSize"]
			videoFileSizeInt, _ := strconv.Atoi(resumableTotalSize[0])

			videoFileSize = int64(videoFileSizeInt)
			chunkSizeInBytesStr, _ := c.Request.URL.Query()["resumableChunkSize"]
			chunkSizeInBytes, _ := strconv.Atoi(chunkSizeInBytesStr[0])

			chunksDir := path
			// Generate an empty file
			os.MkdirAll("./assets/"+assetID+"/"+strings.TrimRight(resumableRelativePath[0], resumableFilename[0]), os.ModePerm)
			f, err := os.Create("./assets/" + assetID + "/" + resumableRelativePath[0])
			if err != nil {
				log.Errorf("create file error ", err)
				c.JSON(http.StatusOK, NewResponse(ExecuteCode, "create file error"))
				return
			}
			defer f.Close()

			// For every chunk, write it to the empty file.
			for i := 1; i <= totalChunks; i++ {
				relativePath := fmt.Sprintf("%s%s%d", chunksDir, "/part", i)

				writeOffset := int64(chunkSizeInBytes * (i - 1))
				if i == 1 {
					writeOffset = 0
				}
				dat, err := ioutil.ReadFile(relativePath)
				size, err := f.WriteAt(dat, writeOffset)
				if err != nil {
					log.Errorf("write file error ", err)
					c.JSON(http.StatusOK, NewResponse(ExecuteCode, "write file error"))
					return
				}
				_ = size
				//log.Infof("%d bytes written offset %d\n", size, writeOffset)
			}

			uploaded = true
			if _, err := exec.Command("rm", "-rf", tempFolder+"/"+resumableIdentifier[0]).Output(); err != nil {
				log.Error(tempFolder+"/"+resumableIdentifier[0], err)
			}
		}

		item.Name = strings.Split(resumableRelativePath[0], "/")[0]
		item.Size += uint64(videoFileSize)
		if uploaded {
			// TODO
			item.Status = dataservice.STATUS_UPLOADED
		} else {
			item.Status = dataservice.STATUS_UPLOAD
		}
		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		if uploaded {
			time := time.Now().Format("2006-01-02")
			var item2 *dataservice.UsedStorage
			if ret := db.Model(&dataservice.UsedStorage{}).Where("user_id = ?", userID).Where("time = ?", time).Find(&item2); ret.Error != nil {
				// c.JSON(http.StatusOK, NewResponse(ExecuteCode, ret.Error))
				return
			} else if ret.RowsAffected == 0 {
				item2 = &dataservice.UsedStorage{
					Time:   time,
					UserID: userID.(uint),
				}
			}
			item2.Num += uint64(videoFileSize)
			db.Save(&item2)
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
func GetFileDownloadHandler(db *dataservice.DataService, nodeFunc func() string) func(c *gin.Context) {
	return func(c *gin.Context) {
		cid := c.Param("cid")
		path := c.Param("path")

		userID, _ := c.Get("id")

		u, _ := url.Parse(nodeFunc())
		proxy := httputil.NewSingleHostReverseProxy(u)
		c.Request.URL.Path = "hop/" + cid + "/" + path
		fmt.Println(c.Request.URL.Path)
		proxy.ModifyResponse = func(response *http.Response) error {
			time := time.Now().Format("2006-01-02")
			var item2 *dataservice.UsedTraffic
			if ret := db.Model(&dataservice.UsedStorage{}).Where("user_id = ?", userID).Where("time = ?", time).Find(&item2); ret.Error != nil {
				return ret.Error
			} else if ret.RowsAffected == 0 {
				item2 = &dataservice.UsedTraffic{
					Time:   time,
					UserID: userID.(uint),
				}
			}
			size, _ := strconv.ParseUint(response.Header.Get("Decompressed-Content-Length"), 10, 64)
			item2.Num += size
			db.Save(item2)
			return nil
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

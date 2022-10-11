package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"net/http"
)

// @Summary list storage bills
// @Schemes
// @Description pagination query storage bills
// @Security ApiKeyAuth
// @Tags bills
// @Accept json
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} dataservice.BillStorage
// @Router /bills/storage [get]
func GetBillsStorageHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize

		userID, _ := c.Get("id")
		total, items, err := db.FindBillsStorage(userID.(uint), offset, pageSize)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		pageTotal := total / pageSize
		if total%pageSize != 0 {
			pageTotal++
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, &List{
			Total:     total,
			PageTotal: pageTotal,
			Items:     items,
		}))
	}
}

// @Summary list traffic bills
// @Schemes
// @Description pagination query traffic bills
// @Tags bills
// @Accept json
// @Produce json
// @Param   page_num     query    int     false        "page number"
// @Param   page_size    query    int     false        "page size"
// @Success 200 {object} dataservice.BillTraffic
// @Router /bills/traffic [get]
func GetBillsTrafficHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		pageNum, pageSize := page(c)
		offset := (pageNum - 1) * pageSize

		userID, _ := c.Get("id")
		total, items, err := db.FindBillsTraffic(userID.(uint), offset, pageSize)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		pageTotal := total / pageSize
		if total%pageSize != 0 {
			pageTotal++
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, &List{
			Total:     total,
			PageTotal: pageTotal,
			Items:     items,
		}))
	}
}

type BillReq struct {
	Hash        string `json:"hash"`
	Amount      string `json:"amount"`
	Size        uint64 `json:"size"`
	Description string `json:"description"`
}

// @Summary add storage bill
// @Schemes
// @Description add storage bill
// @Security ApiKeyAuth
// @Tags bills
// @Accept json
// @Produce json
// @Success 200 {object} dataservice.BillStorage
// @Router /bills/storage [post]
func AddBillsStorageHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req BillReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		userID, _ := c.Get("id")
		item := &dataservice.BillStorage{
			Hash:        req.Hash,
			Amount:      req.Amount,
			Size:        req.Size * 1024 * 1024,
			Description: req.Description,
			UserID:      userID.(uint),
		}
		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

// @Summary add traffic bill
// @Schemes
// @Description add traffic bill
// @Tags bills
// @Accept json
// @Produce json
// @Success 200 {object} dataservice.BillTraffic
// @Router /bills/traffic [post]
func AddBillsTrafficHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req BillReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, NewResponse(RequestCode, err.Error()))
			return
		}

		userID, _ := c.Get("id")
		item := &dataservice.BillTraffic{
			Hash:        req.Hash,
			Amount:      req.Amount,
			Size:        req.Size * 1024 * 1024,
			Description: req.Description,
			UserID:      userID.(uint),
		}
		if err := db.Save(item).Error; err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, item))
	}
}

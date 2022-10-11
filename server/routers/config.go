package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/redesblock/dataserver/dataservice"
	"github.com/shopspring/decimal"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func GetAreasHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		val, ok := os.LookupEnv("DATA_SERVER_AREA")
		if !ok {
			val = "China"
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, strings.Split(val, ",")))
	}
}

func GetNetWorksHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		val, ok := os.LookupEnv("DATA_SERVER_NETWORK")
		if !ok {
			val = "MOP Storage"
		}
		c.JSON(http.StatusOK, NewResponse(OKCode, strings.Split(val, ",")))
	}
}

// @Summary traffic price
// @Schemes
// @Description traffic price
// @Tags bills
// @Accept json
// @Produce json
// @Param   size     query    int     true        "buy size"
// @Success 200 string ok
// @Router /buy/traffic [get]
func BuyTrafficHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		val, ok := os.LookupEnv("DATA_SERVER_TRAFFIC_PRICE")
		if !ok {
			val = "0.0001"
		}

		price, err := decimal.NewFromString(val)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		size, err := strconv.ParseInt(c.DefaultQuery("size", "0"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, &map[string]interface{}{
			"size":      size,
			"amount":    price.Mul(decimal.NewFromInt(size)),
			"receiptor": "0x5529E3F428f42C23DDaCbb80fd46247B775725b1",
		}))
	}
}

// @Summary traffic price
// @Schemes
// @Description traffic price
// @Tags bills
// @Accept json
// @Produce json
// @Param   size     query    int     true        "buy size"
// @Success 200 string ok
// @Router /buy/storage [get]
func BuyStorageHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		val, ok := os.LookupEnv("DATA_SERVER_STORAGE_PRICE")
		if !ok {
			val = "0.0001"
		}

		price, err := decimal.NewFromString(val)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		size, err := strconv.ParseInt(c.DefaultQuery("size", "0"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, NewResponse(ExecuteCode, err))
			return
		}

		c.JSON(http.StatusOK, NewResponse(OKCode, &map[string]interface{}{
			"size":      size,
			"amount":    price.Mul(decimal.NewFromInt(size)),
			"receiptor": "0x5529E3F428f42C23DDaCbb80fd46247B775725b1",
		}))
	}
}

func GetContractHandler(db *dataservice.DataService) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, NewResponse(OKCode, map[string]string{
			"abi":     ERC20ABI,
			"address": "0xcc07B6277c7B4614884eEa59972AfEa17f40716F",
		}))
	}
}

const ERC20ABI = `[
  {
    "inputs": [
      {
        "internalType": "string",
        "name": "name_",
        "type": "string"
      },
      {
        "internalType": "string",
        "name": "symbol_",
        "type": "string"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "constructor"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "owner",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "spender",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "value",
        "type": "uint256"
      }
    ],
    "name": "Approval",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "from",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "to",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "value",
        "type": "uint256"
      }
    ],
    "name": "Transfer",
    "type": "event"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "owner",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "spender",
        "type": "address"
      }
    ],
    "name": "allowance",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "spender",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "approve",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "balanceOf",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "decimals",
    "outputs": [
      {
        "internalType": "uint8",
        "name": "",
        "type": "uint8"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "spender",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "subtractedValue",
        "type": "uint256"
      }
    ],
    "name": "decreaseAllowance",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "spender",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "addedValue",
        "type": "uint256"
      }
    ],
    "name": "increaseAllowance",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "name",
    "outputs": [
      {
        "internalType": "string",
        "name": "",
        "type": "string"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "symbol",
    "outputs": [
      {
        "internalType": "string",
        "name": "",
        "type": "string"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "totalSupply",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "recipient",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "transfer",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "sender",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "recipient",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "transferFrom",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]`

// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/asset/{id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "asset",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket object"
                ],
                "summary": "asset",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "bucket id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "folder id",
                        "name": "fid",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/bills/storage": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "pagination query storage bills",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bills"
                ],
                "summary": "list storage bills",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "page number",
                        "name": "page_num",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "page size",
                        "name": "page_size",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.BillStorage"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "add storage bill",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bills"
                ],
                "summary": "add storage bill",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.BillStorage"
                        }
                    }
                }
            }
        },
        "/bills/traffic": {
            "get": {
                "description": "pagination query traffic bills",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bills"
                ],
                "summary": "list traffic bills",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "page number",
                        "name": "page_num",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "page size",
                        "name": "page_size",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.BillTraffic"
                        }
                    }
                }
            },
            "post": {
                "description": "add traffic bill",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bills"
                ],
                "summary": "add traffic bill",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.BillTraffic"
                        }
                    }
                }
            }
        },
        "/buckets": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "pagination list buckets",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket"
                ],
                "summary": "list buckets",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "page number",
                        "name": "page_num",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "page size",
                        "name": "page_size",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.Bucket"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "add bucket",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket"
                ],
                "summary": "add bucket",
                "parameters": [
                    {
                        "description": "bucket info",
                        "name": "bucket",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/routers.Bucket"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.Bucket"
                        }
                    }
                }
            }
        },
        "/buckets/{id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "bucket info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket"
                ],
                "summary": "bucket info",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "bucket id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.Bucket"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "remove bucket",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket"
                ],
                "summary": "remove bucket",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "bucket id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.Bucket"
                        }
                    }
                }
            }
        },
        "/buckets/{id}/objects": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "pagination query bucket objects",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket object"
                ],
                "summary": "list bucket objects",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "bucket id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "folder id",
                        "name": "fid",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "page number",
                        "name": "page_num",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "page size",
                        "name": "page_size",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.BucketObject"
                        }
                    }
                }
            }
        },
        "/buckets/{id}/objects/{fid}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "bucket object info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket object"
                ],
                "summary": "bucket object info",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "bucket id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "folder id",
                        "name": "fid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.BucketObject"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "remove bucket object",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket object"
                ],
                "summary": "remove bucket object",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "bucket id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "folder id",
                        "name": "fid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.BucketObject"
                        }
                    }
                }
            }
        },
        "/buckets/{id}/objects/{name}": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "add bucket folder",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket object"
                ],
                "summary": "add bucket object",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "bucket id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "folder id",
                        "name": "fid",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "name",
                        "name": "name",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "cid",
                        "name": "cid",
                        "in": "path"
                    },
                    {
                        "description": "object info",
                        "name": "object",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/routers.AddBucketReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.BucketObject"
                        }
                    }
                }
            }
        },
        "/buy/storage": {
            "get": {
                "description": "traffic price",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bills"
                ],
                "summary": "traffic price",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "buy size",
                        "name": "size",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/buy/traffic": {
            "get": {
                "description": "traffic price",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bills"
                ],
                "summary": "traffic price",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "buy size",
                        "name": "size",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/download/{cid}/{path}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "asset",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket object"
                ],
                "summary": "asset",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "cid",
                        "name": "cid",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "path",
                        "name": "path",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/finish/{asset_id}": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "asset finish",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bucket object"
                ],
                "summary": "asset finish",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/login": {
            "post": {
                "description": "user login",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "login"
                ],
                "summary": "user login",
                "parameters": [
                    {
                        "description": "user info",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/routers.LoginReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/overview": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "used storage",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "dashboard"
                ],
                "summary": "used storage",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.UsedTraffic"
                        }
                    }
                }
            }
        },
        "/user": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "user info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "setting"
                ],
                "summary": "user info",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.User"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "update user info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "setting"
                ],
                "summary": "update user info",
                "parameters": [
                    {
                        "description": "user setting",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/routers.UserReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.User"
                        }
                    }
                }
            }
        },
        "/user/actions": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "pagination query user actions",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "dashboard"
                ],
                "summary": "list user actions",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "page number",
                        "name": "page_num",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "page size",
                        "name": "page_size",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dataservice.UserAction"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "dataservice.BillStorage": {
            "type": "object",
            "properties": {
                "amount": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "hash": {
                    "description": "Email       string    ` + "`" + `json:\"email\" gorm:\"index\"` + "`" + `",
                    "type": "string"
                },
                "size": {
                    "type": "integer"
                },
                "size_str": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "dataservice.BillTraffic": {
            "type": "object",
            "properties": {
                "amount": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "hash": {
                    "description": "Email       string    ` + "`" + `json:\"email\" gorm:\"index\"` + "`" + `",
                    "type": "string"
                },
                "size": {
                    "type": "integer"
                },
                "size_str": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "dataservice.Bucket": {
            "type": "object",
            "properties": {
                "access": {
                    "type": "boolean"
                },
                "area": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "description": "Email     string    ` + "`" + `json:\"email\" gorm:\"index\"` + "`" + `",
                    "type": "string"
                },
                "network": {
                    "type": "string"
                },
                "total_num": {
                    "type": "integer"
                },
                "total_size": {
                    "type": "integer"
                },
                "total_size_str": {
                    "type": "string"
                }
            }
        },
        "dataservice.BucketObject": {
            "type": "object",
            "properties": {
                "asset_id": {
                    "type": "string"
                },
                "cid": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "is_folder": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "size": {
                    "type": "integer"
                },
                "size_str": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "total_num": {
                    "type": "integer"
                },
                "total_size": {
                    "type": "integer"
                },
                "total_size_str": {
                    "type": "string"
                },
                "traffic": {
                    "type": "integer"
                },
                "updated_at": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "dataservice.UsedStorage": {
            "type": "object",
            "properties": {
                "num": {
                    "type": "integer"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "dataservice.UsedTraffic": {
            "type": "object",
            "properties": {
                "num": {
                    "type": "integer"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "dataservice.User": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "first_name": {
                    "type": "string"
                },
                "last_name": {
                    "type": "string"
                },
                "total_storage": {
                    "type": "integer"
                },
                "total_storage_str": {
                    "type": "string"
                },
                "total_traffic": {
                    "type": "integer"
                },
                "total_traffic_str": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                },
                "used_storage": {
                    "type": "integer"
                },
                "used_storage_str": {
                    "type": "string"
                },
                "used_traffic": {
                    "type": "integer"
                },
                "used_traffic_str": {
                    "type": "string"
                }
            }
        },
        "dataservice.UserAction": {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "ip": {
                    "type": "string"
                }
            }
        },
        "routers.AddBucketReq": {
            "type": "object",
            "properties": {
                "cid": {
                    "type": "string"
                },
                "fid": {
                    "type": "integer"
                }
            }
        },
        "routers.Bucket": {
            "type": "object",
            "properties": {
                "access": {
                    "type": "boolean"
                },
                "area": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "network": {
                    "type": "string"
                }
            }
        },
        "routers.LoginReq": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "routers.OverView": {
            "type": "object",
            "properties": {
                "buckets": {
                    "type": "integer"
                },
                "objects": {
                    "type": "integer"
                },
                "total_storage": {
                    "type": "integer"
                },
                "total_storage_str": {
                    "type": "string"
                },
                "used_storage": {
                    "type": "integer"
                },
                "used_storage_str": {
                    "type": "string"
                }
            }
        },
        "routers.UserReq": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "first_name": {
                    "type": "string"
                },
                "last_name": {
                    "type": "string"
                },
                "new_password": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "127.0.0.1:8080",
	BasePath:    "/api/v1",
	Schemes:     []string{},
	Title:       "DataServer Swagger API",
	Description: "This is a sample server DataServer server.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register("swagger", &s{})
}

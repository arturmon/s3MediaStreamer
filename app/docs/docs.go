// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
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
        "/albums": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "responds with the list of all albums as JSON.",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "album-controller"
                ],
                "summary": "Show the list of all albums.",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Page number",
                        "name": "page",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "Number of items per page",
                        "name": "page_size",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Field to sort by (e.g., 'created_at')",
                        "name": "sort_by",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Sort order ('asc' or 'desc')",
                        "name": "sort_order",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Filter criteria",
                        "name": "filter",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/model.Album"
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/albums/add": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "adds an album from JSON received in the request body.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "album-controller"
                ],
                "summary": "Adds an album from JSON.",
                "parameters": [
                    {
                        "description": "Album details",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.Album"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/model.Album"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/albums/delete/{code}": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "locates the album whose ID value matches the id parameter and deletes it.",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "album-controller"
                ],
                "summary": "Deletes album whose ID value matches the code.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Code album",
                        "name": "code",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/albums/deleteAll": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Delete ALL.",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "album-controller"
                ],
                "summary": "Complete removal of all albums.",
                "responses": {
                    "204": {
                        "description": "No Content",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/albums/update": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "updates an existing album with new data based on the ID parameter sent by the client.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "album-controller"
                ],
                "summary": "Updates an existing album with new data.",
                "parameters": [
                    {
                        "description": "Updated album details",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.Album"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Album"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/albums/{code}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "locates the album whose ID value matches the id parameter sent by the client,\nthen returns that album as a response.",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "album-controller"
                ],
                "summary": "Album whose ID value matches the id.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Code album",
                        "name": "code",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Album"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/health": {
            "get": {
                "description": "Checks and returns the current health status of the application",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health-controller"
                ],
                "summary": "Get health status of the application",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/monitoring.HealthResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/monitoring.HealthResponse"
                        }
                    }
                }
            }
        },
        "/otp/disable": {
            "post": {
                "description": "Disable OTP for a user by setting 'otp_enabled' to 'false' in the database.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "OTP"
                ],
                "summary": "Disable OTP for a user.",
                "parameters": [
                    {
                        "description": "OTP input data",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.OTPInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OTP disabled successfully",
                        "schema": {
                            "$ref": "#/definitions/model.OkLoginResponce"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "User not found",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to update OTP status",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/otp/generate": {
            "post": {
                "description": "Generate an OTP token for a user and store it in the database.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "OTP"
                ],
                "summary": "Generate OTP for a user.",
                "parameters": [
                    {
                        "description": "OTP input data",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.OTPInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OTP generated successfully",
                        "schema": {
                            "$ref": "#/definitions/model.OkGenerateOTP"
                        }
                    },
                    "400": {
                        "description": "Invalid refresh payload",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Failed to find user or invalid email/password",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to update OTP secret or URL",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/otp/validate": {
            "post": {
                "description": "Validates a One-Time Password (OTP) for a user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "OTP"
                ],
                "summary": "Validates a One-Time Password (OTP).",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "user_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "OTP Input",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.OTPInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OTP Valid",
                        "schema": {
                            "$ref": "#/definitions/model.OkResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request - Invalid OTP",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found - User not found",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/otp/verify": {
            "post": {
                "description": "Verify the OTP token for a user and update 'otp_enabled' and 'otp_verified' fields in the database.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "OTP"
                ],
                "summary": "Verify OTP for a user.",
                "parameters": [
                    {
                        "description": "OTP input data",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.OTPInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OTP verified successfully",
                        "schema": {
                            "$ref": "#/definitions/model.OkLoginResponce"
                        }
                    },
                    "400": {
                        "description": "Invalid token",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to update OTP status",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "Check if the application server is running",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health-controller"
                ],
                "summary": "Application liveness check function",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/users/delete": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Deletes the authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-controller"
                ],
                "summary": "Deletes a user.",
                "responses": {
                    "200": {
                        "description": "Success - User deleted",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "401": {
                        "description": "Unauthorized - User unauthenticated",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found - User not found",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/users/login": {
            "post": {
                "description": "Authenticates a user with provided email and password.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-controller"
                ],
                "summary": "Authenticates a user.",
                "parameters": [
                    {
                        "description": "Login User",
                        "name": "login",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Success",
                        "schema": {
                            "$ref": "#/definitions/model.OkLoginResponce"
                        }
                    },
                    "400": {
                        "description": "Bad Request - Incorrect Password",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found - User not found",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/users/logout": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Clears the authentication cookie, logging out the user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-controller"
                ],
                "summary": "Logs out a user.",
                "responses": {
                    "200": {
                        "description": "Success",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/users/me": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves user information based on JWT in the request's cookies",
                "consumes": [
                    "*/*"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-controller"
                ],
                "summary": "Get user information",
                "responses": {
                    "200": {
                        "description": "Success",
                        "schema": {
                            "$ref": "#/definitions/model.OkLoginResponce"
                        }
                    },
                    "401": {
                        "description": "Unauthenticated",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "User not found",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/users/refresh": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Validates the provided refresh token, generates a new access token, and returns it.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-controller"
                ],
                "summary": "Refreshes the access token using a valid refresh token.",
                "parameters": [
                    {
                        "description": "Refresh token",
                        "name": "refresh_token",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.ParamsRefreshTocken"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully refreshed access token",
                        "schema": {
                            "$ref": "#/definitions/model.ResponceRefreshTocken"
                        }
                    },
                    "400": {
                        "description": "Bad Request - Invalid refresh token",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized - Invalid refresh token",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/users/register": {
            "post": {
                "description": "Register a new user with provided name, email, and password.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user-controller"
                ],
                "summary": "Registers a new user.",
                "parameters": [
                    {
                        "description": "Register User",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.User"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/model.UserResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request - User with this email exists",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.Album": {
            "type": "object",
            "properties": {
                "artist": {
                    "type": "string",
                    "example": "Artist name"
                },
                "code": {
                    "type": "string",
                    "example": "I001"
                },
                "description": {
                    "type": "string",
                    "example": "A short description of the application"
                },
                "price": {
                    "type": "string",
                    "example": "{Number: 1.10, Currency: EUR}"
                },
                "sender": {
                    "type": "string",
                    "example": "amqp or rest"
                },
                "title": {
                    "type": "string",
                    "example": "Title name"
                }
            }
        },
        "model.ErrorResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "model.OTPInput": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.OkGenerateOTP": {
            "type": "object",
            "properties": {
                "base32": {},
                "otpauth_url": {}
            }
        },
        "model.OkLoginResponce": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "otp_enabled": {
                    "type": "boolean"
                },
                "refresh_token": {
                    "type": "string"
                },
                "role": {
                    "type": "string"
                }
            }
        },
        "model.OkResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "model.ParamsRefreshTocken": {
            "type": "object",
            "properties": {
                "refresh_token": {
                    "type": "string",
                    "example": "eyJhbGciOiJIU....FnjPC-zct_EDkIuUviRNI"
                }
            }
        },
        "model.ResponceRefreshTocken": {
            "type": "object",
            "properties": {
                "access_token": {
                    "type": "string",
                    "example": "eyJhbGciOiJIU....FnjPC-zct_EDkIuUviRNI"
                },
                "refresh_token": {
                    "type": "string",
                    "example": "eyJhbGciOiJIU....FnjPC-zct_EDkIuUviRNI"
                }
            }
        },
        "model.User": {
            "description": "User account information with: user _id, name, email, password",
            "type": "object",
            "properties": {
                "email": {
                    "type": "string",
                    "example": "aaaa@aaaa.com"
                },
                "otp_auth_url": {
                    "type": "string"
                },
                "otp_enabled": {
                    "type": "boolean"
                },
                "otp_secret": {
                    "type": "string"
                },
                "otp_verified": {
                    "type": "boolean"
                },
                "password": {
                    "type": "string",
                    "example": "1111"
                }
            }
        },
        "model.UserResponse": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "refresh_token": {
                    "type": "string"
                }
            }
        },
        "monitoring.HealthResponse": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BasicAuth": {
            "type": "basic"
        }
    },
    "externalDocs": {
        "description": "OpenAPI",
        "url": "https://swagger.io/resources/open-api/"
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.0.1",
	Host:             "localhost:10000",
	BasePath:         "/v1",
	Schemes:          []string{"http", "https"},
	Title:            "Sceleton Golang Application API",
	Description:      "This is a sample server Petstore server.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

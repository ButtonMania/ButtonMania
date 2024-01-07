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
            "name": "ButtonMania Team",
            "email": "team@buttonmania.win"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/room/create": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "Create game room",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Client ID",
                        "name": "clientId",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Room ID",
                        "name": "roomId",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "userId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Telegram init data",
                        "name": "initData",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok"
                    },
                    "400": {
                        "description": "Room exists"
                    }
                }
            }
        },
        "/api/room/delete": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "Delete game room",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Client ID",
                        "name": "clientId",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Room ID",
                        "name": "roomId",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "userId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Telegram init data",
                        "name": "initData",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok"
                    },
                    "400": {
                        "description": "Room cannot be deleted"
                    },
                    "404": {
                        "description": "Room not found"
                    }
                }
            }
        },
        "/api/room/stats": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "Get room stats",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Client ID",
                        "name": "clientId",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Room ID",
                        "name": "roomId",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/protocol.GameRoomStats"
                        }
                    },
                    "400": {
                        "description": "Room id is too long"
                    },
                    "404": {
                        "description": "Room not found"
                    }
                }
            }
        },
        "/ws": {
            "get": {
                "summary": "Handles WebSocket connections",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Client ID",
                        "name": "clientId",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Room ID",
                        "name": "roomId",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "userId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "User locale",
                        "name": "locale",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Telegram init data",
                        "name": "initData",
                        "in": "query"
                    }
                ],
                "responses": {}
            }
        }
    },
    "definitions": {
        "protocol.GameRoomStats": {
            "type": "object",
            "properties": {
                "bestOverallDuration": {
                    "type": "integer"
                },
                "bestTodaysDuration": {
                    "type": "integer"
                },
                "countActive": {
                    "type": "integer"
                },
                "countLeaderboard": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "buttonmania.win",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "ButtonMania API",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

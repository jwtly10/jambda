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
            "name": "jwtly10",
            "url": "https://www.github.com/jwtly10/jambda"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/file/upload": {
            "post": {
                "description": "Uploads a zip file, validates its contents, and processes it in storage. The zip file must contain a \"bootstrap\" executable.",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "files"
                ],
                "summary": "Upload and process a file",
                "parameters": [
                    {
                        "type": "file",
                        "description": "File to upload",
                        "name": "upload",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "File uploaded and processed successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1",
	Host:             "localhost:8080",
	BasePath:         "/v1/api",
	Schemes:          []string{},
	Title:            "Jambda - Serverless framework",
	Description:      "A WIP serverless framework for running functions similar to AWS Lambda",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}


window.onload = function() {
  // Build a system
  var url = window.location.search.match(/url=([^&]+)/);
  if (url && url.length > 1) {
    url = decodeURIComponent(url[1]);
  } else {
    url = window.location.origin;
  }
  var options = {
  "swaggerDoc": {
    "openapi": "3.0.3",
    "info": {
      "title": "Lvovich — склонение русских ФИО, городов и организаций",
      "version": "1.0.0",
      "description": "REST-обёртка над SOAP-сервисом. Для SOAP используйте `/soap`, WSDL — `/wsdl`."
    },
    "servers": [
      {
        "url": "/",
        "description": "Сервер склонения"
      }
    ],
    "paths": {
      "/api/incline": {
        "post": {
          "summary": "Склонение ФИО по падежам",
          "tags": [
            "ФИО"
          ],
          "requestBody": {
            "required": true,
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "SurName": {
                      "type": "string",
                      "example": "Иванов",
                      "description": "Фамилия"
                    },
                    "FirstName": {
                      "type": "string",
                      "example": "Иван",
                      "description": "Имя"
                    },
                    "SecondName": {
                      "type": "string",
                      "example": "Иванович",
                      "description": "Отчество"
                    },
                    "declension": {
                      "type": "string",
                      "example": "dative",
                      "description": "Падеж: nominative, genitive, dative, accusative, instrumental, prepositional"
                    },
                    "format": {
                      "type": "string",
                      "example": "full",
                      "enum": [
                        "full",
                        "initials"
                      ],
                      "description": "Формат ответа"
                    }
                  }
                }
              }
            }
          },
          "responses": {
            "200": {
              "description": "Результат склонения",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "FirstName": {
                        "type": "string"
                      },
                      "SurName": {
                        "type": "string"
                      },
                      "SecondName": {
                        "type": "string"
                      },
                      "gender": {
                        "type": "string"
                      },
                      "initials": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      },
      "/api/gender": {
        "post": {
          "summary": "Определение пола по ФИО",
          "tags": [
            "ФИО"
          ],
          "requestBody": {
            "required": true,
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "SurName": {
                      "type": "string",
                      "example": "Смирнова"
                    },
                    "FirstName": {
                      "type": "string",
                      "example": "Анна"
                    },
                    "SecondName": {
                      "type": "string",
                      "example": ""
                    }
                  }
                }
              }
            }
          },
          "responses": {
            "200": {
              "description": "Пол",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "gender": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      },
      "/api/city/in": {
        "post": {
          "summary": "Город в предложном падеже (в каком?)",
          "tags": [
            "Города"
          ],
          "requestBody": {
            "required": true,
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "name": {
                      "type": "string",
                      "example": "Москва"
                    },
                    "gender": {
                      "type": "string",
                      "example": "female"
                    }
                  }
                }
              }
            }
          },
          "responses": {
            "200": {
              "description": "Город в предложном падеже",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "name": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      },
      "/api/city/from": {
        "post": {
          "summary": "Город в родительном падеже (из какого?)",
          "tags": [
            "Города"
          ],
          "requestBody": {
            "required": true,
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "name": {
                      "type": "string",
                      "example": "Москва"
                    },
                    "gender": {
                      "type": "string",
                      "example": "female"
                    }
                  }
                }
              }
            }
          },
          "responses": {
            "200": {
              "description": "Город в родительном падеже",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "name": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      },
      "/api/city/to": {
        "post": {
          "summary": "Город в винительном падеже (в какой?)",
          "tags": [
            "Города"
          ],
          "requestBody": {
            "required": true,
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "name": {
                      "type": "string",
                      "example": "Москва"
                    }
                  }
                }
              }
            }
          },
          "responses": {
            "200": {
              "description": "Город в винительном падеже",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "name": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      },
      "/api/org/in": {
        "post": {
          "summary": "Организация в предложном падеже (в какой?)",
          "tags": [
            "Организации"
          ],
          "requestBody": {
            "required": true,
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "name": {
                      "type": "string",
                      "example": "ООО «Ромашка»"
                    }
                  }
                }
              }
            }
          },
          "responses": {
            "200": {
              "description": "Организация в предложном падеже",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "name": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      },
      "/api/org/from": {
        "post": {
          "summary": "Организация в родительном падеже (из какой?)",
          "tags": [
            "Организации"
          ],
          "requestBody": {
            "required": true,
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "name": {
                      "type": "string",
                      "example": "ООО «Ромашка»"
                    }
                  }
                }
              }
            }
          },
          "responses": {
            "200": {
              "description": "Организация в родительном падеже",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "name": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      },
      "/api/org/to": {
        "post": {
          "summary": "Организация в винительном падеже (в какую?)",
          "tags": [
            "Организации"
          ],
          "requestBody": {
            "required": true,
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "name": {
                      "type": "string",
                      "example": "ООО «Ромашка»"
                    }
                  }
                }
              }
            }
          },
          "responses": {
            "200": {
              "description": "Организация в винительном падеже",
              "content": {
                "application/json": {
                  "schema": {
                    "type": "object",
                    "properties": {
                      "name": {
                        "type": "string"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "customOptions": {}
};
  url = options.swaggerUrl || url
  var urls = options.swaggerUrls
  var customOptions = options.customOptions
  var spec1 = options.swaggerDoc
  var swaggerOptions = {
    spec: spec1,
    url: url,
    urls: urls,
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  }
  for (var attrname in customOptions) {
    swaggerOptions[attrname] = customOptions[attrname];
  }
  var ui = SwaggerUIBundle(swaggerOptions)

  if (customOptions.oauth) {
    ui.initOAuth(customOptions.oauth)
  }

  if (customOptions.preauthorizeApiKey) {
    const key = customOptions.preauthorizeApiKey.authDefinitionKey;
    const value = customOptions.preauthorizeApiKey.apiKeyValue;
    if (!!key && !!value) {
      const pid = setInterval(() => {
        const authorized = ui.preauthorizeApiKey(key, value);
        if(!!authorized) clearInterval(pid);
      }, 500)

    }
  }

  if (customOptions.authAction) {
    ui.authActions.authorize(customOptions.authAction)
  }

  window.ui = ui
}

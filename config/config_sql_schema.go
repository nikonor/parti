package config
var (
SQLConfigSchema = `
{
    "definitions": {},
    "$schema": "http://json-schema.org/draft-07/schema#",
    "$id": "http://example.com/example",
    "type": "object",
    "title": "",
    "properties": {
      "environment": {
        "$id": "/properties/environment",
        "type": "object",
        "title": "Окружение"
      },
      "functional": {
        "$id": "/properties/functional",
        "type": "object",
        "title": "Функционал"
      },
      "integration": {
        "$id": "/properties/integration",
        "type": "object",
        "title": "Интеграция"
      },
      "module": {
        "$id": "/properties/module",
        "type": "object",
        "title": "Модуль",
        "properties": {
          "config": {
            "$id": "/properties/module/properties/config",
            "type": "object",
            "title": "Настройки модуля",
            "properties": {
              "revision": {
                "$id": "/properties/module/properties/config/properties/revision",
                "type": "integer",
                "title": "Номер ревизии"
              },
              "updated": {
                "$id": "/properties/module/properties/config/properties/updated",
                "type": "string",
                "title": "Дата обновления"
              }
            },
            "required": [
              "revision",
              "updated"
            ]
          },
          "http": {
            "$id": "/properties/module/properties/http",
            "type": "object",
            "title": "Настройки сети",
            "properties": {
              "cancelingTimeOut": {
                "$id": "/properties/module/properties/http/properties/cancelingTimeOut",
                "type": "integer",
                "title": "Время завершения, с"
              },
              "port": {
                "$id": "/properties/module/properties/http/properties/port",
                "type": "string",
                "title": "Порт"
              },
              "socket": {
                "$id": "/properties/module/properties/http/properties/socket",
                "type": "string",
                "title": "Имя файла сокета"
              },
              "title": {
                "$id": "/properties/module/properties/http/properties/title",
                "type": "string",
                "title": "Заголовок"
              },
              "useSocket": {
                "$id": "/properties/module/properties/http/properties/useSocket",
                "type": "boolean",
                "title": "Использовать сокет"
              }
            },
            "required": [
              "cancelingTimeOut",
              "port",
              "socket",
              "title",
              "useSocket"
            ]
          }
        },
        "required": [
          "config",
          "http"
        ]
      }
    },
    "required": [
      "environment",
      "functional",
      "integration",
      "module"
    ]
  }
`)


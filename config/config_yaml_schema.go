package config
var (
YAMLConfigSchema = `
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "definitions": {},
    "id": "http://example.com/example.json",
    "properties": {
        "journal_name": {
            "description": "Имя приложения в логах",
            "id": "/properties/journal_name",
            "title": "Имя",
            "type": "string"
        },
        "log_level": {
            "description": "Уровень логирования: LOG_ERR: 3, LOG_WARNING: 4, LOG_INFO: 6, LOG_DEBUG: 7",
            "id": "/properties/log_level",
            "title": "Уровень логирования",
            "type": "integer",
            "maximum": 7,
            "minimum": 3
        },
        "adminka": {
            "description": "Адрес админки",
            "id": "/properties/adminka",
            "title": "Уровень логирования",
            "type": "string"
        },
        "sql": {
            "description": "Настройки SQL",
            "id": "/properties/sql",
            "title": "Настройки SQL",
            "properties": {
                "connect_string": {
                    "description": "Строка подключения к SQL",
                    "id": "/properties/sql/properties/connect_string",
                    "title": "Строка подключения к SQL",
                    "type": "string"
                },
                "pguser": {
                    "id": "/properties/sql/properties/pguser",
                    "title": "Имя пользователя для подключения к SQL",
                    "type": "string"
                },
                "pgpassword": {
                    "id": "/properties/sql/properties/pgpassword",
                    "title": "Пароль для подключения к SQL",
                    "type": "string"
                },
                "pghost": {
                    "id": "/properties/sql/properties/pghost",
                    "title": "Хост для подключения к SQL",
                    "type": "string"
                },
                "pgport": {
                    "id": "/properties/sql/properties/pgport",
                    "title": "Порт для подключения к SQL",
                    "type": "string"
                },
                "pgdatabase": {
                    "id": "/properties/sql/properties/pgdatabase",
                    "title": "Название базы данных для подключения к SQL",
                    "type": "string"
                },
                "pgflags": {
                    "id": "/properties/sql/properties/pgflags",
                    "title": "Флаги для подключения к SQL",
                    "type": "string"
                },
                "pool_size": {
                    "description": "Кол-во соединений с БД в пуле",
                    "id": "/properties/sql/properties/pool_size",
                    "title": "Кол-во соединений с БД в пуле",
                    "type": "integer"
                },
                "statement_timeout": {
                    "id": "/properties/sql/properties/pool_size",
                    "title": "Таймаут sql запросов (в миллисекундах)",
                    "type": "integer"
                }
            },
            "type": "object",
            "required": ["statement_timeout"]
        },
        "stdout": {
            "description": "Вывод логов в stdout",
            "id": "/properties/stdout",
            "title": "Вывод логов в stdout",
            "type": "boolean"
        },
        "syslog": {
            "default": true,
            "description": "Вывод логов в syslog",
            "id": "/properties/syslog",
            "title": "Вывод логов в syslog",
            "type": "boolean"
        }
    },
    "required": [
        "stdout",
        "journal_name",
        "sql",
        "syslog",
        "log_level"
    ],
    "type": "object"
}`)


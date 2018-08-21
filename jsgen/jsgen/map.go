package jsgen

var titlesMap = map[string]string{
	"/properties/module": "Модуль",

	"/properties/module/properties/http":                             "Настройки сети",
	"/properties/module/properties/http/properties/port":             "Порт",
	"/properties/module/properties/http/properties/title":            "Заголовок",
	"/properties/module/properties/http/properties/socket":           "Имя файла сокета",
	"/properties/module/properties/http/properties/useSocket":        "Использовать сокет",
	"/properties/module/properties/http/properties/adminkaTimeOut":   "Таймаут админки",
	"/properties/module/properties/http/properties/cancelingTimeOut": "Время завершения, с",

	"/properties/module/properties/config":                     "Настройки модуля",
	"/properties/module/properties/config/properties/updated":  "Дата обновления",
	"/properties/module/properties/config/properties/revision": "Номер ревизии",

	"/properties/module/properties/instance":                             "Настройка экземпляра модуля",
	"/properties/module/properties/instance/properties/timeout":          "Таймаут соединения с админкой",
	"/properties/module/properties/instance/properties/attempts":         "Попыток зарегистрироваться",
	"/properties/module/properties/instance/properties/maxCount":         "Количество экземпляров",
	"/properties/module/properties/instance/properties/loginTimeout":     "Таймаут регистрации в админке",
	"/properties/module/properties/instance/properties/bgInstancesCount": "Количество экземпляров ФП",

	"/properties/functional": "Функционал",

	"/properties/environment": "Окружение",

	"/properties/environment/properties/redis":                     "Настройки redis",
	"/properties/environment/properties/redis/properties/db":       "База",
	"/properties/environment/properties/redis/properties/host":     "Имя сервера",
	"/properties/environment/properties/redis/properties/port":     "Порт",
	"/properties/environment/properties/redis/properties/password": "Пароль",
	"/properties/environment/properties/redis/properties/poolSize": "Размер пула",

	"/properties/integration": "Интеграция",
}

var queuesHeadMap = map[string]string{
	"/properties/environment/properties/rmq": "Очереди",
}

var queuesMap = map[string]string{
	"/properties/environment/properties/rmq/properties/%s":                                        "Очередь",
	"/properties/environment/properties/rmq/properties/%s/properties/queue":                       "Настройки очереди",
	"/properties/environment/properties/rmq/properties/%s/properties/queue/properties/name":       "Имя",
	"/properties/environment/properties/rmq/properties/%s/properties/queue/properties/durable":    "Непрерывное соединение",
	"/properties/environment/properties/rmq/properties/%s/properties/queue/properties/exclusive":  "Эксклюзивное соединение",
	"/properties/environment/properties/rmq/properties/%s/properties/queue/properties/autoDelete": "Автоудаление",

	"/properties/environment/properties/rmq/properties/%s/properties/title": "Заголовок",

	"/properties/environment/properties/rmq/properties/%s/properties/exchange":                       "Настройки exchange",
	"/properties/environment/properties/rmq/properties/%s/properties/exchange/properties/name":       "Имя",
	"/properties/environment/properties/rmq/properties/%s/properties/exchange/properties/type":       "Тип",
	"/properties/environment/properties/rmq/properties/%s/properties/exchange/properties/durable":    "Непрерывное соединение",
	"/properties/environment/properties/rmq/properties/%s/properties/exchange/properties/exclusive":  "Эксклюзивное соединение",
	"/properties/environment/properties/rmq/properties/%s/properties/exchange/properties/autoDelete": "Автоудаление",

	"/properties/environment/properties/rmq/properties/%s/properties/routingKey": "Rouring key",

	"/properties/environment/properties/rmq/properties/%s/properties/connectConf":                     "Настройки соединения",
	"/properties/environment/properties/rmq/properties/%s/properties/connectConf/properties/addr":     "Адрес сервера",
	"/properties/environment/properties/rmq/properties/%s/properties/connectConf/properties/port":     "Порт",
	"/properties/environment/properties/rmq/properties/%s/properties/connectConf/properties/login":    "Имя пользователя",
	"/properties/environment/properties/rmq/properties/%s/properties/connectConf/properties/rmqType":  "Тип соединения",
	"/properties/environment/properties/rmq/properties/%s/properties/connectConf/properties/password": "Пароль",

	"/properties/environment/properties/rmq/properties/%s/properties/rmqPoolSize": "Размер пула",

	"/properties/environment/properties/rmq/properties/%s/properties/connectToQueue": "Очередь активна",
}

var customTitlesMap = map[string]string{
	"/properties/functional/properties/defaultTimeout": "Таймаут по умолчанию",
	"/properties/functional/properties/serviceID":      "Идентификатор службы",
	"/properties/functional/properties/workers":        "Количество ФП",
}

var regexpTitlesMap = map[string]string{
	"/properties/integration/properties/.+?/properties/.+?/properties/endpoint": "URL",
	"/properties/integration/properties/.+?/properties/.+?/properties/timeout":  "Таймаут",
	"/properties/integration/properties/.+?/properties/.+?/properties/method":   "Метод",
}

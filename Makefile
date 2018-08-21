DATE := $(shell date +"%Y-%m-%dT%H:%M:%S")
COMMIT := $(shell git log -1 --pretty=format:%H)
TARGET := $(shell echo $${PWD\#\#*/})
CFG=./cfg/config.yaml
DIR :=/tmp
GO_VERSION := $(shell go version)
MODULE_VERSION := $(shell [ `git describe --tags --abbrev=0 2>/dev/null | wc -l` -ne "0" ] &&  git describe --tags --abbrev=0 || echo "0.1falfa")
TEMP_VERSION=1.9
PGHOST := $(shell cat ./cfg/config.yaml | awk '/pghost/{print $$2}')
PGPORT := $(shell cat ./cfg/config.yaml | awk '/pgport/{print $$2}')
PGUSER := $(shell cat ./cfg/config.yaml | awk '/pguser/{print $$2}')
PGPASSWD := $(shell cat ./cfg/config.yaml | awk '/pgpassword/{print $$2}')
PGDB := $(shell cat ./cfg/config.yaml | awk '/pgdatabase/{print $$2}')

LDFLAGS=-ldflags '-X=main.moduleCompileTime=$(DATE) -X=main.goTemplateVersion=$(TEMP_VERSION) -X=main.gitCommit=$(COMMIT) -X "main.goVersion=$(GO_VERSION)" -X "main.moduleVersion=$(MODULE_VERSION)"'
CLI_PARAM=""

schema:
	@echo "Создаем JSON-Schema"
	@$(shell cd jsgen && cp ../config/sql_schema.json ../config/sql_schema1.json && go build  -o main ./ &&./main -s ../config/sql_schema1.json > ../config/sql_schema.json && rm -f ./main ../config/sql_schema1.json || mv ../config/sql_schema1.json ../config/sql_schema.json)

config/config_*_schema.go: config/yaml_schema.json config/sql_schema.json
	@echo "Создаем .go-файлы из схем"
	@echo "package config"  > config/config_yaml_schema.go
	@echo "var (" >> config/config_yaml_schema.go
	@echo 'YAMLConfigSchema = `' >> config/config_yaml_schema.go
	@cat config/yaml_schema.json >> config/config_yaml_schema.go
	@echo '`)' >> config/config_yaml_schema.go
	@echo  >> config/config_yaml_schema.go

	@echo "package config"  > config/config_sql_schema.go
	@echo "var (" >> config/config_sql_schema.go
	@echo 'SQLConfigSchema = `' >> config/config_sql_schema.go
	@cat config/sql_schema.json >> config/config_sql_schema.go
	@echo '`)' >> config/config_sql_schema.go
	@echo  >> config/config_sql_schema.go

pg_dump:
	@export PGPASSWORD=$(PGPASSWD) && pg_dump -C --inserts -h $(PGHOST) -p $(PGPORT) -d $(PGDB) -U $(PGUSER) -f ./sql/000-makefile.sql

build: schema config/config_*_schema.go
	@echo "Собираем бинарник"
	go build $(LDFLAGS) -o $(DIR)/$(TARGET)

run:  pg_dump schema build
	@echo "Запуск"
	$(DIR)/$(TARGET) -c $(CFG) -p /tmp/$(TARGET).pid -o /tmp/$(TARGET).out -e /tmp/$(TARGET).err $(CLI_PARAM)

socket: pg_dump schema build
	@echo "Запуск  через сокет"
	@echo
	$(DIR)/$(TARGET) -c $(CFG) -p /tmp/$(TARGET).pid -socket /tmp/$(TARGET).socket -o /tmp/$(TARGET).out -e /tmp/$(TARGET).err $(CLI_PARAM)
	
test:
	@echo "Запуск тестов"
	@go test -v $(CLI_PARAM)
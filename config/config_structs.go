package config

// Структуры конфига

// CfgSql настройки подключения и запросы к БД
type CfgSql struct {
	ConnectString    string `yaml:"connect_string" json:"connect_string"`
	PGUser           string `yaml:"pguser" json:"pguser"`
	PGPassword       string `yaml:"pgpassword" json:"pgpassword"`
	PGHost           string `yaml:"pghost" json:"pghost"`
	PGPort           string `yaml:"pgport" json:"pgport"`
	PGDataBase       string `yaml:"pgdatabase" json:"pgdatabase"`
	PGFlags          string `yaml:"pgflags" json:"pgflags"`
	PoolSize         int    `yaml:"pool_size" json:"pool_size"`
	StatementTimeout *int   `yaml:"statement_timeout" json:"statement_timeout"`
}

// CfgHttp структура для хранения данных http сервера
type CfgHttp struct {
	Title            string `json:"title"`
	Port             string `json:"port"`
	CancelingTimeOut int    `json:"cancelingTimeOut"`
	Socket           string `json:"socket"`
	UseSocket        bool   `json:"useSocket"`
	AdminkaTimeOut   int    `json:"adminkaTimeOut"`
}

// Config структура информации о конфиге
type CfgCfg struct {
	Updated  string `json:"updated"`
	Revision int    `json:"revision"`
}

// InstanceCfg - параметры для работы инстанса
type InstanceCfg struct {
	TimeOut          int `json:"timeout"`
	Attempts         int `json:"attempts"`
	LoginTimeOut     int `json:"loginTimeout"`
	MaxCount         int `json:"maxCount"`
	BGInstancesCount int `json:"bgInstancesCount"`
}

// CfgModule структура для хранения данных модуля
type CfgModule struct {
	Http     CfgHttp     `json:"http"`
	Config   CfgCfg      `json:"config"`
	Instance InstanceCfg `json:"instance"`
}

type ConnectConf struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Addr     string `json:"addr"`
	Port     string `json:"port"`
}

// CfgRmq настройки RabbitMQ
type CfgRmq struct {
	Title          string         `json:"title"`
	Connect        ConnectConf    `json:"connectConf"`
	Exchange       CfgRmqExchange `json:"exchange"`
	Queue          CfgRmqQueue    `json:"queue"`
	RoutingKey     string         `json:"routingKey"`
	PoolSize       int            `json:"rmqPoolSize" yaml:"rmqPoolSize"`
	ConnectToQueue bool           `json:"connectToQueue"`
}

// CfgRmqExchange настройки RabbitMQ exchange
type CfgRmqExchange struct {
	Name       string `yaml:"name" json:"name"`
	Type       string `yaml:"type" json:"type"`
	Durable    bool   `yaml:"durable" json:"durable"`
	AutoDelete bool   `yaml:"autoDelete" json:"autoDelete"`
}

// CfgRmqQueue настройки RabbitMQ очереди
type CfgRmqQueue struct {
	Name       string `yaml:"name" json:"name"`
	Durable    bool   `yaml:"durable" json:"durable"`
	Exclusive  bool   `yaml:"exclusive" json:"exclusive"`
	AutoDelete bool   `yaml:"autoDelete" json:"autoDelete"`
}

// CfgEnvironment структура для хранения данных модуля
type CfgEnvironment struct {
	Rmq     map[string]CfgRmq `yaml:"rmq" json:"rmq"`
	Rediska Redis             `yaml:"redis" json:"redis"`
}

type Redis struct {
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

// Config корневая структура
type Config struct {
	JournalName string         `yaml:"journal_name" json:"journal_name"`
	Syslog      bool           `yaml:"syslog" json:"syslog"`
	Stdout      bool           `yaml:"stdout" json:"stdout"`
	Adminka     string         `yaml:"adminka" json:"adminka"`
	LogLevel    int            `yaml:"log_level" json:"log_level"`
	NumCpu      int            `yaml:"num_cpu" json:"num_cpu"`
	Sql         CfgSql         `yaml:"sql" json:"sql"`
	Module      CfgModule      `json:"module"`
	Environment CfgEnvironment `json:"environment"`
	Functional  CfgFunctional  `json:"functional"`
	Integration CfgIntegration `json:"Integration"`
}

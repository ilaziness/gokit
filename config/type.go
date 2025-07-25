package config

type Mode string

type App struct {
	// 应用id
	Name string `mapstructure:"name"`
	// 服务权重
	Weight int `mapstructure:"weight"`
	// Mode debug, release
	Mode string `mapstructure:"mode"`
	Port uint16 `mapstructure:"port"`
	// 应用根目录，绝对路径
	RootDir       string `mapstructure:"root_dir"`
	Cors          *Cors  `mapstructure:"cors"`
	SessionSecret string `mapstructure:"session_secret"`
	// 是否记录请求日志
	LogReq bool `mapstructure:"log_req"`
	// 是否开启pprof
	Pprof bool `mapstructure:"pprof"`
}

type Cors struct {
	AllowOrigin      []string `mapstructure:"allow_origin"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// DB sql数据库配置
type DB struct {
	// sql方言 sqlite3 postgres mysql pgx
	Dialect         string `mapstructure:"dialect"`
	Debug           bool   `mapstructure:"debug"`
	DSN             string `mapstructure:"dsn"`
	Host            string `mapstructure:"host"`
	Port            uint16 `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DbName          string `mapstructure:"db_name"`
	Charset         string `mapstructure:"charset"`
	Timezone        string `mapstructure:"timezone"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifeTime int    `mapstructure:"conn_max_life_time"`
}

type Redis struct {
	Host string `mapstructure:"host"`
	Port uint16 `mapstructure:"port"`
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
	Db   uint   `mapstructure:"db"`
}

type RocketMq struct {
	Endpoint      string `mapstructure:"endpoint"`
	AccessKey     string `mapstructure:"access_key"`
	SecretKey     string `mapstructure:"secret_key"`
	Transaction   bool   `mapstructure:"transaction"`
	ProducerTopic string `mapstructure:"producer_topic"`
}

// Otel opentelemetry配置
type Otel struct {
	TraceExporterURL string `mapstructure:"trace_exporter_url"`
}

type NacosClient struct {
	NamespaceID         string `mapstructure:"namespace_id"`
	AccessKey           string `mapstructure:"access_key"`
	SecretKey           string `mapstructure:"secret_key"`
	Username            string `mapstructure:"username"`
	Password            string `mapstructure:"password"`
	NotLoadCacheAtStart bool   `mapstructure:"not_load_cache_at_start"`
	LogLevel            string `mapstructure:"log_level"`
}

type NacosServer struct {
	IP   string `mapstructure:"ip"`
	Port uint16 `mapstructure:"port"`
}

// Nacos 配置
type Nacos struct {
	DataID string        `mapstructure:"data_id"`
	Group  string        `mapstructure:"group"`
	Client NacosClient   `mapstructure:"client"`
	Server []NacosServer `mapstructure:"server"`
}

// TCPServer tcp服务配置
type TCPServer struct {
	Debug     bool   `mapstructure:"debug"`
	Address   string `mapstructure:"address"`
	WorkerNum int    `mapstructure:"worker_num"`
	// tls
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// UDPServer UDP服务配置
type UDPServer struct {
	Debug     bool   `mapstructure:"debug"`
	Address   string `mapstructure:"address"`
	WorkerNum int    `mapstructure:"worker_num"`
	// DTLS (TLS over UDP)
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// QUICServer QUIC服务配置
type QUICServer struct {
	Debug     bool   `mapstructure:"debug"`
	Address   string `mapstructure:"address"`
	WorkerNum int    `mapstructure:"worker_num"`
	// TLS配置 (QUIC必需)
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	// QUIC特定配置
	IdleTimeout      int  `mapstructure:"idle_timeout"`      // 空闲超时(秒)
	KeepAlive        int  `mapstructure:"keep_alive"`        // 保活间隔(秒)
	HandshakeTimeout int  `mapstructure:"handshake_timeout"` // 握手超时(秒)
	MaxStreams       int  `mapstructure:"max_streams"`       // 最大流数量
	Allow0RTT        bool `mapstructure:"allow_0rtt"`        // 允许0-RTT
}

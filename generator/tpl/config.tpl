[app]
id = "{{ .Name }}"
port = 9000
mode = "debug"
{{ if .RocketMQ }}
[rocket_mq]
endpoint = "127.0.0.1:8081"
access_key = ""
secret_key = ""
producer_topic = "test1"
{{ end }}

{{- if .Otel }}
[otel]
{{- if .OtelTrace }}
trace_enable = true
trace_exporter_url = "http://localhost:4318"
{{ end }}
{{- end }}

{{- if .Mysql }}
[db]
dsn = "root:root@tcp(127.0.0.1:3306)/ent_test"
{{- end }}
{{ if .Redis }}
[redis]
host = "127.0.0.1"
user = ""
pass = ""
{{- end }}
{{ if .Nacos }}
# nacos 配置中心获取配置
[nacos]
data_id = "web"
group = "gintpl"
    [nacos.client]
    namespace_id = "e4b98b9b-a0d0-402b-b282-602b6c2fa3ca"
    not_load_cache_at_start = true
    [nacos.server]
    ip = "127.0.0.1"
    port = 8848
{{- end }}

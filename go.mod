module github.com/BetaGoRobot/BetaGo

go 1.20

replace (
	github.com/BetaGoRobot/BetaGo/betagovar => ./betagovar
	github.com/BetaGoRobot/BetaGo/httptool => ./httptool
	github.com/BetaGoRobot/BetaGo/neteaseapi => ./neteaseapi
	github.com/BetaGoRobot/BetaGo/scheduletask => ./scheduletask
	github.com/BetaGoRobot/BetaGo/utility => ./utility
)

require (
	github.com/carlmjohnson/requests v0.23.5
	github.com/enescakir/emoji v1.0.0
	github.com/fasthttp/router v1.4.22
	github.com/go-resty/resty/v2 v2.11.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/gotify/go-api-client/v2 v2.0.4
	github.com/jordan-wright/email v4.0.1-0.20210109023952-943e75fe5223+incompatible
	github.com/json-iterator/go v1.1.12
	github.com/kevinmatthe/zaplog v0.1.6
	github.com/lonelyevil/kook v0.0.33
	github.com/lonelyevil/kook/log_adapter/plog v0.0.34
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/phuslu/log v1.0.88
	github.com/prometheus/client_golang v1.17.0
	github.com/redis/go-redis/v9 v9.3.1
	github.com/spyzhov/ajson v0.9.0
	github.com/tencentyun/cos-go-sdk-v5 v0.7.45
	github.com/valyala/fasthttp v1.51.0
	github.com/wcharczuk/go-chart/v2 v2.1.1
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.5.0 // indirect
	github.com/bits-and-blooms/bloom/v3 v3.3.1 // indirect
	github.com/blend/go-sdk v1.20220411.3 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/clbanning/mxj v1.8.4 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.21.4 // indirect
	github.com/go-openapi/errors v0.20.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/loads v0.21.2 // indirect
	github.com/go-openapi/runtime v0.25.0 // indirect
	github.com/go-openapi/spec v0.20.8 // indirect
	github.com/go-openapi/strfmt v0.21.7 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-openapi/validate v0.22.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mozillazg/go-httpheader v0.3.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/savsgio/gotils v0.0.0-20230208104028-c358bd845dee // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.11.3 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/image v0.11.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

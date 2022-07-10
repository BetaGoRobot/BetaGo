module github.com/BetaGoRobot/BetaGo

go 1.18

replace (
	github.com/BetaGoRobot/BetaGo/betagovar => ./betagovar
	github.com/BetaGoRobot/BetaGo/httptool => ./httptool
	github.com/BetaGoRobot/BetaGo/neteaseapi => ./neteaseapi
	github.com/BetaGoRobot/BetaGo/scheduletask => ./scheduletask
	github.com/BetaGoRobot/BetaGo/utility => ./utility
)

require (
	github.com/TwiN/go-away v1.6.7
	github.com/enescakir/emoji v1.0.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/heyuhengmatt/zaplog v0.1.2
	github.com/jordan-wright/email v4.0.1-0.20210109023952-943e75fe5223+incompatible
	github.com/json-iterator/go v1.1.12
	github.com/lonelyevil/khl v0.0.27
	github.com/lonelyevil/khl/log_adapter/plog v0.0.27
	github.com/phuslu/log v1.0.80
	github.com/tencentyun/cos-go-sdk-v5 v0.7.35
	github.com/wcharczuk/go-chart/v2 v2.1.0
	gorm.io/driver/postgres v1.3.8
	gorm.io/gorm v1.23.8
)

require (
	github.com/bits-and-blooms/bitset v1.2.2 // indirect
	github.com/bits-and-blooms/bloom/v3 v3.2.0 // indirect
	github.com/clbanning/mxj v1.8.4 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.12.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.11.0 // indirect
	github.com/jackc/pgx/v4 v4.16.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mozillazg/go-httpheader v0.3.1 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/image v0.0.0-20220617043117-41969df76e82 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

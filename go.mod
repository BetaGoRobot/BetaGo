module github.com/BetaGoRobot/BetaGo

go 1.17

replace (
	github.com/BetaGoRobot/BetaGo/betagovar => ./betagovar
	github.com/BetaGoRobot/BetaGo/httptool => ./httptool
	github.com/BetaGoRobot/BetaGo/neteaseapi => ./neteaseapi
	github.com/BetaGoRobot/BetaGo/scheduletask => ./scheduletask
	github.com/BetaGoRobot/BetaGo/utility => ./utility
)

require (
	github.com/lonelyevil/khl v0.0.19
	github.com/lonelyevil/khl/log_adapter/plog v0.0.19
	github.com/phuslu/log v1.0.71
	gorm.io/driver/postgres v1.2.3
	gorm.io/gorm v1.22.5
)

require (
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.10.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.2.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.9.1 // indirect
	github.com/jackc/pgx/v4 v4.14.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/text v0.3.7 // indirect
)

require (
	github.com/bits-and-blooms/bitset v1.2.1 // indirect
	github.com/bits-and-blooms/bloom/v3 v3.1.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/json-iterator/go v1.1.12
)

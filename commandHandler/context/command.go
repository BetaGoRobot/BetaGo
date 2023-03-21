package context

// const 长命令组
//
//	@param CommandContextTypeAddAdmin
const (
	CommandContextTypeAddAdmin    = "ADDADMIN"
	CommandContextTypeRemoveAdmin = "REMOVEADMIN"
	CommandContextTypeShowAdmin   = "SHOWADMIN"
	CommandContextTypeDeleteAll   = "DELETEALL"
	CommandContextReconnect       = "RECONNECT"
	CommandContextTypeCal         = "SHOWCAL"
	CommandContextTypeCalLocal    = "SHOWCALLOCAL"
	CommandContextTypeHelper      = "HELP"
	CommandContextTypeHitokoto    = "ONEWORD"
	CommandContextTypeMusic       = "SEARCHMUSIC"
	CommandContextTypeRoll        = "ROLL"
	CommandContextTypePing        = "PING"
	CommandContextTypeUser        = "GETUSER"
	CommandContextTypeTryPanic    = "TRYPANIC"
	CommandContextTypeDailyRate   = "RATE"
	CommandContextTypeGPT         = "GPT"
	CommandContextTypeNews        = "NEWS"
)

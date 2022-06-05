package main

const (
	//MemberJoined 新成员加入Type
	MemberJoined = "joined_guild"

	//MemberExited 成员退出Type
	MemberExited = "exited_guild"

	//MemberUpdate 成员信息更新Type
	MemberUpdate = "updated_guild_member"

	//MemberOnline 成员上线Type
	MemberOnline = "guild_member_online"

	//MemberOffline 成员离线Type
	MemberOffline = "guild_member_offline"

	//ChannelAddReaction 频道内用户添加 reaction
	ChannelAddReaction = "added_reaction"

	//ChannelDelReaction 频道内用户取消 reaction
	ChannelDelReaction = "deleted_reaction"

	//ChannelMessageUpdate 频道消息更新
	ChannelMessageUpdate = "updated_message"

	//ChannelMessageRemove 频道消息被删除
	ChannelMessageRemove = "deleted_message"

	//ChannelAdded 新增频道
	ChannelAdded = "added_channel"

	//ChannelModified 修改频道
	ChannelModified = "added_channel"

	//ChannelDeleted 删除频道
	ChannelDeleted = "deleted_channel"

	//ChannelTopMessage 新增频道置顶消息
	ChannelTopMessage = "pinned_message"

	//ChannelTopMessageCancel 取消频道置顶消息
	ChannelTopMessageCancel = "unpinned_message"
)

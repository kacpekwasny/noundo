package noundo

import "time"

func CreateUserInfo(user UserIdentityIface, usingHistoryName string) UserInfo {
	return UserInfo{
		username:       user.Username(),
		FUsername:      user.FullUsername(),
		parentServer:   user.ParentServerName(),
		UserProfileURL: ProfileURL(user, usingHistoryName),
	}
}

func CreateTimeStamp() TimeStampable {
	return TimeStampable{
		Timestamp: UnixTimeNow(),
	}
}

func UnixTimeNow() int64 {
	return time.Now().Unix()
}

package data

type RefreshToken struct {
	/*
	 UserUUID is the public identity that crosses service boundaries; the
	 numeric DB id stays internal to auth's database.
	*/
	UserUUID string `json:"user_uuid"`
	Device   string `json:"device"`
}

package dto

type UserBasicInfo struct {
	UID      string `json:"uid" form:"uid"`
	Username string `json:"username" form:"username"`
	Email    string `json:"email" form:"email"`
	Phone    string `json:"phone" form:"phone"`
}

type UserProfile struct {
	Username          string `json:"username" form:"username"`
	PersonalSignature string `json:"personal_signature" form:"personal_signature"`
	Level             string `json:"level" form:"level"`
	TotalGames        int    `json:"total_games" form:"total_games"`
	TotalWins         int    `json:"total_wins" form:"total_wins"`
	TotalLosses       int    `json:"total_losses" form:"total_losses"`
}

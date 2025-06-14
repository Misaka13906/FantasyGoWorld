package request

type UpdateUserProfile struct {
	UID               string `json:"uid" form:"uid" binding:"required"`
	Username          string `json:"username" form:"username" binding:"omitempty"`
	PersonalSignature string `json:"personal_signature" form:"personal_signature"`
}

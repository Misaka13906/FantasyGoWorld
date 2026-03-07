package e

type Code int

const (
	// 200xx Success
	SUCCESS        Code = 20000
	NoRecordsFound Code = 20001

	// 400xx Client Error
	RequestNotSatisfied Code = 40001
	RequestFieldError   Code = 40002

	// 401xx Authentication
	Unauthorized        Code = 40100
	TokenExpired        Code = 40101
	RefreshTokenExpired Code = 40102
	PasswordTooShort    Code = 40110
	UsernameExists      Code = 40111
	InvalidCredentials  Code = 40112

	// 402xx Room
	RoomFull       Code = 40200
	RoomNotWaiting Code = 40201
	NotRoomOwner   Code = 40202
	WrongRoomPass  Code = 40203

	// 403xx Engine/Rules
	Forbidden     Code = 40300
	GameNotActive Code = 40300 // Reusing 40300 as per spec or split if needed
	NotYourTurn   Code = 40301
	Occupied      Code = 40302
	Suicide       Code = 40303
	Ko            Code = 40304
	InvalidStep   Code = 40305
	InvalidPos    Code = 40306

	// 404xx Not Found
	NotFound Code = 40400

	// 409xx Conflict
	Conflict Code = 40900

	// 500xx Server Error
	ERROR         Code = 50001
	DatabaseError Code = 50002
	RedisError    Code = 50003
)

var MsgFlags = map[Code]string{
	SUCCESS:        "success",
	NoRecordsFound: "no records found",

	RequestNotSatisfied: "request does not satisfy requirements",
	RequestFieldError:   "request field error",

	Unauthorized:        "unauthorized",
	TokenExpired:        "token expired",
	RefreshTokenExpired: "refresh token expired",
	PasswordTooShort:    "password too short",
	UsernameExists:      "username already exists",
	InvalidCredentials:  "invalid credentials",

	RoomFull:       "room is full",
	RoomNotWaiting: "room not in waiting state",
	NotRoomOwner:   "not room owner",
	WrongRoomPass:  "wrong room password",

	Forbidden:   "forbidden / game not active",
	NotYourTurn: "not your turn",
	Occupied:    "illegal move: occupied",
	Suicide:     "illegal move: suicide",
	Ko:          "illegal move: ko",
	InvalidStep: "invalid step number",
	InvalidPos:  "invalid position",

	NotFound: "not found",
	Conflict: "conflict",

	ERROR:         "internal server error",
	DatabaseError: "database error",
	RedisError:    "redis error",
}

func GetMsg(code Code) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[ERROR]
}

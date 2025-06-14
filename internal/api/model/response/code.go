package response

import "net/http"

const (
	SuccessCode = iota
	ErrorCodeRecordNotFound
	ErrorCodeListNoRecords
	ErrorCodeConflict
	ErrorCodeDBError
	ErrorCodeHashingError
	ErrorCodeUIDGeneration
	ErrorCodeTokenGeneration
	ErrorCodeBadRequest     = http.StatusBadRequest
	ErrorCodeUnauthorized   = http.StatusUnauthorized
	ErrorCodeInternalServer = http.StatusInternalServerError
)

var MsgMap = map[uint]string{
	SuccessCode:              "success",
	ErrorCodeRecordNotFound:  "Record not found", // 预期一条记录，但未找到
	ErrorCodeListNoRecords:   "Record num 0",     // 预期一条或多条记录，但未找到任何记录
	ErrorCodeBadRequest:      "Invalid request parameters",
	ErrorCodeUnauthorized:    "Invalid username or password",
	ErrorCodeConflict:        "Resource conflict", // e.g., Username already exists
	ErrorCodeInternalServer:  "Internal server error",
	ErrorCodeDBError:         "Database error",
	ErrorCodeTokenGeneration: "Failed to generate token",
	ErrorCodeHashingError:    "Failed to process password",
	ErrorCodeUIDGeneration:   "Failed to generate unique identifier",
}

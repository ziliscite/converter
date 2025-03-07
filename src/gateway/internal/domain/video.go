package domain

// Video will be sent to the converter service
type Video struct {
	UserId    int64  `json:"user_id"`
	UserEmail string `json:"user_email"`
	FileName  string `json:"file_name"`
	FileKey   string `json:"file_key"`
}

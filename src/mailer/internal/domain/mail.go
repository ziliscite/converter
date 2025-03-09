package domain

type Metadata struct {
	UserId   int64  `json:"user_id"`
	FileName string `json:"file_name"`
	VideoKey string `json:"video_key"`
	AudioKey string `json:"audio_key"`
}

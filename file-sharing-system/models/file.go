package models

type File struct {
    ID       int    `json:"id"`
    UserID   int    `json:"user_id"`
    FileName string `json:"file_name"`
    Size     int64  `json:"size"`
    S3URL    string `json:"s3_url"`
}

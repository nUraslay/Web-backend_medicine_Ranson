package handler


type ErrorResponse struct {
	Status      string `json:"status" example:"error"`
	Description string `json:"description" example:"описание ошибки"`
}

type MessageResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message"`
}

type SignResponse struct {
	Status  string         `json:"status" example:"success"`
	Data    SignSerializer `json:"data"`
	Message string         `json:"message,omitempty"`
}

type SignListResponse struct {
	Status string           `json:"status" example:"success"`
	Data   []SignSerializer `json:"data"`
}

type LikeData struct {
	SignID     uint `json:"sign_id"`
	IsLiked    int  `json:"is_liked" example:"1"`
	LikesCount int  `json:"likes_count"`
}

type LikeResponse struct {
	Status  string   `json:"status" example:"success"`
	Message string   `json:"message"`
	Data    LikeData `json:"data"`
}

type UserData struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

type RegisterResponse struct {
	Status  string   `json:"status" example:"success"`
	Data    UserData `json:"data"`
	Message string   `json:"message"`
}

type LoginData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int64  `json:"expires_in" example:"3600"`
}

type LoginResponse struct {
	Status  string    `json:"status" example:"success"`
	Data    LoginData `json:"data"`
	Message string    `json:"message"`
}

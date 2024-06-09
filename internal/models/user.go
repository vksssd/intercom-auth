package models


type User struct {
	ID string `json:"_id" bson:"_id"`
	UserId string `json:"userid" bson:"userid"`
	Username string `json:"username" bson:"username"`
	Email string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	CreateAt string `json:"created_at" bson:"created_at"`
	UpdateAt string `json:"updated_at" bson:"update_at"`
}
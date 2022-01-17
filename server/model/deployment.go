package model

type Deployment struct {
	BaseModel
	ContainerId string `json:"container_id" gorm:"not null"`
	ProjectId   uint   `json:"project_id" gorm:"not null"`
	UserId      uint   `json:"user_id" gorm:"not null"`
	Status      string `json:"status"`
}

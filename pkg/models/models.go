package models

type Message struct {
	ID      uint   `json:"id,omitempty" gorm:"primaryKey"`
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}

type HandledMessage struct {
	ID      uint `gorm:"primaryKey"`
	From    string
	To      string
	Message string
	Handled bool
}

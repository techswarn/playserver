package models

// type Response[T any] struct {
// 	Success bool `json:"success"`
// 	Message string `json:"message"`
// 	Data T `json:"Data"`
// }
type Response[T any] struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data T `json:"Data"`
}


type ErrorResponse struct {
	ErrorMessage string `json:"error_message"`
	Field        string `json:"field"`
}
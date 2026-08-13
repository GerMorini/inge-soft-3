package dto

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Fields  map[string][]string `json:"fields,omitempty"`
}

type RegisterRequest struct {
	FirstName string         `json:"firstName"`
	LastName  string         `json:"lastName"`
	Phone     string         `json:"phone"`
	Address   AddressRequest `json:"address"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
}

type AddressRequest struct {
	Street    string `json:"street"`
	Number    string `json:"number"`
	Apartment string `json:"apartment"`
	City      string `json:"city"`
	Province  string `json:"province"`
}

type RegisteredUserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int    `json:"expiresIn"`
}

type CurrentUserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

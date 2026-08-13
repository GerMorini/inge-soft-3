package dao

type CreateUserParams struct {
	FirstName    string
	LastName     string
	Phone        string
	Street       string
	StreetNumber string
	Apartment    *string
	City         string
	Province     string
	Username     string
	Email        string
	PasswordHash string
}

type CreatedUser struct {
	ID       int64
	Username string
	Email    string
}

type Credentials struct {
	ID           int64
	Username     string
	PasswordHash string
}

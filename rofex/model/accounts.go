package model

// Account is a user account as returned by /rest/accounts.
// At least 'name' is present per the Postman collection; extra fields are ignored.
type Account struct {
	Name string `json:"name"`
}

// AccountsResponse is the container for an account listing.
//
//	{
//	  "accounts": [ { "name": "..." }, ... ]
//	}
type AccountsResponse struct {
	Status   string    `json:"status,omitempty"`
	Accounts []Account `json:"accounts"`
}

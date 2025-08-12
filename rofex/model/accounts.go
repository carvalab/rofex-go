package model

import "encoding/json"

// Account representa una cuenta de usuario como la devuelta por /rest/accounts.
// El esquema puede incluir más campos; conservadoramente mapeamos los conocidos.
// Al menos 'name' está presente según el uso de la colección Postman.
type Account struct {
	Name string `json:"name"`
	// Captura campos adicionales sin fallar en decodificación estricta
	Extra json.RawMessage `json:"-"`
}

// AccountsResponse es el contenedor para el listado de cuentas.
//
//	{
//	  "accounts": [ { "name": "..." }, ... ]
//	}
type AccountsResponse struct {
	Status   string    `json:"status,omitempty"`
	Accounts []Account `json:"accounts"`
}

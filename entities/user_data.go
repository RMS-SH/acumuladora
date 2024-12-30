////////////////////////////////////////////////////////////////////////////////
// entities/user_data.go
////////////////////////////////////////////////////////////////////////////////

package entities

// UserData contém os dados específicos de um usuário (UserNS),
// incluindo o corpo da requisição e a URL para onde esses dados
// serão enviados posteriormente.
type UserData struct {
	Body   []BodyItem
	URL    string
	UserNS string
}

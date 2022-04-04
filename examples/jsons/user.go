//go:generate easyjson -all

package jsons

type User struct {
	Id        string `json: "Id"`
	FirstName string `json: "FirstName"`
	LastName  string `json: "LastName"`
	Stamp     int    `json: "Stamp"`
}

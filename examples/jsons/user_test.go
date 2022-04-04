package jsons

import (
	"github.com/mailru/easyjson"
	Assert "github.com/stretchr/testify/assert"
	"testing"
)

func TestUser(t *testing.T) {
	//json := []byte("{\"Id\":\"some-id\",\"FirstName\":\"Dude\",\"LastName\":\"Lastnameski\"}")
	t.Run("Json Marshalling", func(t *testing.T) {
		//a := assert.New(t)

		user := &User{Id: "Id", FirstName: "FirstName", LastName: "Lastname"}
		marshaledByteArray, err := easyjson.Marshal(user)
		if err != nil {
			t.Error(err)
		}
		//marshaledString := string(marshaledByteArray)

		newUser := &User{}
		e := easyjson.Unmarshal(marshaledByteArray, newUser)
		if e != nil {
			t.Error(e)
		}
		t.Run("Information Accuracy", func(t *testing.T) {
			assert := Assert.New(t)
			assert.Equal(user.Id, "Id", "ID is not equal")
			assert.Equal(user.FirstName, "FirstName", "Invalid value")
			assert.Equal(user.LastName, "Lastname", "Invalid value")
		})
	})
}

package user

import (
	"os"
	"io/ioutil"
	"testing"
)

func TempFilename(t *testing.T) string {
	f, err := ioutil.TempFile("", "filestoredb")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestFilestoreCreation(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	store, err := NewFileStore(tempFilename)
	if err != nil {
		t.Error(err)
	}else if store == nil {
		t.Error("Got back a nil FileStore")
	}

	u := &User{
		Username: "Alfred",
		Email: "alfred@thebatcave.com ",
	}

	err = store.CreateUser(u)
	if err != nil {
		t.Error(err)
	}

	u.SetPassword("moo")

	otherUser := store.GetUser("Username", "Alfred")

	if !otherUser.CheckPassword("moo"){
		t.Error("Passwords failed to match")
	}

	if otherUser.CheckPassword("Moo"){
		t.Error("Passwords should not have matched")
	}

	u.SetPassword("Moo")

	otherUser = store.GetUser("Username", "  alfred")

	if !otherUser.CheckPassword("Moo"){
		t.Error("Passwords failed to match")
	}

	if otherUser.CheckPassword("moo"){
		t.Error("Passwords should not have matched")
	}

	err = store.CreateUser(&User{
		Username: "  ALFred",
		Email: "  ALfred@thebatcave.com  ",
	})
	if err == nil {
		t.Error("Should have field to insert non-unique email")
	}

}
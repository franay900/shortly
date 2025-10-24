package jwt

import (
	"testing"
)

func TestJwtCreate(t *testing.T) {
	const email = "email4@mail.ru"
	jwtService := NewJWT("/2+XnmJGz1j3ehIVI/5P9kl+CghrE3DcS7rnT+qar5w=")

	token, err := jwtService.Create(JWTData{
		Email: email,
	})

	if err != nil {
		t.Fatal(err)
	}
	isValid, data := jwtService.Parse(token)

	if !isValid {
		t.Fatalf("token is not valid")
	}

	if data.Email != email {
		t.Fatalf("data.Email != email")
	}

}

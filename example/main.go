package main

import (
	"fmt"
	"time"

	"github.com/vladevelops/motp"
)

type SessionData struct {
	UserID string
}

func main() {

	otp := motp.NewMemoryOtp[SessionData]()

	// Create an OTP
	code, err := otp.GenerateOTP(motp.OtpData[SessionData]{
		Key:        "user_123",
		Expiration: time.Now().Add(5 * time.Minute),
		Data:       SessionData{UserID: "usr_42"},
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Generated OTP:", code)
}

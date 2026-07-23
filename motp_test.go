package motp

import (
	"testing"
	"time"
)

func TestCodeLive(t *testing.T) {

	type OtpEmailType int

	const (
		REGISTER OtpEmailType = iota
		LOGIN
		EMAIL_UPDATE
	)
	otp := NewMemoryOtp(WithOTPLen[OtpEmailType](10))

	key := "myemail"
	code, err := otp.GenerateOTP(OtpData[OtpEmailType]{
		Key:        key,
		Data:       LOGIN,
		Expiration: time.Now().Add(time.Second * 30),
	})

	if err != nil {
		t.Errorf("err: %v \n", err)
		return
	}

	otp_data, err := otp.CheckOTP(key, code)

	if err != nil {
		t.Errorf("err: %v \n", err)
		return
	}

	if otp_data.Code != code {
		t.Errorf("ERROR: expected code: %v got: %v \n", code, otp_data.Code)
		return
	}

	if err := otp.DeleteOTP(key); err != nil {
		t.Errorf("ERROR: code deletion error by key %v\n", key)
	}

	_, err = otp.CheckOTP(key, code)

	if err != nil {
		if err.Error() != ErrCodeNotFound.Error() {
			t.Errorf("err: %v \n", err)
			return
		}
	}

}

func TestCodeLiveCleaner(t *testing.T) {

	type OtpEmailType int

	const (
		REGISTER OtpEmailType = iota
		LOGIN
		EMAIL_UPDATE
	)
	otp := NewMemoryOtp(WithOTPLen[OtpEmailType](10), WithOTPCleanupTime[OtpEmailType](time.Second*3))

	key := "myemail"
	code, err := otp.GenerateOTP(OtpData[OtpEmailType]{
		Key:        key,
		Data:       LOGIN,
		Expiration: time.Now().Add(time.Second * 2),
	})

	if err != nil {
		t.Errorf("err: %v \n", err)
		return
	}

	time.Sleep(time.Second * 4)

	_, err = otp.CheckOTP(key, code)

	if err != nil {
		if err.Error() != ErrCodeNotFound.Error() {
			t.Errorf("err: %v \n", err)
			return
		}
	}

}

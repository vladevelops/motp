// Package motp provides a thread-safe, generic in-memory OTP manager.
//
// It supports customizable character sets, expiration windows, and automatic
// background purging of stale tokens.
//
// Basic usage:
//
//	store := motp.NewMemoryOtp[MyData]()
//	code, err := store.GenerateOTP(data)
package motp

import (
	"context"
	"crypto/rand"
	"errors"
	"sync"
	"time"
)

var ErrCodeNotFound = errors.New("code not present")

type OtpData[T any] struct {
	Key        string
	Expiration time.Time
	Code       string
	Data       T
}

type MemoryOtp[T any] interface {
	GenenareOTP(otp_data OtpData[T]) (code string, err error)
	CheckOTP(key, code string) (ret_otp OtpData[T], err error)
	DeleteOTP(key string) error
}

type otp_manager[T any] struct {
	code_len     int
	otp_alphabet string
	cleanup_time time.Duration

	mu         sync.RWMutex
	stored_otp map[string]OtpData[T]
}

type OtpConf[T any] func(*otp_manager[T])

// NewMemoryOtp creates and initializes a new in-memory OTP manager.
//
// By default, it is initialized with:
//   - Code length: 7 digits
//   - Alphabet: "0973546281" (numeric)
//   - Cleanup interval: 15 seconds
//
// Custom configuration can be applied by passing functional options (confs).
// On creation, it automatically spins up a background goroutine to periodically
// purge expired OTPs from memory.
//
// Returns a thread-safe implementation of the MemoryOtp interface.
func NewMemoryOtp[T any](confs ...OtpConf[T]) MemoryOtp[T] {
	otp := otp_manager[T]{
		code_len:     7,
		otp_alphabet: "0973546281",
		cleanup_time: time.Second * 15,
		stored_otp:   make(map[string]OtpData[T]),
	}

	for _, f := range confs {
		f(&otp)
	}

	go otp.delete_expired_codes(context.Background())
	return &otp
}

// WithOTPLen sets a custom length for generated OTP codes.
//
// If not specified, the manager defaults to generating 7-character codes.
func WithOTPLen[T any](opt_len int) OtpConf[T] {
	return func(opt *otp_manager[T]) {
		opt.code_len = opt_len
	}
}

// WithOTPAlphabet configures the set of characters used when generating OTP codes.
//
// If not specified, the manager defaults to "0973546281".
func WithOTPAlphabet[T any](new_alphabet string) OtpConf[T] {
	return func(opt *otp_manager[T]) {
		opt.otp_alphabet = new_alphabet
	}
}

// WithOTPCleanupTime sets the interval at which the background purger runs
// to remove expired OTPs from memory.
//
// If not specified, the manager defaults to cleaning up every 15 seconds.
func WithOTPCleanupTime[T any](new_time time.Duration) OtpConf[T] {
	return func(opt *otp_manager[T]) {
		opt.cleanup_time = new_time
	}
}

// GenerateOTP stores a new OTP payload mapped to its unique key.
//
// If otp_data.Code is empty, a random code of length code_len is automatically generated.
// If an active code already exists for the provided Key, GenerateOTP returns an error
// to prevent overwriting unexpired tokens.
//
// Returns the generated (or provided) OTP code string, or an error if the key is already in use.
// This method is thread-safe.
func (otp *otp_manager[T]) GenenareOTP(otp_data OtpData[T]) (code string, err error) {
	if otp_data.Code == "" {
		otp_data.Code = otp.generate_otp_code(otp.code_len)
	}

	otp.mu.Lock()
	defer otp.mu.Unlock()

	_, present := otp.stored_otp[otp_data.Key]

	if present {
		return "", errors.New("key already present, wait TTL, or delete")
	}

	otp.stored_otp[otp_data.Key] = otp_data

	return otp_data.Code, nil
}

// CheckOTP validates the provided code against the active OTP stored for the given key.
//
// It performs an O(1) lookup using the key parameter, verifies the code matches, and ensures
// the OTP has not passed its expiration time.
//
// Returns the stored OtpData if valid. Returns an error if the key is missing, the code
// is incorrect, or the code has expired. This method is thread-safe.
func (otp *otp_manager[T]) CheckOTP(key, code string) (ret_otp OtpData[T], err error) {

	otp.mu.RLock()
	defer otp.mu.RUnlock()
	check_otp, is_present := otp.stored_otp[key]

	if !is_present || check_otp.Code != code {
		return ret_otp, ErrCodeNotFound
	}

	if time.Now().After(check_otp.Expiration) {
		return ret_otp, ErrCodeNotFound
	}

	return check_otp, nil
}

// DeleteOTP removes an active OTP entry from memory by its key.
//
// It performs a thread-safe lookup and deletes the entry mapped to the provided key.
// Returns an error if no active OTP exists for the specified key.
//
// This method is thread-safe and can be used to manually invalidate or consume
// an OTP once it has been verified.
func (otp *otp_manager[T]) DeleteOTP(key string) error {
	otp.mu.Lock()
	defer otp.mu.Unlock()
	_, is_present := otp.stored_otp[key]
	if !is_present {
		return ErrCodeNotFound
	}
	delete(otp.stored_otp, key)
	return nil
}

func (otp *otp_manager[T]) delete_expired_codes(ctx context.Context) {
	tick := time.NewTicker(otp.cleanup_time)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			otp.cleanup()
		}
	}
}

func (otp *otp_manager[T]) generate_otp_code(length int) string {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return ""
	}
	otpCharsLength := len(otp.otp_alphabet)
	for i := range length {
		buffer[i] = otp.otp_alphabet[int(buffer[i])%otpCharsLength]
	}
	return string(buffer)
}

func (otp *otp_manager[T]) cleanup() {
	now := time.Now()
	otp.mu.Lock()
	defer otp.mu.Unlock() // Now defer cleanly releases lock when cleanup() finishes

	for key, otpData := range otp.stored_otp {
		if now.After(otpData.Expiration) {
			delete(otp.stored_otp, key)
		}
	}
}

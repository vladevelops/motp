# motp

A fast, thread-safe, generic in-memory OTP (One-Time Password) manager for Go with automatic cleanup.

## Features

- **Generic Payloads**: Attach any custom metadata to your OTP using Go generics.
- **Thread-Safe**: Safe for concurrent access across goroutines.
- **Auto-Purge**: Background worker automatically cleans up expired tokens.
- **Customizable**: Functional options to tweak code length, character set, and cleanup intervals.

## Installation

```bash

go get github.com/vladevelops/motp

```

## Quick Start

```go
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
	// 1. Initialize the generic OTP store for SessionData
	otp := motp.NewMemoryOtp[SessionData]()

	// 2. Generate an OTP
	code, err := otp.GenerateOTP(motp.OtpData[SessionData]{
		Key:        "user_123",
		Expiration: time.Now().Add(5 * time.Minute),
		Data:       SessionData{UserID: "usr_42"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Generated OTP:", code)

	// 3. Verify / Check the OTP
	data, err := otp.CheckOTP("user_123", code)
	if err != nil {
		fmt.Println("Invalid or expired OTP:", err)
		return
	}
	fmt.Printf("OTP valid! User ID: %s\n", data.Data.UserID)

	// 4. Manually revoke or clear an OTP
	if err := otp.DeleteOTP("user_123"); err != nil {
		fmt.Println("Error deleting OTP:", err)
	}
}

```

## API Reference

The `NewMemoryOtp[T]()` constructor returns a thread-safe `MemoryOtp[T]` interface exposing three core methods:

```go

type MemoryOtp[T any] interface {
    GenerateOTP(otp_data OtpData[T]) (code string, err error)
    CheckOTP(key, code string) (ret_otp OtpData[T], err error)
    DeleteOTP(key string) error
}

```

### Method Overview

#### `GenerateOTP(otp_data OtpData[T]) (code string, err error)`

Generates a random OTP code tied to the provided `Key` and stores it alongside your custom data payload `T`.

* **Errors:** Returns an error if the key is empty, the expiration is already in the past, or code generation fails.

#### `CheckOTP(key, code string) (ret_otp OtpData[T], err error)`

Validates an incoming OTP code against the given lookup key.

* **On Success:** Returns the original `OtpData[T]` struct (including your custom payload).
* **Errors:** Returns an error if the key doesn't exist, the token has expired, or the provided `code` doesn't match.

#### `DeleteOTP(key string) error`

Manually removes an OTP entry from memory before its scheduled expiration. Useful for explicitly invalidating sessions, handling manual cancellations, or cleaning up after consumption.

## License

MIT

```

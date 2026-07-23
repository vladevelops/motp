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
	// Initialize with options
	otp := motp.NewMemoryOtp[SessionData]()

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
```

## License

MIT

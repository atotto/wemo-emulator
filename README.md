# Alexa Multiple Belkin wemo switch emulator

Package wemo provides a library that emulates the belkin wemo switch.
In this library, it is possible to define multiple switches and implement their behavior with callback functions.

example/hello/main.go:
```go
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	wemo "github.com/atotto/wemo-emulator"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	light1 := wemo.ConfigSwitchService("my light 1", "8080", "1234", "abcd",
		func(ctx context.Context, state bool) bool {
			log.Println("light 1 on")
			// write code
			return true
		}, func(ctx context.Context, state bool) bool {
			log.Println("light 1 off")
			// write code
			return false
		})

	light2 := wemo.ConfigSwitchService("my light 2", "8081", "123", "abc",
		func(ctx context.Context, state bool) bool {
			log.Println("light 2 on")
			// write code
			return true
		}, func(ctx context.Context, state bool) bool {
			log.Println("light 2 off")
			// write code
			return false
		})

	if err := wemo.StartSwitchServices(ctx, light1, light2); err != nil {
		if err != context.Canceled {
			log.Fatal(err)
		}
	}
}
```

How to use:

1. Run `go install github.com/atotto/wemo-emulator@latest`
2. Copy the `example/hello/main.go` code from this repository.
3. Define switches and callbacks.
4. Run `go run main.go`
5. Enjoy
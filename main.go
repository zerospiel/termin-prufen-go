package main

import (
	"fmt"

	"github.com/zerospiel/termin-prufen-go/pkg/prufen"
)

func main() {
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	// defer cancel()

	runner := prufen.NewRunner(prufen.Options{
		// DebugFunc: log.Printf,
	})
	fmt.Println(runner.RunOnce())
	// if err := runner.Run(ctx); err != nil {
	// 	log.Fatalf("failed to run server: %v", err)
	// }
}

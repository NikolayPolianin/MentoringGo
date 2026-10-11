package main

import (
	"checkgo/TwoService/app"
	"context"
)

func main() {
	ctx := context.Background()
	app.Run(ctx)
}

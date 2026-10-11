package main

import "context"
import "checkgo/OneService/app"

func main() {
	ctx := context.Background()
	app.Run(ctx)
}

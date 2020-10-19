package main

import (
	"os"

	"github.com/jochenboesmans/forward-financial-statements/predict"
	"github.com/jochenboesmans/forward-financial-statements/pull"
	"github.com/joho/godotenv"
)

func main() {
	switch os.Args[1] {
	case "pull":
		pull.Pull()
	case "predict":
		predict.Predict()
	}
}

func init() {
	godotenv.Load()
}

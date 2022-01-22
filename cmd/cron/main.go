package main

import (
	"log"

	"github.com/robfig/cron"
)

func main() {
	log.Println("Starting...")

	c := cron.New()
	c.AddFunc("*/10 * * * * *", func() {
		log.Println("Run every ten seconds...")
	})

	c.Start()

	select {}
}

package main

import (
	"context"
	"fmt"
	"os"

	"game-metrics-reporter/infrai"
)

type matchBatch struct {
	Mode            string
	MatchesStarted  int
	MatchesFinished int
	PlayersOnline   int
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	client := &infrai.Client{Key: key}
	batch := matchBatch{Mode: "ranked", MatchesStarted: 12, MatchesFinished: 10, PlayersOnline: 384}
	tags := map[string]string{"mode": batch.Mode, "service": "game-backend"}

	metrics := []infrai.Metric{
		{Type: "counter", Name: "game.matches_started", Value: float64(batch.MatchesStarted), Tags: tags},
		{Type: "counter", Name: "game.matches_finished", Value: float64(batch.MatchesFinished), Tags: tags},
		{Type: "gauge", Name: "game.players_online", Value: float64(batch.PlayersOnline), Tags: tags},
	}
	for _, metric := range metrics {
		if err := client.Report(context.Background(), metric); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	fmt.Println("reported game match counters and online-player gauge")
}

// The client call is intentionally named after the capability: infrai.metrics.report.

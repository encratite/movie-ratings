package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"slices"

	"github.com/encratite/commons"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"gonum.org/v1/gonum/stat"
)

type movieRatings struct {
	name string
	column int
	scale float64
	ratings []float64
	rmse float64
	meanDelta float64
}

func main() {
	analyzeRatings()
}

func analyzeRatings() {
	userRatings, ratings := readRatings()
	header := []string{
		"Name",
		"RMSE",
		"Mean Delta",
	}
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	rmseValues := []float64{}
	for i := range ratings {
		r := ratings[i]
		rmse := getRMSE(userRatings.ratings, r.ratings)
		meanDelta := getMeanDelta(r.ratings, userRatings.ratings)
		ratings[i].rmse = rmse
		ratings[i].meanDelta = meanDelta
		rmseValues = append(rmseValues, rmse)
	}
	minRMSE := slices.Min(rmseValues)
	maxRMSE := slices.Max(rmseValues)
	rows := [][]string{}
	for _, r := range ratings {
		rmseString := fmt.Sprintf("%.3f", r.rmse)
		if r.rmse >= maxRMSE {
			rmseString = red(rmseString)
		}
		if r.rmse <= minRMSE {
			rmseString = green(rmseString)
		}
		meanDeltaString := fmt.Sprintf("+%.3f", r.meanDelta)
		if r.meanDelta > 0 {
			meanDeltaString = green(meanDeltaString)
		}
		if r.meanDelta < 0 {
			meanDeltaString = red(meanDeltaString)
		}
		row := []string{
			r.name,
			rmseString,
			meanDeltaString,
		}
		rows = append(rows, row)
	}
	alignments := []tw.Align{
		tw.AlignDefault,
		tw.AlignRight,
		tw.AlignRight,
	}
	tableConfig := tablewriter.WithConfig(tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{AutoFormat: tw.Off},
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
		}},
	)
	alignmentConfig := tablewriter.WithAlignment(alignments)
	fmt.Printf("\n")
	table := tablewriter.NewTable(os.Stdout, tableConfig, alignmentConfig)
	table.Header(header)
	table.Bulk(rows)
	table.Render()
	fmt.Printf("\n")
	movies := len(userRatings.ratings)
	fmt.Printf("Movies evaluated: %d\n\n", movies)
}

func readRatings() (movieRatings, []movieRatings) {
	ratings := []movieRatings{
		newMovieRatings("My Rating", 1, 100.0),
		newMovieRatings("IMDB", 2, 10.0),
		newMovieRatings("Rotten Tomatoes Critics", 3, 100.0),
		newMovieRatings("Rotten Tomatoes Users", 4, 100.0),
		newMovieRatings("Metacritic Critics", 5, 100.0),
		newMovieRatings("Metacritic Users", 6, 10.0),
		newMovieRatings("Letterboxd", 7, 5.0),
	}
	commons.ReadCSV("ratings.csv", func (columns []string) {
		name := columns[0]
		for i := range ratings {
			currentRating := &ratings[i]
			ratingString := columns[currentRating.column]
			rating, err := commons.ParseFloat(ratingString)
			if err != nil {
				log.Fatalf("Failed to parse rating for \"%s\" of movie %s: %s", currentRating.name, name, ratingString)
			}
			rating /= currentRating.scale
			currentRating.add(rating, name)
		}
	})
	return ratings[0], ratings[1:]
}

func (m *movieRatings) add(rating float64, name string) {
	if rating < 0.0 || rating > 1.0 || math.IsNaN(rating) {
		log.Fatalf("Invalid rating for movie %s", name)
	}
	m.ratings = append(m.ratings, rating)
}

func getRMSE(values1, values2 []float64) float64 {
	if len(values1) != len(values2) {
		log.Fatalf("Invalid number of values, can't calculate RMSE")
	}
	sum := 0.0
	for i := range values1 {
		delta := values1[i] - values2[i]
		sum += delta * delta
	}
	rmse := math.Sqrt(sum / float64(len(values1)))
	return rmse
}

func getMeanDelta(values1, values2 []float64) float64 {
	if len(values1) != len(values2) {
		log.Fatalf("Invalid number of values, can't calculate mean delta")
	}
	deltas := []float64{}
	for i := range values1 {
		delta := values1[i] - values2[i]
		deltas = append(deltas, delta)
	}
	mean := stat.Mean(deltas, nil)
	return mean
}

func newMovieRatings(name string, column int, scale float64) movieRatings {
	return movieRatings{
		name: name,
		column: column,
		scale: scale,
	}
}
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
	userRatings := newMovieRatings("My Rating")
	imdbRatings := newMovieRatings("IMDB")
	rottenCriticsRatings := newMovieRatings("Rotten Tomatoes Critics")
	rottenUsersRatings := newMovieRatings("Rotten Tomatoes Users")
	metacriticCriticsRatings := newMovieRatings("Metacritic Critics")
	metacriticUsersRatings := newMovieRatings("Metacritic Users")
	commons.ReadCSV("ratings.csv", func (columns []string) {
		name := columns[0]
		userRatingString := columns[1]
		imdbRatingString := columns[2]
		rottenCriticsString := columns[3]
		rottenUsersString := columns[4]
		metacriticCriticsString := columns[5]
		metacriticUsersString := columns[6]
		userRating, err := commons.ParseFloat(userRatingString)
		if err != nil {
			log.Fatalf("Failed to parse user rating of movie %s: %s", name, userRatingString)
		}
		imdbRating, err := commons.ParseFloat(imdbRatingString)
		if err != nil {
			log.Fatalf("Failed to parse IMDB rating of movie %s: %s", name, imdbRatingString)
		}
		rottenCritics, err := commons.ParseFloat(rottenCriticsString)
		if err != nil {
			log.Fatalf("Failed to parse Rotten Tomatoes critics' rating of movie %s: %s", name, rottenCriticsString)
		}
		rottenUsers, err := commons.ParseFloat(rottenUsersString)
		if err != nil {
			log.Fatalf("Failed to parse Rotten Tomatoes users' rating of movie %s: %s", name, rottenUsersString)
		}
		metacriticCritics, err := commons.ParseFloat(metacriticCriticsString)
		if err != nil {
			log.Fatalf("Failed to parse Metacritic critics' rating of movie %s: %s", name, metacriticCriticsString)
		}
		metacriticUsers, err := commons.ParseFloat(metacriticUsersString)
		if err != nil {
			log.Fatalf("Failed to parse Metacritic users' rating of movie %s: %s", name, metacriticUsersString)
		}
		userRating /= 100.0
		imdbRating /= 10.0
		rottenCritics /= 100.0
		rottenUsers /= 100.0
		metacriticCritics /= 100.0
		metacriticUsers /= 10.0
		userRatings.add(userRating, name)
		imdbRatings.add(imdbRating, name)
		rottenCriticsRatings.add(rottenCritics, name)
		rottenUsersRatings.add(rottenUsers, name)
		metacriticCriticsRatings.add(metacriticCritics, name)
		metacriticUsersRatings.add(metacriticUsers, name)
	})
	ratings := []movieRatings{
		imdbRatings,
		rottenCriticsRatings,
		rottenUsersRatings,
		metacriticCriticsRatings,
		metacriticUsersRatings,
	}
	return userRatings, ratings
}

func newMovieRatings(name string) movieRatings {
	return movieRatings{
		name: name,
	}
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
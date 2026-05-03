package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Fixture struct {
	Postcode string `json:"postcode"`
	Recipe   string `json:"recipe"`
	Delivery string `json:"delivery"`
}

type RecipeCount struct {
	Recipe string `json:"recipe"`
	Count  int    `json:"count"`
}

type BusiestPostcode struct {
	PostCode      string `json:"postcode"`
	DeliveryCount int    `json:"delivery_count"`
}

type Output struct {
	UniqueRecipeCount       int               `json:"unique_recipe_count"`
	CountPerRecipe          []RecipeCount     `json:"count_per_recipe"`
	MaxDeliveries           BusiestPostcode   `json:"busiest_postcode"`
	CountPerPostcodeAndTime PostcodeTimeCount `json:"count_per_postcode_and_time"`
	MatchByName             []string          `json:"match_by_name"`
}

type PostcodeTimeCount struct {
	Postcode      string `json:"postcode"`
	From          string `json:"from"`
	To            string `json:"to"`
	DeliveryCount int    `json:"delivery_count"`
}

func parseHour(s string) int {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "AM") {
		h, _ := strconv.Atoi(strings.TrimSuffix(s, "AM"))
		if h == 12 {
			return 0 // midnight
		}
		return h
	}
	h, _ := strconv.Atoi(strings.TrimSuffix(s, "PM"))
	if h == 12 {
		return 12 // noon
	}
	return h + 12
}

func main() {
	filePath := flag.String("file", "", "file path to read JSON file ")
	postcode := flag.String("postcode", "", "postcode to filter delivery")
	from := flag.String("from", "", "delivery window start eg 10 AM")
	to := flag.String("to", "", "delivery window end eg 3 PM")
	recipeName := flag.String("recipe", "", "partial recipe name to search")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("-- file is required")
	}

	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("--file is required")
	}

	defer file.Close()

	recipeCounts := make(map[string]int)   //to get unique recipe counts and count per recipe from json
	postcodeCounts := make(map[string]int) // to get number of deliveries and postcode from json
	partialRecipe := make(map[string]bool) //to match the partial input recipe from input
	postcodeTimeCount := 0

	filterByTime := *postcode != "" && *from != "" && *to != ""
	fromHour, toHour := 0, 0
	if filterByTime {
		fromHour = parseHour(*from)
		toHour = parseHour(*to)
	}

	decoder := json.NewDecoder(file)
	t, err := decoder.Token()
	if err != nil {
		log.Fatalf("could not read token %v", err)
	}
	_ = t //discard '[' from json file

	// data, err := os.ReadFile(*filePath) //first tried with json unmarshall to read the file, but it failed due to truncated file i think. So used NewDecode function to read from json file
	// if err != nil {
	// 	log.Fatalf("could not read file %v", err)
	// }
	// var fixtures []Fixture
	// if err := json.Unmarshal(data, &fixtures); err != nil {
	// 	log.Fatalf("could not parse json %v", err)
	// }

	for { //it runs forever till EOF or bad data is encountered
		var fixture Fixture
		err := decoder.Decode(&fixture)
		if err == nil {

		} else if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		} else {
			log.Printf("stopping at bad data %v", err)
			break
		}
		recipeCounts[fixture.Recipe]++ //loop through and get the unique counts for each recipe, and delivery counts for postcode
		postcodeCounts[fixture.Delivery]++

		parts := strings.Fields(fixture.Delivery)
		if len(parts) == 4 {
			delFrom := parseHour(parts[1])
			delTo := parseHour(parts[3])
			if delFrom < toHour && delTo > fromHour {
				postcodeTimeCount++
			}
		}

		if *recipeName != "" {
			if strings.Contains(strings.ToLower(fixture.Recipe), strings.ToLower(*recipeName)) {
				partialRecipe[fixture.Recipe] = true
			}
		}

	}

	busiestPostCode := ""
	busiestCount := 0
	for postcode, count := range postcodeCounts { //find busiest delivery count and its postcode
		if count > busiestCount {
			busiestCount = count
			busiestPostCode = postcode
		}
	}

	var countPerRecipe []RecipeCount
	for recipe, count := range recipeCounts {
		countPerRecipe = append(countPerRecipe, RecipeCount{Recipe: recipe, Count: count})
	}

	sort.Slice(countPerRecipe, func(i, j int) bool {
		return countPerRecipe[i].Recipe < countPerRecipe[j].Recipe //sorted count per recipe
	})

	var matches []string
	for name := range partialRecipe {
		matches = append(matches, name)
	}

	sort.Strings(matches) //use sort directly on string

	output := Output{ //populate the output struct with results
		UniqueRecipeCount: len(recipeCounts),
		CountPerRecipe:    countPerRecipe,
		MaxDeliveries: BusiestPostcode{
			PostCode:      busiestPostCode,
			DeliveryCount: busiestCount,
		},
		CountPerPostcodeAndTime: PostcodeTimeCount{
			Postcode:      *postcode,
			From:          *from,
			To:            *to,
			DeliveryCount: postcodeTimeCount,
		},
		MatchByName: matches,
	}

	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", " ") //to display in pretty print format
	if err := out.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding output %v", err)
		os.Exit(1)
	}

}

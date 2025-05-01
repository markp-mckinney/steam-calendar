package main

import (
	"fmt"
	"log"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

func main() {
	outDir := getEnvVar("OUT")
	steamID := getEnvVar("STEAMID")

	wishlistItems, err := getWishlistItems(steamID, outDir)
	if err != nil {
		log.Fatalf("Failed to get wishlist items: %v", err)
	}
	if len(wishlistItems) == 0 {
		log.Fatalf("No wishlist items returned")
	}

	icsString := createIcs(wishlistItems)
	err = writeFile(outDir, "wishlist.ics", []byte(icsString))
	if err != nil {
		log.Fatalf("Failed to create ICS file: %v", err)
	}

	upcoming := createUpcomingJson(wishlistItems)
	err = writeJSONFile(outDir, "upcoming.json", upcoming)
	if err != nil {
		log.Fatalf("Failed to create upcoming JSON file: %v", err)
	}
}

func getWishlistItems(steamID, outDir string) ([]StoreItem, error) {
	wishlistIDs, err := getWishlistIDs(steamID)
	if err != nil {
		return nil, err
	}

	if len(wishlistIDs) == 0 {
		return []StoreItem{}, nil
	}

	var unfilteredItems []StoreItem
	var filteredItems []StoreItem
	batchNum := 0

	for batch := range slices.Chunk(wishlistIDs, 200) {
		batchNum++
		log.Printf("Fetching %d store items in batch %d", len(batch), batchNum)
		res, err := getItems(batch)
		if err != nil {
			return nil, err
		}

		log.Printf("Got data for %d items", len(res.Response.StoreItems))

		for _, item := range res.Response.StoreItems {
			if item.AppID != 0 && item.Visible {
				filteredItems = append(filteredItems, item)
			}
			unfilteredItems = append(unfilteredItems, item)
		}
	}

	log.Printf("Total unfiltered items: %d", len(unfilteredItems))
	log.Printf("Total filtered items: %d", len(filteredItems))

	err = writeJSONFile(outDir, "unfiltered.json", unfilteredItems)
	if err != nil {
		log.Print(err)
	}

	err = writeJSONFile(outDir, "filtered.json", filteredItems)
	if err != nil {
		log.Print(err)
	}

	return filteredItems, nil
}

func createIcs(storeItems []StoreItem) string {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)

	for _, item := range storeItems {
		if item.Release.SteamReleaseDate > 0 {
			event := cal.AddEvent(strconv.Itoa(item.AppID))
			event.SetSummary(item.Name)
			event.SetStartAt(time.Unix(int64(item.Release.SteamReleaseDate), 0))
			event.SetDescription(item.BasicInfo.ShortDescription)
			event.SetURL(fmt.Sprintf("https://store.steampowered.com/%s", item.StoreURLPath))
		}
	}

	return cal.Serialize()
}

type Upcoming struct {
	Items []UpcomingItem `json:"items"`
}

type UpcomingItem struct {
	Name         string `json:"name"`
	ReleaseDate  string `json:"releaseDate"` // Formatted date or custom string
	StoreURLPath string `json:"storeUrlPath"`
	RawDate      int    `json:"-"` // Used for sorting, ignored in JSON output
}

func createUpcomingJson(storeItems []StoreItem) Upcoming {
	upcomingItems := []UpcomingItem{}
	now := time.Now()

	for _, item := range storeItems {
		upcomingItem := UpcomingItem{
			Name:         item.Name,
			StoreURLPath: item.StoreURLPath,
		}
		if item.Release.SteamReleaseDate > 0 {
			itemDate := time.Unix(int64(item.Release.SteamReleaseDate), 0)
			if itemDate.After(now) {
				upcomingItem.ReleaseDate = itemDate.Format(time.RFC3339)
				upcomingItem.RawDate = item.Release.SteamReleaseDate
				upcomingItems = append(upcomingItems, upcomingItem)
			}
		} else {
			upcomingItem.ReleaseDate = item.Release.CustomReleaseDateMessage
			upcomingItem.RawDate = math.MaxInt
			upcomingItems = append(upcomingItems, upcomingItem)
		}
	}

	log.Printf("Got %d upcoming items", len(upcomingItems))

	slices.SortFunc(upcomingItems, func(a, b UpcomingItem) int {
		rawDateDiff := a.RawDate - b.RawDate
		if rawDateDiff != 0 {
			return rawDateDiff
		}

		return strings.Compare(a.Name, b.Name)
	})

	return Upcoming{upcomingItems}
}

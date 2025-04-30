package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WishlistItem struct {
	AppID int `json:"appid"`
}

type GetWishlistResponse struct {
	Response struct {
		Items []WishlistItem `json:"items"`
	}
}

// getWishlistIDs returns IDs for a given steamID using IWishlistService/GetWishlist/v1.
func getWishlistIDs(steamID string) ([]WishlistItem, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.steampowered.com/IWishlistService/GetWishlist/v1/?steamid=%s", steamID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wishlist for steamID %q: %w", steamID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetWishlist API returned non-200 status: %d", resp.StatusCode)
	}

	var res GetWishlistResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return res.Response.Items, nil
}

type StoreItem struct {
	AppID        int       `json:"appid"`
	Name         string    `json:"name"`
	Visible      bool      `json:"visible"`
	StoreURLPath string    `json:"store_url_path"`
	Release      Release   `json:"release"`
	BasicInfo    BasicInfo `json:"basic_info"`
}

type Release struct {
	SteamReleaseDate         int    `json:"steam_release_date"` // Unix timestamp
	IsComingSoon             bool   `json:"is_coming_soon"`
	CustomReleaseDateMessage string `json:"custom_release_date_message"`
}

type BasicInfo struct {
	ShortDescription string `json:"short_description"`
}

type Context struct {
	Language    string `json:"language"`
	CountryCode string `json:"country_code"`
	SteamRealm  int    `json:"steam_realm"`
}

type DataRequest struct {
	IncludeRelease   bool `json:"include_release"`
	IncludeBasicInfo bool `json:"include_basic_info"`
}

type GetItemsInput struct {
	IDs         []WishlistItem `json:"ids"`
	Context     Context        `json:"context"`
	DataRequest DataRequest    `json:"data_request"`
}

type GetItemsResponse struct {
	Response struct {
		StoreItems []StoreItem `json:"store_items"`
	} `json:"response"`
}

func getItems(wishlistItems []WishlistItem) (GetItemsResponse, error) {
	input := GetItemsInput{
		IDs: wishlistItems,
		Context: Context{
			Language:    "english",
			CountryCode: "US",
			SteamRealm:  1,
		},
		DataRequest: DataRequest{
			IncludeRelease:   true,
			IncludeBasicInfo: true,
		},
	}

	inputJson, err := json.Marshal(input)
	if err != nil {
		return GetItemsResponse{}, fmt.Errorf("failed to marshal GetItems input: %w", err)
	}

	resp, err := http.Get(fmt.Sprintf("https://api.steampowered.com/IStoreBrowseService/GetItems/v1?input_json=%s", inputJson))
	if err != nil {
		return GetItemsResponse{}, fmt.Errorf("failed to fetch items: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GetItemsResponse{}, fmt.Errorf("GetItems API returned non-200 status: %d", resp.StatusCode)
	}

	var res GetItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return GetItemsResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return res, nil
}

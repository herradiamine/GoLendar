package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LocationData représente les données de localisation retournées par ipapi.co
type LocationData struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

// GetLocationFromIP récupère la localisation géographique à partir d'une adresse IP
func GetLocationFromIP(ip string) string {
	// Ignorer les IPs locales
	if ip == "127.0.0.1" || ip == "localhost" || ip == "::1" {
		return "Local"
	}

	// Créer un client HTTP avec timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Appeler l'API ipapi.co
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	resp, err := client.Get(url)
	if err != nil {
		return ip // fallback sur l'IP en cas d'erreur
	}
	defer resp.Body.Close()

	// Lire la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ip
	}

	// Parser la réponse JSON
	var locationData LocationData
	if err := json.Unmarshal(body, &locationData); err != nil {
		return ip
	}

	// Vérifier si la requête a réussi
	if locationData.Status != "success" {
		return ip
	}

	// Construire la chaîne de localisation
	var location string
	if locationData.City != "" {
		location = locationData.City
		if locationData.RegionName != "" {
			location += ", " + locationData.RegionName
		}
		if locationData.Country != "" {
			location += ", " + locationData.Country
		}
	} else if locationData.RegionName != "" {
		location = locationData.RegionName
		if locationData.Country != "" {
			location += ", " + locationData.Country
		}
	} else if locationData.Country != "" {
		location = locationData.Country
	} else {
		location = ip // fallback si aucune info géographique
	}

	return location
}

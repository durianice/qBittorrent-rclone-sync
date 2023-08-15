package main

import (
	"fmt"
	"qbittorrentRcloneSync/http"
)

func main() {
	http.Login()
	list := http.GetInfo()

	for _, obj := range list {
		name, nameExists := obj["name"].(string)
		hash, hashExists := obj["hash"].(string)

		if nameExists {
			fmt.Println("Name:", name)
		}
		if hashExists {
			fmt.Println("Hash:", hash)
		}

		subList := http.GetDetail(hash)

		for _, subObj := range subList {
			name, nameExists := subObj["name"].(string)
			size, sizeExists := subObj["size"].(int)
			progress, progressExists := subObj["progress"].(float64)

			if nameExists {
				fmt.Println("Name:", name)
			}
			if sizeExists {
				fmt.Println("Size:", size)
			}
			if progressExists {
				fmt.Println("Progress:", progress)
			}
		}
	}
}
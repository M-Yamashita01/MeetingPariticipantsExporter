package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func main() {
	var accessToken string
	var meetingUUID string
	fmt.Println("Input access token")
	fmt.Scan(&accessToken)
	fmt.Println("Input meetingId")
	fmt.Scan(&meetingUUID)

	meetingParticipantsUrl := "https://api.zoom.us/v2/past_meetings/"
	url := meetingParticipantsUrl + meetingUUID + "/participants"
	req, _ := http.NewRequest("GET", url, nil)
	bearerAccessToken := "Bearer " + accessToken
	req.Header.Set("Authorization", bearerAccessToken)

	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("%s", dump)

	client := new(http.Client)
	resp, _ := client.Do(req)

	dumpResp, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("%s", dumpResp)
}

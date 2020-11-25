package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/kelseyhightower/envconfig"
)

type Env struct {
	SecretKey string `required:"true"`
	ApiKey    string `required:"true"`
}

type MeetingParticipantsResponseBody struct {
	PageCount     int                        `json: page_count`
	PageSize      int                        `json: page_size`
	TotalRecord   int                        `json: total_records `
	NextPageToken string                     `json: next_page_token`
	Participants  []ParticipantsResponseBody `json: participants`
}

type ParticipantsResponseBody struct {
	Id        string `json: id`
	Name      string `json: name`
	UserEmail string `json: user_email`
}

func main() {
	var goenv Env
	if err := envconfig.Process("", &goenv); err != nil {
		log.Printf("ERROR Environment is not set: %s", err)
		return
	}
	fmt.Println(goenv)

	// Generate access token by JWT
	key := []byte(goenv.SecretKey)
	signer, _ := jwt.NewSignerHS(jwt.HS256, key)

	now := time.Now()
	after := now.Add(time.Minute)
	claims := jwt.StandardClaims{
		Issuer:    goenv.ApiKey,
		ExpiresAt: jwt.NewNumericDate(after),
	}

	builder := jwt.NewBuilder(signer)
	token, _ := builder.Build(claims)
	fmt.Printf("Access token : %s\n", token)

	var tokenStr string
	var meetingUUID string
	tokenStr = token.String()
	fmt.Println("Input meetingId")
	fmt.Scan(&meetingUUID)

	meetingParticipantsUrl := "https://api.zoom.us/v2/past_meetings/"
	url := meetingParticipantsUrl + meetingUUID + "/participants"
	req, _ := http.NewRequest("GET", url, nil)
	bearerAccessToken := "Bearer " + tokenStr
	req.Header.Set("Authorization", bearerAccessToken)

	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("%s", dump)

	client := new(http.Client)
	resp, _ := client.Do(req)

	dumpResp, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("%s\n", dumpResp)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Response body: ", string(body))

	//Unmarshall json
	var meetingParticipantsResponseBody MeetingParticipantsResponseBody
	if err := json.Unmarshal(body, &meetingParticipantsResponseBody); err != nil {
		log.Fatal(err)
		// ("ERROR Environment is not set: %s", err)
		return
	}

	fmt.Printf("pageCount: %v, pageSize: %v", meetingParticipantsResponseBody.PageCount, meetingParticipantsResponseBody.PageSize)

}

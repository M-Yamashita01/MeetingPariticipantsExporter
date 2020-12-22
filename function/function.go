package function

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/kelseyhightower/envconfig"
)

type Env struct {
	SecretKey string `required:"true"`
	ApiKey    string `required:"true"`
}

type MeetingParticipantsResponseBody struct {
	Pagecount     int                        `json:"page_count"`
	PageSize      int                        `json:"page_size"`
	TotalRecord   int                        `json:"total_records"`
	NextPageToken string                     `json:"next_page_token"`
	Participants  []ParticipantsResponseBody `json:"participants"`
}

type ParticipantsResponseBody struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	UserEmail string `json:"user_email"`
}

func ExportParticipants(w http.ResponseWriter, r *http.Request) {

	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	if signingSecret == "" {
		fmt.Println("ERROR Environment for Slack API missing.")
		return
	}

	slackTimeStamp := r.Header.Get("X-Slack-Request-Timestamp")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("unable to read slack data in message body", err)
		http.Error(w, "Failed to process request", http.StatusBadRequest)
		return
	}
	slackSigningBaseString := "v0:" + slackTimeStamp + ":" + string(b)
	slackSignature := r.Header.Get("X-Slack-Signature")

	if !matchSignature(slackSignature, signingSecret, slackSigningBaseString) {
		fmt.Println("Signature did not match!")
		http.Error(w, "Function was not invoked by Slack", http.StatusForbidden)
		return
	}

	fmt.Println("Slack request verified successfully")

	//parse the application/x-www-form-urlencoded data sent by Slack
	vals, err := parse(b)
	if err != nil {
		fmt.Println("unable to parse data sent by slack", err)
		http.Error(w, "Failed to process request", http.StatusBadRequest)
		return
	}
	meetingUUID := vals.Get("text")
	fmt.Println("Meeting UUID:", meetingUUID)

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
	tokenStr = token.String()

	meetingParticipantsUrl := "https://api.zoom.us/v2/past_meetings/"
	url := meetingParticipantsUrl + meetingUUID + "/participants"
	req, _ := http.NewRequest("GET", url, nil)
	bearerAccessToken := "Bearer " + tokenStr
	req.Header.Set("Authorization", bearerAccessToken)

	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("DumpRequest: %s", dump)

	client := new(http.Client)
	resp, _ := client.Do(req)

	dumpResp, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("DumpResponse: %s\n", dumpResp)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("Response body: %s", body)
	if err != nil {
		fmt.Printf("Error ocurred while executing ReadAll. %s", err)
		log.Fatal(err)
		return
	}

	//Unmarshall json
	var meetingParticipantsResponseBody MeetingParticipantsResponseBody
	if err := json.Unmarshal(body, &meetingParticipantsResponseBody); err != nil {
		fmt.Printf("Error ocurred while unmarshal body. %s", err)
		log.Fatal(err)
		return
	}

	fmt.Println("Participant")
	participantNames := ""
	for _, participant := range meetingParticipantsResponseBody.Participants {
		fmt.Printf("%s\n", participant.Name)
		participantNames += participant.Name
		participantNames += "\n"
	}

	slackResponse := SlackResponse{Text: participantNames}

	//slack needs the content-type to be set explicitly - https://api.slack.com/slash-commands#responding_immediate_response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slackResponse)
	fmt.Println("Sent response to Slack")
}

func matchSignature(slackSignature, signingSecret, slackSigningBaseString string) bool {

	//calculate SHA256 of the slackSigningBaseString using signingSecret
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(slackSigningBaseString))

	//hex encode the SHA256
	calculatedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	match := hmac.Equal([]byte(slackSignature), []byte(calculatedSignature))
	return match
}

//adapted from from net/http/request.go --> func parsePostForm(r *Request) (vs url.Values, err error)
func parse(b []byte) (url.Values, error) {
	vals, e := url.ParseQuery(string(b))
	if e != nil {
		fmt.Println("unable to parse", e)
		return nil, e
	}
	return vals, nil
}

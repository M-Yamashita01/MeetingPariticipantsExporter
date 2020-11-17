package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yaegashi/msgraph.go/jsonx"
	msauth "github.com/yaegashi/msgraph.go/msauth"
	msgraph "github.com/yaegashi/msgraph.go/v1.0"
	"golang.org/x/oauth2"
)

const (
	defaultTenantID       = "common"
	defaultClientID       = "45c7f99c-0a94-42ff-a6d8-a8d657229e8c"
	defaultTokenCachePath = "token_cache.json"
)

var defaultScopes = []string{"offline_access", "User.Read", "Calendars.Read", "Files.Read", "Group.Read.All", "Team.ReadBasic.All"}

func dump(o interface{}) {
	enc := jsonx.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(o)
}

func main() {
	// var string
	// var filePath string
	fmt.Println("Hello, world!")
	var tenantID, clientID, tokenCachePath string

	flag.StringVar(&tenantID, "tenant-id", defaultTenantID, "TenantID")
	flag.StringVar(&clientID, "client-id", defaultClientID, "ClientID")
	flag.StringVar(&tokenCachePath, "token-cache-path", defaultTokenCachePath, "TokenCachePath")
	flag.Parse()

	ctx := context.Background()
	m := msauth.NewManager()
	m.LoadFile(tokenCachePath)
	ts, err := m.DeviceAuthorizationGrant(ctx, tenantID, clientID, defaultScopes, nil)
	if err != nil {
		log.Fatal(err)
	}
	m.SaveFile(tokenCachePath)

	// fmt.Printf(ts.Token().)

	httpClient := oauth2.NewClient(ctx, ts)
	graphClient := msgraph.NewClient(httpClient)
	{
		log.Printf("Get current logged in user information.")
		req := graphClient.Me().Request()
		log.Printf("Get %s", req.URL())
		user, err := req.Get(ctx)
		if err == nil {
			dump(user)
		} else {
			log.Println(err)
		}
	}

	{
		log.Printf("Get joined team.")
		req := graphClient.Me().JoinedTeams().Request()
		log.Printf("Get %s", req.URL())
		user, err := req.Get(ctx)
		if err == nil {
			dump(user)
		} else {
			log.Println(err)
		}
	}
}

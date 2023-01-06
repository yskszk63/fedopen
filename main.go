package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

type federationRequest struct{
	SessionId *string `json:"sessionId"`
	SessionKey *string `json:"sessionKey"`
	SessionToken *string `json:"sessionToken"`
}

type federationResponse struct {
	SigninToken string
}

func main() {
	cx := context.Background()

	cfg, err := config.LoadDefaultConfig(cx)
	if err != nil {
		log.Fatal(err)
	}

	name := "suzukixxx"

	svc := sts.NewFromConfig(cfg)

	policyArn := "arn:aws:iam::aws:policy/AdministratorAccess"
	resp, err := svc.GetFederationToken(cx, &sts.GetFederationTokenInput{
		Name: &name,
		PolicyArns: []types.PolicyDescriptorType{
			{ Arn: &policyArn },
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	req := federationRequest{
		SessionId: resp.Credentials.AccessKeyId,
		SessionKey: resp.Credentials.SecretAccessKey,
		SessionToken: resp.Credentials.SessionToken,
	}
	tmpCred, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}

	q := url.Values{}
	q.Add("Action", "getSigninToken")
	q.Add("Session", string(tmpCred))
	u := fmt.Sprintf("https://signin.aws.amazon.com/federation?%s", q.Encode())

	r, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		io.Copy(os.Stdout, r.Body)
		log.Fatal(r.Status)
	}

	var res federationResponse
	dec := json.NewDecoder(r.Body)
	dec.Decode(&res)

	q = url.Values{}
	q.Add("Action", "login")
	q.Add("Issuer", "http://example.com")
	q.Add("Destination", "https://console.aws.amazon.com/")
	q.Add("SigninToken", res.SigninToken)
	u = fmt.Sprintf("https://signin.aws.amazon.com/federation?%s", q.Encode())
	fmt.Printf("%s\n", u)
}

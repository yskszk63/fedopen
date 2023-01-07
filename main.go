package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

type getSigninTokenRequest struct{
	SessionId *string `json:"sessionId"`
	SessionKey *string `json:"sessionKey"`
	SessionToken *string `json:"sessionToken"`
}

type getSigninTokenResponse struct {
	SigninToken string
}

func decideName(cx context.Context, client *iam.Client, defaultName string) string {
	resp, err := client.GetUser(cx, &iam.GetUserInput{})
	if err != nil || resp.User == nil || resp.User.UserName == nil {
		return defaultName
	}

	return *resp.User.UserName
}

func getFederationToken(cx context.Context, client *sts.Client, name, policyArn string) (*types.Credentials, error) {
	resp, err := client.GetFederationToken(cx, &sts.GetFederationTokenInput{
		Name: &name,
		PolicyArns: []types.PolicyDescriptorType{
			{ Arn: &policyArn },
		},
	})
	if err != nil {
		return nil, err
	}
	if resp.Credentials == nil {
		return nil, errors.New("resp.Credentials is nil")
	}
	return resp.Credentials, nil
}

func getSigninToken(cx context.Context, cred *types.Credentials) (string, error) {
	req := getSigninTokenRequest{
		SessionId: cred.AccessKeyId,
		SessionKey: cred.SecretAccessKey,
		SessionToken: cred.SessionToken,
	}
	tmpCred, err := json.Marshal(req)
	if err != nil {
		return "", nil
	}

	q := url.Values{}
	q.Add("Action", "getSigninToken")
	q.Add("Session", string(tmpCred))
	u := fmt.Sprintf("https://signin.aws.amazon.com/federation?%s", q.Encode())

	r, err := http.Get(u)
	if err != nil {
		return "", nil
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			b = []byte{}
		}
		return "", errors.New(string(b))
	}

	var res getSigninTokenResponse
	dec := json.NewDecoder(r.Body)
	dec.Decode(&res)

	return res.SigninToken, nil
}

func genLoginUrl(token string) string {
	q := url.Values{}
	q.Add("Action", "login")
	q.Add("Issuer", "fedopen")
	q.Add("Destination", "https://console.aws.amazon.com/")
	q.Add("SigninToken", token)
	return fmt.Sprintf("https://signin.aws.amazon.com/federation?%s", q.Encode())
}

func openOrPrint(url string) {
	p, err := exec.LookPath("xdg-open")
	if err == nil {
		if err := exec.Command(p, url).Run(); err == nil {
			return
		}
	}

	fmt.Printf("%s\n", url)
}

func main() {
	cx := context.Background()

	cfg, err := config.LoadDefaultConfig(cx)
	if err != nil {
		log.Fatal(err)
	}

	iamc := iam.NewFromConfig(cfg)
	stsc := sts.NewFromConfig(cfg)

	name := decideName(cx, iamc, "fedopen")
	policyArn := "arn:aws:iam::aws:policy/AdministratorAccess"

	cred, err := getFederationToken(cx, stsc, name, policyArn)
	if err != nil {
		log.Fatal(err)
	}

	token, err := getSigninToken(cx, cred)
	if err != nil {
		log.Fatal(err)
	}

	url := genLoginUrl(token)
	openOrPrint(url)
}

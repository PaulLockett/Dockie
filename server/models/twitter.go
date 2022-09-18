package models

import (
	"fmt"
	"net/http"
)

type Authorize struct {
	Token string
}

func (a Authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

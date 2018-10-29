package controllers

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

func NewURL(host string, username string, password string) string {
	url := fmt.Sprintf("https://%s:%s@%s"+vim25.Path, username, password, host)
	return url

}

func NewClient(ctx context.Context, url string) (*govmomi.Client, error) {

	u, err := soap.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return govmomi.NewClient(ctx, u, true)
}

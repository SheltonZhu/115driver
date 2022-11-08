package driver

import (
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/pkg/errors"
)

// LoginCheck check login status
func (c *Pan115Client) LoginCheck() error {
	result := LoginResp{}
	req := c.NewRequest().
		SetQueryParam("_", NowMilli().String()).
		SetResult(&result)
	resp, err := req.Get(ApiLoginCheck)
	if err = CheckErr(err, &result, resp); err != nil {
		return err
	}
	c.UserID = result.Data.UserID
	return nil
}

// ImportCredential import uid, cid, seid
func (c *Pan115Client) ImportCredential(cr *Credential) *Pan115Client {
	cookies := map[string]string{
		CookieNameUid:  cr.UID,
		CookieNameCid:  cr.CID,
		CookieNameSeid: cr.SEID,
	}
	c.ImportCookies(cookies, CookieDomain115)
	return c
}

func (c *Pan115Client) ImportCookies(cookies map[string]string, domains ...string) {
	for _, domain := range domains {
		c.importCookies(cookies, domain, "/")
	}
}

func (c *Pan115Client) importCookies(cookies map[string]string, domain string, path string) {
	// Make a dummy URL for saving cookie
	url := &neturl.URL{
		Scheme: "https",
		Path:   "/",
	}
	if domain[0] == '.' {
		url.Host = "www" + domain
	} else {
		url.Host = domain
	}
	// Prepare cookies
	cks := make([]*http.Cookie, 0, len(cookies))
	for name, value := range cookies {
		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Domain:   domain,
			Path:     path,
			HttpOnly: true,
		}
		cks = append(cks, cookie)
	}
	// Save cookies
	c.SetCookies(cks...)
}

type Credential struct {
	UID  string `json:"UID"`
	CID  string `json:"CID"`
	SEID string `json:"SEID"`
}

// FromCookie get uid, cid, seid from cookie string
func (cr *Credential) FromCookie(cookie string) error {
	items := strings.Split(cookie, ";")
	if len(items) < 3 {
		return errors.Wrap(ErrBadCookie, "number of cookie paris < 3")
	}

	cookieMap := map[string]string{}
	for _, item := range items {
		pairs := strings.Split(strings.TrimSpace(item), "=")
		if len(pairs) != 2 {
			return ErrBadCookie
		}
		key := pairs[0]
		value := pairs[1]
		cookieMap[strings.ToUpper(key)] = value
	}
	cr.UID = cookieMap["UID"]
	cr.CID = cookieMap["CID"]
	cr.SEID = cookieMap["SEID"]
	if cr.CID == "" || cr.UID == "" || cr.SEID == "" {
		return errors.Wrap(ErrBadCookie, "bad cookie, miss UID, CID or SEID")
	}
	return nil
}

// GetUser get user information
func (c *Pan115Client) GetUser() (*UserInfo, error) {
	result := UserInfoResp{}
	req := c.NewRequest().
		SetQueryParam("_", Now().String()).
		SetResult(&result)
	resp, err := req.Get(ApiUserInfo)
	return &result.UserInfo, CheckErr(err, &result, resp)
}

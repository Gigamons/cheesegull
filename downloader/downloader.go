// Package downloader implements downloading from the osu! website, through,
// well, mostly scraping and dirty hacks.
package downloader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

var downloadHostname = "osu.ppy.sh"

// LogIn logs in into an osu! account and returns a Client.
func LogIn(username, password string, downloadhostname string) (*Client, error) {
	logger.Debug("Try to Login into Osu!")
	j, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Jar: j,
	}
	vals := url.Values{}
	vals.Add("redirect", "/")
	vals.Add("sid", "")
	vals.Add("username", username)
	vals.Add("password", password)
	vals.Add("autologin", "on")
	vals.Add("login", "login")
	loginResp, err := c.PostForm("https://osu.ppy.sh/forum/ucp.php?mode=login", vals)
	if err != nil {
		return nil, err
	}
	if loginResp.Request.URL.Path != "/home" && loginResp.Request.URL.Path != "/" {
		return nil, errors.New("cheesegull/downloader: could not log in (was not redirected to index)")
	}

	downloadHostname = downloadhostname
	if downloadhostname == "" {
		logger.Debug("WARNING! set downloadHostname to Default osu.ppy.sh")
		downloadHostname = "osu.ppy.sh"
	}

	return (*Client)(c), nil
}

// Client is a wrapper around an http.Client which can fetch beatmaps from the
// osu! website.
type Client http.Client

// HasVideo checks whether a beatmap has a video.
func (c *Client) HasVideo(setID int) (bool, error) {
	logger.Debug("Check if SetID %v has Video.", setID)
	h := (*http.Client)(c)

	page, err := h.Get(fmt.Sprintf("https://osu.ppy.sh/s/%d", setID))
	if err != nil {
		logger.Debug("SetID %v don't has video!", setID)
		return false, err
	}
	defer page.Body.Close()
	body, err := ioutil.ReadAll(page.Body)
	if err != nil {
		logger.Debug("SetID %v has video!", setID)
		return false, err
	}
	return bytes.Contains(body, []byte(fmt.Sprintf(`href="/d/%dn"`, setID))), nil
}

// Download downloads a beatmap from the osu! website. noVideo specifies whether
// we should request the beatmap to not have the video.
func (c *Client) Download(setID int, noVideo bool) (io.ReadCloser, error) {
	suffix := ""
	if noVideo {
		suffix = "n"
	}

	logger.Debug("Download SetID %v. has video: %t", setID, noVideo)

	return c.getReader(strconv.Itoa(setID) + suffix)
}

// ErrNoRedirect is returned from Download when we were not redirect, thus
// indicating that the beatmap is unavailable.
var ErrNoRedirect = errors.New("cheesegull/downloader: no redirect happened, beatmap could not be downloaded")

var errNoZip = errors.New("cheesegull/downloader: file is not a zip archive")

const zipMagic = "PK\x03\x04"

func (c *Client) getReader(str string) (io.ReadCloser, error) {
	h := (*http.Client)(c)
	var uri string

	if downloadHostname == "osu.ppy.sh" {
		uri = fmt.Sprintf("https://%s/beatmapsets/%s/download", downloadHostname, str)
	} else {
		uri = fmt.Sprintf("https://%s/d/%s", downloadHostname, str)
	}

	resp, err := h.Get(uri)
	if err != nil {
		return nil, err
	}

	if resp.Request.URL.Host == "osu.ppy.sh" {
		resp.Body.Close()
		return nil, ErrNoRedirect
	}

	// check that it is a zip file
	first4 := make([]byte, 4)
	_, err = resp.Body.Read(first4)
	if err != nil {
	}
		return nil, err
	if string(first4) != zipMagic {
		return nil, errNoZip
	}

	return struct {
		io.Closer
		io.Reader
	}{
		io.MultiReader(strings.NewReader(zipMagic), resp.Body),
		resp.Body,
	}, nil
}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	site := os.Getenv("DD_SITE")
	if site == "" {
		site = "datadog.com"
	}
	var (
		keyF     = flag.String("key", os.Getenv("DD_API_KEY"), "A Datadog API key for your account. Defaults to DD_API_KEY.")
		serviceF = flag.String("service", "dd-prof-upload", "The name of the service to assign for the uploaded profiles.")
		siteF    = flag.String("site", site, `The datadog site to upload to. Defaults to DD_SITE or "datadog.com".`)
		envF     = flag.String("env", "dev", "The name of the environment to assign to the uploaded profiles.")
		runtimeF = flag.String("runtime", "go", "The name of the runtime to attribute the profiles to.")
	)
	flag.Parse()

	u := Upload{
		URL:    "https://intake.profile." + *siteF + "/v1/input",
		ApiKey: *keyF,
		Tags: []string{
			"service:" + *serviceF,
			"env:" + *envF,
			"runtime:" + *runtimeF,
		},
		Runtime: *runtimeF,
		Files:   flag.Args(),
	}
	return u.Upload()
}

type Upload struct {
	URL     string
	Runtime string
	ApiKey  string
	Tags    []string
	Files   []string
}

func (u *Upload) Upload() error {
	req, err := u.newRequest()
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("error: http %d: %s", res.StatusCode, body)
	}
	return nil
}

func (u *Upload) newRequest() (*http.Request, error) {
	var body bytes.Buffer
	var err error

	mw := multipart.NewWriter(&body)
	// write all of the profile metadata (including some useless ones)
	// with a small helper function that makes error tracking less verbose.
	writeField := func(k, v string) {
		if err == nil {
			err = mw.WriteField(k, v)
		}
	}
	writeField("version", "3")
	writeField("family", u.Runtime)
	writeField("start", time.Now().Format(time.RFC3339))
	// TODO(fg) is Add(time.Minute) the right thing to do here?
	writeField("end", time.Now().Add(time.Minute).Format(time.RFC3339))
	for _, tag := range u.Tags {
		writeField("tags[]", tag)
	}
	if err != nil {
		return nil, err
	}
	for _, path := range u.Files {
		formFile, err := mw.CreateFormFile(fmt.Sprintf("data[%s]", filepath.Base(path)), "pprof-data")
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if _, err := formFile.Write(data); err != nil {
			return nil, err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	// TODO(fg) use NewRequestWithContext once go 1.12 support is dropped.
	req, err := http.NewRequest("POST", u.URL, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("DD-API-KEY", u.ApiKey)
	return req, nil
}

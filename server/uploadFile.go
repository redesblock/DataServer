package server

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	PinHeader            = "Swarm-Pin"
	TagHeader            = "Swarm-Tag"
	EncryptHeader        = "Swarm-Encrypt"
	IndexDocumentHeader  = "Swarm-Index-Document"
	ErrorDocumentHeader  = "Swarm-Error-Document"
	FeedIndexHeader      = "Swarm-Feed-Index"
	FeedIndexNextHeader  = "Swarm-Feed-Index-Next"
	CollectionHeader     = "Swarm-Collection"
	PostageBatchIdHeader = "Swarm-Postage-Batch-Id"
	DeferredUploadHeader = "Swarm-Deferred-Upload"

	ContentTypeHeader = "Content-Type"
	MultiPartFormData = "multipart/form-data"
	ContentTypeTar    = "application/x-tar"
)

func node() string {
	if val, ok := os.LookupEnv("DATA_SERVER_MOP"); ok {
		return val
	}
	return "http://183.131.181.164:1683"
}

func voucher() string {
	if val, ok := os.LookupEnv("DATA_SERVER_VOUCHER"); ok {
		return val
	}
	return "3a10f61116ce8d85d2433b309c68365c95b2b37390b1d7bfc218418599275636"
}

func uploadFiles(url, batchID, assetID, name string) (string, error) {
	buf, filename, err := tarFiles(assetID, name)
	if err != nil {
		return "", err
	}

	url += "/mop"
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return "", fmt.Errorf("http request %v", err)
	}
	req.Header.Add(DeferredUploadHeader, "true")
	req.Header.Add(PostageBatchIdHeader, batchID)
	req.Header.Add(CollectionHeader, "true")
	req.Header.Add(ContentTypeHeader, ContentTypeTar)
	req.Header.Add(IndexDocumentHeader, filename)
	req.Header.Add(ErrorDocumentHeader, "")
	req.Header.Add(EncryptHeader, "false")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http do %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("http resp got status %s, want %s", resp.Status, http.StatusText(http.StatusCreated))
	}

	var ret map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return "", fmt.Errorf("http resp %v", err)
	}
	return ret["reference"], nil
}

func tarFiles(assetID, name string) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	tempFolder := "assets/" + assetID

	if s, err := os.Stat(filepath.Join(tempFolder, name)); err != nil {
		return nil, "", err
	} else if s.IsDir() {
		tempFolder = filepath.Join(tempFolder, name)
	}

	n := len(strings.Split(tempFolder, "/"))
	filename := ""
	filepath.Walk(tempFolder, func(path string, info fs.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed tar, read data %v", err)
		}
		if filepath.Join(strings.Split(path, "/")[n:]...) == info.Name() {
			filename = info.Name()
		}
		hdr := &tar.Header{
			Name: filepath.Join(strings.Split(path, "/")[n:]...),
			Mode: 0600,
			Size: info.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("failed tar, write header %v", err)
		}
		if _, err := tw.Write(data); err != nil {
			return fmt.Errorf("failed tar, write data %v", err)
		}
		return nil
	})
	// finally close the tar writer
	if err := tw.Close(); err != nil {
		return nil, filename, fmt.Errorf("failed tar, close %v", err)
	}
	return &buf, filename, nil
}

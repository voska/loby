package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
)

// buildBody marshals req.Body / req.Form / req.Files into a byte buffer plus
// Content-Type. Body wins; otherwise Files implies multipart; otherwise Form
// implies application/x-www-form-urlencoded; otherwise no body.
func buildBody(req *Request) ([]byte, string, error) {
	switch {
	case req.Body != nil:
		buf, err := json.Marshal(req.Body)
		if err != nil {
			return nil, "", fmt.Errorf("marshal body: %w", err)
		}
		return buf, "application/json", nil
	case len(req.Files) > 0:
		return multipartBody(req)
	case len(req.Form) > 0:
		return []byte(req.Form.Encode()), "application/x-www-form-urlencoded", nil
	default:
		return nil, "", nil
	}
}

func multipartBody(req *Request) ([]byte, string, error) {
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	for k, vs := range req.Form {
		for _, v := range vs {
			if err := mw.WriteField(k, v); err != nil {
				return nil, "", fmt.Errorf("multipart field %s: %w", k, err)
			}
		}
	}
	for _, fp := range req.Files {
		w, err := mw.CreateFormFile(fp.Field, fp.Filename)
		if err != nil {
			return nil, "", fmt.Errorf("multipart file %s: %w", fp.Field, err)
		}
		if _, err := io.Copy(w, fp.Reader); err != nil {
			return nil, "", fmt.Errorf("copy file %s: %w", fp.Field, err)
		}
	}
	if err := mw.Close(); err != nil {
		return nil, "", fmt.Errorf("close multipart: %w", err)
	}
	return buf.Bytes(), mw.FormDataContentType(), nil
}

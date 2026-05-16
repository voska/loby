package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
)

// MaxMultipartBytes caps in-memory multipart payloads. Multipart requests must
// be buffered so retry attempts can replay the body; this guard prevents an
// agent from accidentally OOM'ing the host by piping a multi-GB CSV.
const MaxMultipartBytes = 100 * 1024 * 1024

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
		name := filepath.Base(fp.Filename)
		w, err := mw.CreateFormFile(fp.Field, name)
		if err != nil {
			return nil, "", fmt.Errorf("multipart file %s: %w", fp.Field, err)
		}
		limited := &io.LimitedReader{R: fp.Reader, N: int64(MaxMultipartBytes) + 1 - int64(buf.Len())}
		if _, err := io.Copy(w, limited); err != nil {
			return nil, "", fmt.Errorf("copy file %s: %w", fp.Field, err)
		}
		if buf.Len() > MaxMultipartBytes {
			return nil, "", fmt.Errorf("multipart payload exceeds %d-byte cap (file %q)", MaxMultipartBytes, name)
		}
	}
	if err := mw.Close(); err != nil {
		return nil, "", fmt.Errorf("close multipart: %w", err)
	}
	return buf.Bytes(), mw.FormDataContentType(), nil
}

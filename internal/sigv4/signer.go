package sigv4

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type Signer struct {
	Creds   Credentials
	Region  string
	Service string
	Now     func() time.Time
}

// Sign signs an HTTP request using AWS Signature Version 4.
// It expects:
// - req.URL set
// - req.Body readable (if non-nil)
// It will set:
// - Authorization
// - x-amz-date
// - x-amz-security-token (if session token present)
func (s *Signer) Sign(req *http.Request) error {
	if s.Now == nil {
		s.Now = time.Now
	}
	t := s.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	payloadHash, body, err := hashPayload(req)
	if err != nil {
		return err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	// Required headers
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	if s.Creds.SessionToken != "" {
		req.Header.Set("x-amz-security-token", s.Creds.SessionToken)
	}

	canonicalRequest, signedHeaders := buildCanonicalRequest(req, payloadHash)
	crHash := sha256Hex([]byte(canonicalRequest))

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.Region, s.Service)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		crHash,
	}, "\n")

	signingKey := deriveSigningKey(s.Creds.SecretAccessKey, dateStamp, s.Region, s.Service)
	signature := hmacSHA256Hex(signingKey, []byte(stringToSign))

	authHeader := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.Creds.AccessKeyID,
		credentialScope,
		signedHeaders,
		signature,
	)

	req.Header.Set("Authorization", authHeader)
	return nil
}

func hashPayload(req *http.Request) (string, []byte, error) {
	if req.Body == nil {
		empty := sha256Hex([]byte{})
		return empty, []byte{}, nil
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return "", nil, err
	}
	return sha256Hex(b), b, nil
}

func buildCanonicalRequest(req *http.Request, payloadHash string) (string, string) {
	canonicalURI := sanitizePath(req.URL.Path)
	canonicalQuery := canonicalQueryString(req.URL)

	// Collect headers
	headers := make(map[string]string)
	for k, vv := range req.Header {
		lk := strings.ToLower(k)
		// Trim spaces and join multiple with comma
		vals := make([]string, 0, len(vv))
		for _, v := range vv {
			vals = append(vals, strings.TrimSpace(v))
		}
		headers[lk] = strings.Join(vals, ",")
	}

	// Host must be included
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	headers["host"] = host

	// Sort header keys
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var canonicalHeaders strings.Builder
	for _, k := range keys {
		canonicalHeaders.WriteString(k)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(normalizeHeaderValue(headers[k]))
		canonicalHeaders.WriteString("\n")
	}

	signedHeaders := strings.Join(keys, ";")

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")

	return canonicalRequest, signedHeaders
}

func sanitizePath(p string) string {
	if p == "" {
		return "/"
	}
	// AWS expects URI-encoded path, but keep "/" safe.
	segments := strings.Split(p, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	out := strings.Join(segments, "/")
	if !strings.HasPrefix(out, "/") {
		out = "/" + out
	}
	return out
}

func canonicalQueryString(u *url.URL) string {
	if u.RawQuery == "" {
		return ""
	}
	// Parse query and sort by key then value
	m, _ := url.ParseQuery(u.RawQuery)
	type kv struct {
		k string
		v string
	}
	var pairs []kv
	for k, vs := range m {
		for _, v := range vs {
			pairs = append(pairs, kv{
				k: url.QueryEscape(k),
				v: url.QueryEscape(v),
			})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].k == pairs[j].k {
			return pairs[i].v < pairs[j].v
		}
		return pairs[i].k < pairs[j].k
	})

	var b strings.Builder
	for i, p := range pairs {
		if i > 0 {
			b.WriteString("&")
		}
		b.WriteString(p.k)
		b.WriteString("=")
		b.WriteString(p.v)
	}
	return b.String()
}

func normalizeHeaderValue(v string) string {
	// Collapse internal whitespace to single spaces
	fields := strings.Fields(v)
	return strings.Join(fields, " ")
}

func deriveSigningKey(secret, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func hmacSHA256Hex(key, data []byte) string {
	return hex.EncodeToString(hmacSHA256(key, data))
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
package api

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"slices"
)

// maxBodyBytes is the largest request body the API reads, 64 KiB.
const maxBodyBytes = 64 << 10

// fields holds the members of a request object by name. Members whose value
// is null are left out, because null counts as a missing field.
type fields map[string]jsontext.Value

// decodeFields reads the request body as one JSON object whose member names
// are all in allowed. It checks the media type, the body size, the JSON
// syntax and the member names, in the order of docs/SPEC.md.
func decodeFields(w http.ResponseWriter, r *http.Request, allowed []string) (fields, *problem) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return nil, &problem{code: codeUnsupportedMediaType, detail: "The Content-Type must be application/json."}
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if _, tooLarge := errors.AsType[*http.MaxBytesError](err); tooLarge {
		return nil, &problem{code: codePayloadTooLarge, detail: "The request body is larger than 64 KiB."}
	}
	invalid := &problem{code: codeInvalidJSON, detail: "The request body is not a single valid JSON object."}
	if err != nil {
		return nil, invalid
	}
	// IsValid also rejects invalid UTF-8, duplicate member names and data after the value (D35).
	if v := jsontext.Value(body); !v.IsValid() || v.Kind() != '{' {
		return nil, invalid
	}

	// The body is a valid object at this point, so reading its tokens cannot fail.
	dec := jsontext.NewDecoder(bytes.NewReader(body))
	_, _ = dec.ReadToken()
	result := fields{}
	for dec.PeekKind() != '}' {
		token, _ := dec.ReadToken()
		name := token.String()
		value, _ := dec.ReadValue()
		if !slices.Contains(allowed, name) {
			return nil, &problem{
				code:   codeUnknownField,
				detail: fmt.Sprintf("The field %q is not defined for this endpoint.", name),
				field:  name,
			}
		}
		if value.Kind() != 'n' {
			result[name] = value.Clone()
		}
	}
	return result, nil
}

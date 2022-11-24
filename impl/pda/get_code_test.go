package pda

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const codeCate = "test_cate"

var codes string = `
    <?xml version="1.0" encoding="utf-8"?>
    <string xmlns="http://tempuri.org/">
        <root>
            <code0>test_0</code0>
            <code_dsc0>test_dsc_0</code_dsc0>
            <code1>test_1</code1>
            <code_dsc1>test_dsc_1</code_dsc1>
            <code2>test_2</code2>
            <code_dsc2>test_dsc_2</code_dsc2>
        </root>
    </string>
`

func TestGetCodeFromBRM(t *testing.T) {
	assert := assert.New(t)

	// 404 error.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		defer ts.Close()
		url := ts.URL

		webService := &webService{endpoint: url}

		_, err := webService.GetCodeFromBRM(codeCate)
		assert.Error(err)
	}
	// no response body, unmarshal error.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if r.Method != "GET" {
				t.Errorf("Expected 'GET' request, got '%s'", r.Method)
			}
			if r.URL.EscapedPath() != getCodePath {
				t.Errorf("Expected request to '%s', got '%s'", getCodePath, r.URL.EscapedPath())
			}
			assert.NoError(r.ParseForm())
			topic := r.Form.Get("code_cate")
			if topic != codeCate {
				t.Errorf("Expected request to have 'code_cate=test_cate', got: '%s'", topic)
			}
		}))

		defer ts.Close()
		url := ts.URL

		webService := &webService{endpoint: url}

		_, err := webService.GetCodeFromBRM(codeCate)
		assert.Error(err)
	}
	// normal.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(codes))
			assert.NoError(err)
			if r.Method != "GET" {
				t.Errorf("Expected 'GET' request, got '%s'", r.Method)
			}
			if r.URL.EscapedPath() != getCodePath {
				t.Errorf("Expected request to '%s', got '%s'", getCodePath, r.URL.EscapedPath())
			}
			assert.NoError(r.ParseForm())
			topic := r.Form.Get("code_cate")
			if topic != codeCate {
				t.Errorf("Expected request to have 'code_cate=tes_cate', got: '%s'", topic)
			}
		}))

		defer ts.Close()
		url := ts.URL

		webService := &webService{endpoint: url}

		code, err := webService.GetCodeFromBRM(codeCate)
		assert.NoError(err)
		assert.Equal(Code{
			Code0:    "test_0",
			CodeDsc0: "test_dsc_0",
			Code1:    "test_1",
			CodeDsc1: "test_dsc_1",
			Code2:    "test_2",
			CodeDsc2: "test_dsc_2",
		}, code)
	}
}

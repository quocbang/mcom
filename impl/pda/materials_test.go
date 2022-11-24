package pda

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const barcode = "test_barcode"

const (
	materialInfo = `
    <?xml version="1.0" encoding="utf-8"?>
    <string xmlns="http://tempuri.org/">
        <root>
            <mtrl_product_id>test_product_id</mtrl_product_id>
            <mtrl_lot_id>test_barcode</mtrl_lot_id>
            <mtrl_cate>test_cate</mtrl_cate>
            <mtrl_stat>HOLD</mtrl_stat>
            <mtrl_qty>1.23</mtrl_qty>
            <comment>test</comment>
            <seq_no>001</seq_no>
            <expire_date>2020-02-20</expire_date>
            <code_cnt>0003</code_cnt>
            <code_ext0>ADD</code_ext0>
            <code_dsc0>ADD</code_dsc0>
            <code_ext1>AVAL</code_ext1>
            <code_dsc1>AVAL</code_dsc1>
            <code_ext2>HOLD</code_ext2>
            <code_dsc2>HOLD</code_dsc2>
            <spread_date>3</spread_date>
            <rtn_mesg></rtn_mesg>
        </root>
    </string>
`
	noMaterialInfo = `
    <?xml version="1.0" encoding="utf-8"?>
    <string xmlns="http://tempuri.org/">013：材料條碼不存在!!!</string>
`
)

func TestGetMaterialsInfo(t *testing.T) {
	assert := assert.New(t)

	// 404 error.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		webService := &webService{endpoint: ts.URL}

		_, err := webService.GetMaterialsInfo(barcode)
		assert.Error(err)
	}
	// no response body, unmarshal error.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if r.Method != "GET" {
				t.Errorf("Expected 'GET' request, got '%s'", r.Method)
			}
			if r.URL.EscapedPath() != getMaterialPath {
				t.Errorf("Expected request to '%s', got '%s'", getMaterialPath, r.URL.EscapedPath())
			}
			assert.NoError(r.ParseForm())
			topic := r.Form.Get("strBarcode")
			if topic != barcode {
				t.Errorf("Expected request to have 'strBarcode=test_barcode', got: '%s'", topic)
			}
		}))
		defer ts.Close()

		webService := &webService{endpoint: ts.URL}

		_, err := webService.GetMaterialsInfo(barcode)
		assert.Error(err)
	}
	// barcode not found.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(noMaterialInfo))
			assert.NoError(err)
			if r.Method != "GET" {
				t.Errorf("Expected 'GET' request, got '%s'", r.Method)
			}
			if r.URL.EscapedPath() != getMaterialPath {
				t.Errorf("Expected request to '%s', got '%s'", getMaterialPath, r.URL.EscapedPath())
			}
			assert.NoError(r.ParseForm())
		}))
		defer ts.Close()

		webService := &webService{endpoint: ts.URL}

		m, err := webService.GetMaterialsInfo(barcode)
		assert.NoError(err)
		assert.Nil(m)
	}
	// normal.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(materialInfo))
			assert.NoError(err)
			if r.Method != "GET" {
				t.Errorf("Expected 'GET' request, got '%s'", r.Method)
			}
			if r.URL.EscapedPath() != getMaterialPath {
				t.Errorf("Expected request to '%s', got '%s'", getMaterialPath, r.URL.EscapedPath())
			}
			assert.NoError(r.ParseForm())
			topic := r.Form.Get("strBarcode")
			if topic != barcode {
				t.Errorf("Expected request to have 'strBarcode=test_barcode', got: '%s'", topic)
			}
		}))
		defer ts.Close()

		webService := &webService{endpoint: ts.URL}

		code, err := webService.GetMaterialsInfo(barcode)
		assert.NoError(err)
		assert.Equal(&Material{
			MaterialProductID: "test_product_id",
			MaterialID:        "test_barcode",
			MaterialType:      "test_cate",
			Sequence:          "001",
			Status:            "HOLD",
			Quantity:          "1.23",
			Comment:           "test",
			ExpireDate:        "2020-02-20",
			CodeCnt:           "0003",
			CodeExt0:          "ADD",
			CodeDsc0:          "ADD",
			CodeExt1:          "AVAL",
			CodeDsc1:          "AVAL",
			CodeExt2:          "HOLD",
			CodeDsc2:          "HOLD",
			SpreadDate:        "3",
			ReturnMessage:     "",
		}, code)
	}
}

var xmlString string = "<root><mtrl_product_id>test_product_id</mtrl_product_id><mtrl_lot_id>test_barcode</mtrl_lot_id><mtrl_cate>test_type</mtrl_cate><mtrl_qty>1.23</mtrl_qty><date_add>3</date_add><expire_date>2018-04-24</expire_date><Ename>tester</Ename><mtrl_stat_old>AVAL</mtrl_stat_old><mtrl_stat_new>AVAL</mtrl_stat_new><chg_reason>test</chg_reason><prod_cate>test</prod_cate><chg_area>test</chg_area></root>"

var result string = `
	<?xml version="1.0" encoding="utf-8"?>
	<string xmlns="http://tempuri.org/">
		<root>		
			<rtn_mesg>test</rtn_mesg>
		</root>
	</string>
`

func TestUpdateMaterials(t *testing.T) {
	assert := assert.New(t)

	// 404 error.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		defer ts.Close()
		url := ts.URL

		webService := &webService{endpoint: url}

		_, err := webService.UpdateMaterials(xmlString)
		assert.Error(err)

	}
	// no response body, unmarshal error.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if r.Method != "POST" {
				t.Errorf("Expected 'POST' request, got '%s'", r.Method)
			}
			if r.URL.EscapedPath() != updateMaterialPath {
				t.Errorf("Expected request to '%s', got '%s'", updateMaterialPath, r.URL.EscapedPath())
			}
			assert.NoError(r.ParseForm())
			topic := r.Form.Get("XmlStr")
			if topic != xmlString {
				t.Errorf("Expected request to have 'XmlStr=%v', got: '%s'", xmlString, topic)
			}
		}))

		defer ts.Close()
		url := ts.URL

		webService := &webService{endpoint: url}

		_, err := webService.UpdateMaterials(xmlString)
		assert.Error(err)
	}
	// normal.
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(result))
			assert.NoError(err)
			if r.Method != "POST" {
				t.Errorf("Expected 'POST' request, got '%s'", r.Method)
			}
			if r.URL.EscapedPath() != updateMaterialPath {
				t.Errorf("Expected request to '%s', got '%s'", updateMaterialPath, r.URL.EscapedPath())
			}
			assert.NoError(r.ParseForm())
			topic := r.Form.Get("XmlStr")
			if topic != xmlString {
				t.Errorf("Expected request to have 'XmlStr=%v', got: '%s'", xmlString, topic)
			}
		}))

		defer ts.Close()
		url := ts.URL

		webService := &webService{endpoint: url}

		result, err := webService.UpdateMaterials(xmlString)
		assert.NoError(err)
		assert.Equal("test", result)
	}
}

package pda

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	getMaterialPath    = "/EFAC0/EFAC001/EFAC001.asmx/CheckMat_BarCode_Change"
	updateMaterialPath = "/EFAC0/EFAC001/EFAC001.asmx/SaveChange"
)

const errBarcodeNotFound = "013：材料條碼不存在!!!"

func (s *webService) GetMaterialsInfo(materialID string) (*Material, error) {
	url := fmt.Sprintf("%s%s?%s=%s", s.endpoint, getMaterialPath, "strBarcode", materialID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, newErrorf(url, "failed to call service: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, newErrorf(url, "bad http status code: %v", resp.StatusCode)
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newErrorf(url, "failed to read response data: %v", err)
	}

	res := &MaterialRoot{}
	err = xml.Unmarshal([]byte(RootStringConvert(string(respBody))), &res)
	if err != nil {
		return nil, newErrorf(url, "failed to unmarshal: %v", err)
	}
	if res.Value == errBarcodeNotFound {
		return nil, nil
	}
	return &res.Root, nil
}

// UpdateMaterials incoming xml string.
func (s *webService) UpdateMaterials(req string) (string, error) {
	xmlStr := map[string][]string{"XmlStr": {req}}

	url := fmt.Sprintf("%s%s", s.endpoint, updateMaterialPath)

	resp, err := http.PostForm(url, xmlStr)
	if err != nil {
		return "", newErrorf(url, "failed to call service: %v", err)
	}
	if resp.StatusCode != 200 {
		return "", newErrorf(url, "bad http status code: %v", resp.StatusCode)
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", newErrorf(url, "failed to read response data: %v", err)
	}

	res := UpdateMaterialResult{}
	err = xml.Unmarshal([]byte(RootStringConvert(string(respBody))), &res)
	if err != nil {
		return "", newErrorf(url, "failed to unmarshal: %v", err)
	}

	return strings.TrimSpace(res.Result.Msg), nil
}

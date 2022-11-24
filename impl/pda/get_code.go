package pda

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

const getCodePath = "/EFAC0/EFAC001/EFAC001.asmx/Read_Code_From_BRM"

func (s *webService) GetCodeFromBRM(codeCate string) (Code, error) {
	url := fmt.Sprintf("%s%s?%s=%s", s.endpoint, getCodePath, "code_cate", codeCate)

	resp, err := http.Get(url)
	if err != nil {
		return Code{}, newErrorf(url, "failed to call service: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Code{}, newErrorf(url, "bad http status code: %v", resp.StatusCode)
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Code{}, newErrorf(url, "failed to read response data: %v", err)
	}

	res := &CodeRoot{}
	err = xml.Unmarshal([]byte(RootStringConvert(string(respBody))), &res)
	if err != nil {
		return Code{}, newErrorf(url, "failed to unmarshal: %v", err)
	}
	return res.Root, nil
}

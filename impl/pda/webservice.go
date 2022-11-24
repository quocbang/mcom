package pda

import "encoding/xml"

// WebService is for PDA web service.
type WebService interface {
	GetCodeFromBRM(codeCate string) (Code, error)
	GetMaterialsInfo(materialID string) (*Material, error)
	UpdateMaterials(reqXML string) (string, error)
}

type webService struct {
	endpoint string
}

// NewWebService creates a new PDA web service instance.
func NewWebService(endpoint string) WebService {
	return &webService{endpoint: endpoint}
}

// Code used to receive the xml from http body.
type Code struct {
	Code0         string `xml:"code0"`
	CodeDsc0      string `xml:"code_dsc0"`
	Code1         string `xml:"code1"`
	CodeDsc1      string `xml:"code_dsc1"`
	Code2         string `xml:"code2"`
	CodeDsc2      string `xml:"code_dsc2"`
	Code3         string `xml:"code3"`
	CodeDsc3      string `xml:"code_dsc3"`
	Code4         string `xml:"code4"`
	CodeDsc4      string `xml:"code_dsc4"`
	Code5         string `xml:"code5"`
	CodeDsc5      string `xml:"code_dsc5"`
	Code6         string `xml:"code6"`
	CodeDsc6      string `xml:"code_dsc6"`
	Code7         string `xml:"code7"`
	CodeDsc7      string `xml:"code_dsc7"`
	Code8         string `xml:"code8"`
	CodeDsc8      string `xml:"code_dsc8"`
	Code9         string `xml:"code9"`
	CodeDsc9      string `xml:"code_dsc9"`
	ReturnMessage string `xml:"rtn_mesg"`
}

// CodeRoot used to receive the xml root struct from http body.
type CodeRoot struct {
	Root Code `xml:"root"`
}

// Material used to receive the xml from http body.
type Material struct {
	MaterialProductID string `xml:"mtrl_product_id"`
	MaterialID        string `xml:"mtrl_lot_id"`
	MaterialType      string `xml:"mtrl_cate"`
	Sequence          string `xml:"seq_no"`
	Status            string `xml:"mtrl_stat"`
	Quantity          string `xml:"mtrl_qty"`
	Comment           string `xml:"comment"`
	ExpireDate        string `xml:"expire_date"`
	CodeCnt           string `xml:"code_cnt"`
	CodeExt0          string `xml:"code_ext0"`
	CodeDsc0          string `xml:"code_dsc0"`
	CodeExt1          string `xml:"code_ext1"`
	CodeDsc1          string `xml:"code_dsc1"`
	CodeExt2          string `xml:"code_ext2"`
	CodeDsc2          string `xml:"code_dsc2"`
	CodeExt3          string `xml:"code_ext3"`
	CodeDsc3          string `xml:"code_dsc3"`
	CodeExt4          string `xml:"code_ext4"`
	CodeDsc4          string `xml:"code_dsc4"`
	CodeExt5          string `xml:"code_ext5"`
	CodeDsc5          string `xml:"code_dsc5"`
	CodeExt6          string `xml:"code_ext6"`
	CodeDsc6          string `xml:"code_dsc6"`
	SpreadDate        string `xml:"spread_date"`
	ReturnMessage     string `xml:"rtn_mesg"`
}

// MaterialRoot used to receive the xml root struct from http body.
type MaterialRoot struct {
	Root  Material `xml:"root"`
	Value string   `xml:",chardata"`
}

// UpdateMaterialResultMessage used to receive the xml from http body.
type UpdateMaterialResultMessage struct {
	Msg string `xml:"rtn_mesg"`
}

// UpdateMaterialResult definition.
type UpdateMaterialResult struct {
	Result UpdateMaterialResultMessage `xml:"root"`
}

// SaveChangeRequest used to marshal XML string.
type SaveChangeRequest struct {
	XMLName           xml.Name `xml:"root"`
	MaterialProductID string   `xml:"mtrl_product_id"`
	MaterialID        string   `xml:"mtrl_lot_id"`
	MaterialType      string   `xml:"mtrl_cate"`
	Quantity          string   `xml:"mtrl_qty"`
	DateAdd           string   `xml:"date_add"`
	ExpireDate        string   `xml:"expire_date"`
	User              string   `xml:"Ename"`
	OldStatus         string   `xml:"mtrl_stat_old"`
	NewStatus         string   `xml:"mtrl_stat_new"`
	Reason            string   `xml:"chg_reason"`
	ProduceCate       string   `xml:"prod_cate"`
	Area              string   `xml:"chg_area"`
}

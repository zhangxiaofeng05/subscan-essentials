package subscan

import "github.com/shopspring/decimal"

type ChainBlock struct {
	BlockNum       int    `json:"block_num"`
	BlockTimestamp int    `json:"block_timestamp"`
	Hash           string `sql:"default: null;size:100" json:"hash"`
	ParentHash     string `sql:"default: null;size:100" json:"parent_hash"`
	StateRoot      string `sql:"default: null;size:100" json:"state_root"`
	//ExtrinsicsRoot  string `sql:"default: null;size:100" json:"extrinsics_root"`
	Logs            string `json:"logs" sql:"type:text;"`
	Extrinsics      string `json:"extrinsics" sql:"type:MEDIUMTEXT;"`
	EventCount      int    `json:"event_count"`
	ExtrinsicsCount int    `json:"extrinsics_count"`
	Event           string `json:"event" sql:"type:MEDIUMTEXT;"`
	SpecVersion     int    `json:"spec_version"`
	Validator       string `json:"validator"`
	CodecError      bool   `json:"codec_error"`
	//Finalized   bool   `json:"finalized"`
}

type ChainEvent struct {
	ID            uint        `gorm:"primary_key" json:"-"`
	EventIndex    string      `sql:"default: null;size:100;" json:"event_index"`
	BlockNum      int         `json:"block_num" `
	ExtrinsicIdx  int         `json:"extrinsic_idx"`
	Type          string      `json:"-"`
	ModuleId      string      `json:"module_id" `
	EventId       string      `json:"event_id" `
	Params        interface{} `json:"params" sql:"type:text;" `
	ExtrinsicHash string      `json:"extrinsic_hash" sql:"default: null" `
	EventIdx      int         `json:"event_idx"`
}

type ChainExtrinsic struct {
	ID                 uint            `gorm:"primary_key"`
	ExtrinsicIndex     string          `json:"extrinsic_index" sql:"default: null;size:100"`
	BlockNum           int             `json:"block_num" `
	BlockTimestamp     int             `json:"block_timestamp"`
	ExtrinsicLength    int             `json:"extrinsic_length"`
	VersionInfo        string          `json:"version_info"`
	CallCode           string          `json:"call_code"`
	CallModuleFunction string          `json:"call_module_function"  sql:"size:100"`
	CallModule         string          `json:"call_module"  sql:"size:100"`
	Params             interface{}     `json:"params" sql:"type:MEDIUMTEXT;" `
	AccountId          string          `json:"account_id"`
	Signature          string          `json:"signature"`
	Nonce              int             `json:"nonce"`
	Era                string          `json:"era"`
	ExtrinsicHash      string          `json:"extrinsic_hash" sql:"default: null" `
	IsSigned           bool            `json:"is_signed"`
	Success            bool            `json:"success"`
	Fee                decimal.Decimal `json:"fee" sql:"type:decimal(30,0);"`
	BatchIndex         int             `json:"-" gorm:"-"`
}

type ExtrinsicParam struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Value    interface{} `json:"value"`
	ValueRaw string      `json:"valueRaw"`
}

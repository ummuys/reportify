package models

type SSCDBCreateRequest struct {
	AllowCTAS            bool                   `json:"allow_ctas"`
	AllowCVAS            bool                   `json:"allow_cvas"`
	AllowDML             bool                   `json:"allow_dml"`
	AllowFileUpload      bool                   `json:"allow_file_upload"`
	AllowRunAsync        bool                   `json:"allow_run_async"`
	CacheTimeout         int                    `json:"cache_timeout"`
	ConfigurationMethod  string                 `json:"configuration_method"`
	DatabaseName         string                 `json:"database_name"`
	Driver               string                 `json:"driver"`
	Engine               string                 `json:"engine"`
	ExposeInSQLLab       bool                   `json:"expose_in_sqllab"`
	ExternalURL          string                 `json:"external_url"`
	Extra                string                 `json:"extra"`
	ForceCTASSchema      string                 `json:"force_ctas_schema"`
	ImpersonateUser      bool                   `json:"impersonate_user"`
	IsManagedExternally  bool                   `json:"is_managed_externally"`
	MaskedEncryptedExtra string                 `json:"masked_encrypted_extra"`
	Parameters           map[string]interface{} `json:"parameters"`
	ServerCert           string                 `json:"server_cert"`
	SQLAlchemyURI        string                 `json:"sqlalchemy_uri"`
	SSHTunnel            *SSHTunnel             `json:"ssh_tunnel"`
	UUID                 string                 `json:"uuid"`
}

type SSHTunnel struct {
	ID                 int    `json:"id"`
	Password           string `json:"password"`
	PrivateKey         string `json:"private_key"`
	PrivateKeyPassword string `json:"private_key_password"`
	ServerAddress      string `json:"server_address"`
	ServerPort         int    `json:"server_port"`
	Username           string `json:"username"`
}

type SSCAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Refresh  bool   `json:"refresh"`
	Provider string `json:"provider"`
}

type SSCCreateDatasetRequest struct {
	AlwaysFilterMainDttm bool   `json:"always_filter_main_dttm"`
	Catalog              string `json:"catalog"`
	Database             int    `json:"database"`
	ExternalURL          string `json:"external_url"`
	IsManagedExternally  bool   `json:"is_managed_externally"`
	NormalizeColumns     bool   `json:"normalize_columns"`
	Owners               []int  `json:"owners"`
	Schema               string `json:"schema"`
	SQL                  string `json:"sql"`
	TableName            string `json:"table_name"`
}

type SSCChartCreateRequest struct {
	CacheTimeout           int     `json:"cache_timeout,omitempty"`
	CertificationDetails   string  `json:"certification_details,omitempty"`
	CertifiedBy            string  `json:"certified_by,omitempty"`
	Dashboards             []int64 `json:"dashboards,omitempty"`
	DatasourceID           int64   `json:"datasource_id"`
	DatasourceName         string  `json:"datasource_name,omitempty"`
	DatasourceType         string  `json:"datasource_type"`
	Description            string  `json:"description,omitempty"`
	ExternalURL            string  `json:"external_url,omitempty"`
	IsManagedExternally    bool    `json:"is_managed_externally,omitempty"`
	Owners                 []int64 `json:"owners,omitempty"`
	Params                 string  `json:"params"`
	QueryContext           string  `json:"query_context,omitempty"`
	QueryContextGeneration bool    `json:"query_context_generation,omitempty"`
	SliceName              string  `json:"slice_name"`
	VizType                string  `json:"viz_type"`
}

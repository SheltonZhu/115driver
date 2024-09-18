package driver

// GetAppVersion get app version (win, android, mac, mac_arc, etc...)
func (c *Pan115Client) GetAppVersion() ([]AppVersion, error) {
	result := VersionResp{}
	req := c.NewRequest().
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8")

	resp, err := req.Get(ApiGetVersion)

	err = CheckErr(err, &result, resp)
	if err != nil {
		return nil, err
	}

	return result.Data.GetAppVersions(), nil
}

type VersionResp struct {
	BasicResp
	Data Versions `json:"data"`
}

type Versions map[string]map[string]any

func (v Versions) GetAppVersions() []AppVersion {
	vers := make([]AppVersion, len(v))
	for app, ver := range v {
		vers = append(vers, AppVersion{
			AppName: app,
			Version: ver["version_code"].(string),
		})
	}
	return vers
}

type AppVersion struct {
	AppName string
	Version string
}

package driver

// GetInfo get space info and login device info.
func (c *Pan115Client) GetInfo() (InfoData, error) {
	result := InfoResponse{}
	req := c.NewRequest().
		SetResult(&result).
		ForceContentType("application/json;charset=UTF-8")

	resp, err := req.Get(ApiFileIndexInfo)

	if err = CheckErr(err, &result, resp); err != nil {
		return InfoData{}, err
	}
	return result.Data, nil
}

type InfoResponse struct {
	BasicResp
	Data InfoData `json:"data"`
}

type InfoData struct {
	SpaceInfo        SpaceInfo        `json:"space_info"`
	LoginDevicesInfo LoginDevicesInfo `json:"login_devices_info"`
	ImeiInfo         bool             `json:"imei_info"`
}

type TotalSize struct {
	Size       int64  `json:"size"`
	SizeFormat string `json:"size_format"`
}

type RemainSize struct {
	Size       int64  `json:"size"`
	SizeFormat string `json:"size_format"`
}

type UseSize struct {
	Size       int64  `json:"size"`
	SizeFormat string `json:"size_format"`
}

type SpaceInfo struct {
	AllTotal  TotalSize  `json:"all_total"`
	AllRemain RemainSize `json:"all_remain"`
	AllUse    UseSize    `json:"all_use"`
}

type LastDevice struct {
	IP       string `json:"ip"`
	Device   string `json:"device"`
	DeviceID string `json:"device_id"`
	Network  string `json:"network"`
	Os       string `json:"os"`
	City     string `json:"city"`
	Utime    int    `json:"utime"`
}

type Device struct {
	IsCurrent int    `json:"is_current"`
	Ssoent    string `json:"ssoent"`
	Utime     int    `json:"utime"`
	Device    string `json:"device"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Desc      string `json:"desc"`
	IP        string `json:"ip"`
	City      string `json:"city"`
	IsUnusual int    `json:"is_unusual"`
}

type LoginDevicesInfo struct {
	Last LastDevice `json:"last"`
	List []Device   `json:"list"`
}

package models

import "gofr.dev/pkg/gofr/file"

type ImageDetails struct {
	Data file.Zip `file:"image"`
	Name string   `form:"name"`
	Tag  string   `form:"tag"`

	ServiceDetails
}

type ServiceDetails struct {
	ServiceID     string `form:"serviceID"`
	ServiceCreds  any    `form:"serviceCreds"`
	Repository    string `form:"repository"`
	Region        string `form:"region"`
	LoginServer   string `form:"loginServer"`
	ServiceName   string `form:"serviceName"`
	AccountID     string `form:"accountID"`
	NameSpace     string `form:"nameSpace"`
	CloudPlatform string `form:"cloudPlatform"`
}

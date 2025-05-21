package configurations

import "github.com/davmontas/teacherapp/docs" // This is required to generete swagger docs

func SwaggerInfo(version, host string) {
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.BasePath = "/v1"
}

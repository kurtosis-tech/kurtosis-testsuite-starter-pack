package my_custom_service
/*
	NEW USER ONBOARDING:
	- Rename this package, this file, and the containing directory to reflect the functionality of your custom test.
	- Rename all structs and functions within to reflect your custom service.
*/

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
)


type MyCustomService struct {
	serviceCtx *services.ServiceContext
	port       int
}

func NewMyCustomService(serviceCtx *services.ServiceContext, port int) *MyCustomService {
	return &MyCustomService{serviceCtx: serviceCtx, port: port}
}

// ===========================================================================================
//                              Service interface methods
// ===========================================================================================
func (service MyCustomService) IsAvailable() bool {
	/*
		NEW USER ONBOARDING:
		- Write logic, likely using the port property of your service object, to verify that your service is available and ready to be tested.
	*/
	return true
}
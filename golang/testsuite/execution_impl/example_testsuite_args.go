/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package execution_impl

type ExampleTestsuiteArgs struct {

	/*
		NEW USER ONBOARDING:
		- Change this property name to reflect the name of your custom service image.
		- Change the string after "json:" to reflect the customServiceImage key in the json in build-and-run.sh.
	*/
	MyCustomServiceImage string		`json:"myCustomServiceImage"`
}

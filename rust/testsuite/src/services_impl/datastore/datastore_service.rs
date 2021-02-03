use reqwest;

use kurtosis_rust_lib::services::service;
use reqwest::Request;

const HEALTHCHECK_URL_SLUG: &str = "health";
const HEALTHY_VALUE: &str = "healthy";
const TEXT_CONTENT_TYPE: &str = "text/plain";
const KEY_ENDPOINT: &str = "key";

struct DatastoreService {
    service_id: service::ServiceId,
    ip_addr: &'static str,
    port: i32,
}

impl service::Service for DatastoreService {
    fn get_service_id(&self) -> &str {
        return self.service_id;
    }

    fn get_ip_address(&self) -> &str {
        return self.ip_addr;
    }

    fn is_available(&self) -> bool {
        let client = reqwest::Client::new();
        let url = format!(
            "http://{}:{}/{}",
            self.ip_addr,
            self.port,
            HEALTHCHECK_URL_SLUG,
        );
        let resp_or_err = client.get(url)
            .send()
            .await;
        if resp_or_err.is_err() {
            debug!(
                "An HTTP error occurred when polling the health endpoint: {}",
                resp_or_err.unwrap_err().to_string()
            );
            return false;
        }
        let resp = resp_or_err.unwrap();
        if !resp.status().is_success() {
            debug!("Received non-OK status code: {}", resp.status().as_u16());
            return false;
        }

        let resp_body_or_err = resp.text().await;
        if resp_body_or_err.is_err() {
            debug!(
                "An error occurred reading the response body: {}",
                resp_body_or_err.unwrap_err().to_string()
            );
            return false;
        }
        let resp_body = resp_body_or_err.unwrap();
        return resp_body == HEALTHY_VALUE;
    }
}

/*

func NewDatastoreService(serviceId services.ServiceID, ipAddr string, port int) *DatastoreService {
	return &DatastoreService{serviceId: serviceId, ipAddr: ipAddr, port: port}
}

// ===========================================================================================
//                              Service interface methods
// ===========================================================================================
func (service DatastoreService) GetServiceID() services.ServiceID {
	return service.serviceId
}

func (service DatastoreService) GetIPAddress() string {
	return service.ipAddr
}


func (service DatastoreService) IsAvailable() bool {
	url := fmt.Sprintf("http://%v:%v/%v", service.GetIPAddress(), service.port, healthcheckUrlSlug)
	resp, err := http.Get(url)
	if err != nil {
		logrus.Debugf("An HTTP error occurred when polliong the health endpoint: %v", err)
		return false
	}
	if resp.StatusCode != http.StatusOK {
		logrus.Debugf("Received non-OK status code: %v", resp.StatusCode)
		return false
	}

	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		logrus.Debugf("An error occurred reading the response body: %v", err)
		return false
	}
	bodyStr := string(bodyBytes)

	return bodyStr == healthyValue
}

// ===========================================================================================
//                         Datastore service-specific methods
// ===========================================================================================
func (service DatastoreService) GetPort() int {
	return service.port
}

func (service DatastoreService) Exists(key string) (bool, error) {
	url := service.getUrlForKey(key)
	resp, err := http.Get(url)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred requesting data for key '%v'", key)
	}
	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		return false, stacktrace.NewError("Got an unexpected HTTP status code: %v", resp.StatusCode)
	}
}

func (service DatastoreService) Get(key string) (string, error) {
	url := service.getUrlForKey(key)
	resp, err := http.Get(url)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred requesting data for key '%v'", key)
	}
	if resp.StatusCode != http.StatusOK {
		return "", stacktrace.NewError("A non-%v status code was returned", resp.StatusCode)
	}
	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the response body")
	}
	return string(bodyBytes), nil
}

func (service DatastoreService) Upsert(key string, value string) error {
	url := service.getUrlForKey(key)
	resp, err := http.Post(url, textContentType, strings.NewReader(value))
	if err != nil {
		return stacktrace.Propagate(err, "An error requesting to upsert data '%v' to key '%v'", value, key)
	}
	if resp.StatusCode != http.StatusOK {
		return stacktrace.NewError("A non-%v status code was returned", resp.StatusCode)
	}
	return nil
}

func (service DatastoreService) getUrlForKey(key string) string {
	return fmt.Sprintf("http://%v:%v/%v/%v", service.GetIPAddress(), service.port, keyEndpoint, key)
}

 */
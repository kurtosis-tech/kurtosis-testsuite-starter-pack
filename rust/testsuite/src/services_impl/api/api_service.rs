use futures::executor::block_on;
use kurtosis_rust_lib::services::service::Service;
use reqwest::header::CONTENT_TYPE;
use anyhow::{anyhow, Context, Result};
use serde::{Deserialize, Serialize};

const HEALTHCHECK_URL_SLUG: &str = "health";
const HEALTHY_VALUE: &str = "healthy";

const TEXT_CONTENT_TYPE: &str = "text/plain";

const PERSON_ENDPOINT: &str = "person";
const INCREMENT_BOOKS_READ_ENDPOINT: &str = "incrementBooksRead";

#[derive(Serialize, Deserialize)]
pub struct Person {
    #[serde(rename = "booksRead")]
    pub books_read: u64,
}

pub struct ApiService {
    service_id: String,
    ip_addr: String,
    port: u32,
}

impl ApiService {
    pub fn new(service_id: String, ip_addr: String, port: u32) -> ApiService {
        return ApiService{
            service_id,
            ip_addr,
            port,
        }
    }

	pub fn add_person(&self, id: u32) -> Result<()> {
		let client = reqwest::blocking::Client::new();
		let url = self.get_person_url_for_id(id);
        let resp = client.post(&url)
            .header(CONTENT_TYPE, TEXT_CONTENT_TYPE)
            .send()
			.context(format!("An error occurred making the request to add person with ID '{}'", id))?;
		let resp_status = resp.status();
        if !resp_status.is_success() {
            return Err(anyhow!(
                "A non-successful error code was returned: {}", 
                resp_status.as_u16()
            ));
        }
		return Ok(());
	}

	pub fn get_person(&self, id: u32) -> Result<Person> {
		let url = self.get_person_url_for_id(id);
		let resp = reqwest::blocking::get(&url)
			.context(format!("An error occurred making the request to get person with ID '{}'", id))?;
		let resp_status = resp.status();
        if !resp_status.is_success() {
            return Err(anyhow!(
                "A non-successful error code was returned: {}", 
                resp_status.as_u16()
            ));
        }
		let resp_body = resp.text()
			.context("An error occurred reading the response body")?;
		let person: Person = serde_json::from_str(&resp_body)
			.context("An error occurred deserializing the Person JSON")?;
		return Ok(person);
	}

	pub fn increment_books_read(&self, id: u32) -> Result<()> {
		let client = reqwest::blocking::Client::new();
		let url = format!("http://{}:{}/{}/{}", self.ip_addr, self.port, INCREMENT_BOOKS_READ_ENDPOINT, id);
		let resp = client.post(&url)
			.header(CONTENT_TYPE, TEXT_CONTENT_TYPE)
			.send()			
            .context(format!("An error occurred making the request to increment the books read of person with ID '{}'", id))?;
		let resp_status = resp.status();
        if !resp_status.is_success() {
            return Err(anyhow!(
                "A non-successful error code was returned: {}", 
                resp_status.as_u16()
            ));
        }
		return Ok(());
	}

	fn get_person_url_for_id(&self, id: u32) -> String {
		return format!("http://{}:{}/{}/{}", self.ip_addr, self.port, PERSON_ENDPOINT, id);
	}
}


// ===========================================================================================
//                              Service interface methods
// ===========================================================================================
impl Service for ApiService {
    fn get_service_id(&self) -> &str {
		return &self.service_id;
    }

    fn get_ip_address(&self) -> &str {
		return &self.ip_addr;
    }

    fn is_available(&self) -> bool {
        let client = reqwest::blocking::Client::new();
        let url = format!(
            "http://{}:{}/{}",
            self.ip_addr,
            self.port,
            HEALTHCHECK_URL_SLUG,
        );
        let resp_or_err = client.get(&url).send();
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

        let resp_body_or_err = resp.text();
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



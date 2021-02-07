use reqwest;

use kurtosis_rust_lib::services::service;
use reqwest::Request;
use kurtosis_rust_lib::services::service::Service;
use std::error::Error;
use std::fs::File;
use reqwest::header::CONTENT_TYPE;

const HEALTHCHECK_URL_SLUG: &str = "health";
const HEALTHY_VALUE: &str = "healthy";
const TEXT_CONTENT_TYPE: &str = "text/plain";
const KEY_ENDPOINT: &str = "key";
const NOT_FOUND_ERR_CODE: u16 = 404;

pub struct DatastoreService {
    service_id: service::ServiceId,
    ip_addr: &'static str,
    port: u32,
}

impl DatastoreService {
    pub fn new(service_id: &str, ip_addr: &str, port: u32) -> DatastoreService {
        return DatastoreService{
            service_id,
            ip_addr,
            port,
        };
    }

    pub fn get_port(&self) -> u32 {
        return self.port;
    }

    pub fn exists(&self, key: &str) -> Result<bool, dyn Error> {
        self.get_url_for_key(key);

        let url = self.get_url_for_key(key);
        let resp = reqwest::get(url)
            .await?;
        let resp_status = resp.status();
        if resp_status.is_success() {
            return Ok(true);
        } else if resp_status.as_u16() == NOT_FOUND_ERR_CODE {
            return Ok(false);
        } else {
            return Err(format!("Got an unexpected HTTP status code: {}", resp_status));
        }
    }

    pub fn get(&self, key: &str) -> Result<String, dyn Error> {
        url = self.get_url_for_key(key);
        let resp = reqwest::get(url)
            .await?;
        let resp_status = resp.status();
        if !resp_status.is_success() {
            return Err(format!("A non-successful error code was returned: {}", resp_status.as_u16()))
        }
        let resp_body = resp.text().await?;
        return Ok(resp_body)
    }

    pub fn upsert(&self, key: &str, value: &str) -> Result<(), dyn Error> {
        let url = self.get_url_for_key(key);
        let client = reqwest::Client::new();
        let resp = client.post(url)
            .header(CONTENT_TYPE, TEXT_CONTENT_TYPE)
            .body(value)
            .send()
            .await?;
        let resp_status = resp.status();
        if !resp_status.is_success() {
            return Err(format!("Got non-OK status code: {}", resp_status.as_u16()))
        }
        return Ok(());
    }

    // ==========================================================================================
    //                                Private helper functions
    // ==========================================================================================
    fn get_url_for_key(&self, key: &str) -> String {
        return format!(
            "http://{}:{}/{}/{}",
            self.get_ip_address(),
            self.get_port(),
            KEY_ENDPOINT,
            key
        );
    }
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


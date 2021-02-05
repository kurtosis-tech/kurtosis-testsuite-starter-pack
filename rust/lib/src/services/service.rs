/*
The developer should implement their own use-case-specific interface that extends this one
 */
pub trait Service {
    fn get_service_id(&self) -> &str;

    // Returns the IP address of the service
    fn get_ip_address(&self) -> &str;

    // Returns true if the service is available
    fn is_available(&self) -> bool;
}
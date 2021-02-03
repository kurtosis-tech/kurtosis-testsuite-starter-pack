pub trait Service {
    fn get_service_id(&self) -> &str;

    fn get_ip_address(&self) -> &str;

    fn is_available(&self) -> bool;
}
use crate::services::service::Service;

// This is necessary as a trait due to Rust's rule that every closure is a completely different type (so we can't
// share types)
pub trait ServiceInterfaceWrapper<T: Service> {
    fn wrap(&self, service_id: &str, service_ip_addr: &str) -> T;
}
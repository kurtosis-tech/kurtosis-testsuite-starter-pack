use downcast_rs::{Downcast, impl_downcast};

/*
The developer should implement their own use-case-specific interface that extends this one
 */
pub trait Service: Downcast {
    // Returns true if the service is available
    fn is_available(&self) -> bool;
}
impl_downcast!(Service);
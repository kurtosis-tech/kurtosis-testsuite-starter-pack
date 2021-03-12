use downcast_rs::{DowncastSync, impl_downcast};

/*
The developer should implement their own use-case-specific interface that extends this one
 */
pub trait Service: DowncastSync {
    // Returns true if the service is available
    fn is_available(&self) -> bool;
}
impl_downcast!(sync Service);
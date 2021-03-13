use downcast_rs::{Downcast, impl_downcast};

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub trait Service: Downcast {
    // Returns true if the service is available
    fn is_available(&self) -> bool;
}
impl_downcast!(Service);
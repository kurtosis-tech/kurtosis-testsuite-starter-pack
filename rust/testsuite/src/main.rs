extern crate pretty_env_logger;
#[macro_use] extern crate log;

use crate::services_impl::datastore::datastore_service::DatastoreService;

mod services_impl;

fn main() {
    let service = DatastoreService::new(
        "test-service-id",
        "1.2.3.4",
        4567
    );

    println!(service.get_port());
}

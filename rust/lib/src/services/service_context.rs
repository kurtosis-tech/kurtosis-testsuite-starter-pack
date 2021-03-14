use std::rc::Rc;

use anyhow::{Context, Result};
use tokio::runtime::Runtime;
use tonic::transport::Channel;

use crate::core_api_bindings::api_container_api::{ExecCommandArgs, test_execution_service_client::TestExecutionServiceClient};

use super::service::ServiceId;

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct ServiceContext {
    async_runtime: Rc<Runtime>,
    client: TestExecutionServiceClient<Channel>,
    service_id: ServiceId,
    ip_address: String,
}

impl ServiceContext {
    pub fn new(async_runtime: Rc<Runtime>, client: TestExecutionServiceClient<Channel>, service_id: ServiceId, ip_address: String) -> ServiceContext {
        return ServiceContext{
            async_runtime,
            client,
            service_id,
            ip_address,
        }
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn get_service_id(&self) -> ServiceId {
        return self.service_id.clone();
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn get_ip_address(&self) -> &str {
        return &self.ip_address;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn exec_command(&self, command: Vec<String>) -> Result<(i32, Vec<u8>)> {
        let args = ExecCommandArgs{
            service_id: self.service_id.clone(),
            command_args: command.clone(),
        };
        let req = tonic::Request::new(args);
        // The client needs to be mutable, so the correct/supported way according to the `tonic` docs
        // is to clone the client
        let mut client = self.client.clone();
        let resp = self.async_runtime.block_on(client.exec_command(req))
            .context(format!("An error occurred executing command '{:?}' on service '{}'", &command, self.service_id))?
            .into_inner();
        return Ok((resp.exit_code, resp.log_output));
    }
}
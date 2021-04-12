use std::{borrow::Borrow, collections::{HashMap, HashSet}, path::PathBuf, rc::Rc};

use anyhow::{Context, Result};
use tokio::runtime::Runtime;
use tonic::transport::Channel;

use crate::core_api_bindings::api_container_api::{ExecCommandArgs, FileGenerationOptions, GenerateFilesArgs, file_generation_options::FileTypeToGenerate, test_execution_service_client::TestExecutionServiceClient};

use super::service::ServiceId;

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct GeneratedFileFilepaths {
    pub absolute_filepath_on_testsuite_container: PathBuf,
    pub absolute_filepath_on_service_container: PathBuf,
}

// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct ServiceContext {
    async_runtime: Rc<Runtime>,
    client: TestExecutionServiceClient<Channel>,
    service_id: ServiceId,
    ip_address: String,
    test_volume_mountpoint_on_testsuite_container: String,
    test_volume_mountpoint_on_service_container: String,
}

impl ServiceContext {
    pub fn new(
            async_runtime: Rc<Runtime>, 
            client: TestExecutionServiceClient<Channel>, 
            service_id: ServiceId, 
            ip_address: String,
            test_volume_mountpoint_on_testsuite_container: String,
            test_volume_mountpoint_on_service_container: String) -> ServiceContext {
        return ServiceContext{
            async_runtime,
            client,
            service_id,
            ip_address,
            test_volume_mountpoint_on_testsuite_container,
            test_volume_mountpoint_on_service_container,
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

    // Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
    pub fn generate_files(&self, files_to_generate: HashSet<String>) -> Result<HashMap<String, GeneratedFileFilepaths>> {
        let mut file_generation_opts: HashMap<String, FileGenerationOptions> = HashMap::new();
        for file_id in &files_to_generate {
            let opts = FileGenerationOptions{
                file_type_to_generate: FileTypeToGenerate::File.into(),
            };
            file_generation_opts.insert(file_id.to_owned(), opts);
        }
        let args = GenerateFilesArgs{
            service_id: self.service_id.clone(),
            files_to_generate: file_generation_opts,
        };
        // The client needs to be mutable, so the correct/supported way according to the `tonic` docs
        // is to clone the client
        let mut client = self.client.clone();
        let req = tonic::Request::new(args);
        let resp = self.async_runtime.block_on(client.generate_files(req))
            .context(format!("An error occurred generating files for service ID '{}'", self.service_id))?
            .into_inner();
        let generated_file_relative_filepaths = resp.generated_file_relative_filepaths;

        let mut result: HashMap<String, GeneratedFileFilepaths> = HashMap::new();
        for file_id in &files_to_generate {
            let relative_filepath = generated_file_relative_filepaths.get(file_id)
                .context(
                    format!(
                        "No filepath (relative to test volume root) was returned for file '{}', even though we requested it; this is a Kurtosis bug",
                        file_id
                    )
                )?;
            let abs_filepath_on_testsuite: PathBuf = [self.test_volume_mountpoint_on_testsuite_container.borrow(), relative_filepath].iter().collect();
            let abs_filepath_on_service: PathBuf = [self.test_volume_mountpoint_on_service_container.borrow(), relative_filepath].iter().collect();
            let generated_file_filepaths = GeneratedFileFilepaths{
                absolute_filepath_on_testsuite_container: abs_filepath_on_testsuite,
                absolute_filepath_on_service_container: abs_filepath_on_service,
            };
            result.insert(file_id.to_owned(), generated_file_filepaths);
        }
        return Ok(result);
    }
}
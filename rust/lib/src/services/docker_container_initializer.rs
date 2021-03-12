

// The ID of an artifact containing files that should be mounted into a service container
// type FilesArtifactID string

use std::{collections::{HashSet, HashMap}, path::PathBuf};
use crate::services::service::Service;
use std::fs::File;
use anyhow::Result;

use super::service_context::ServiceContext;


// TODO Create a DockerContainerInitializerBuilder rather than forcing users to update their code with a new
//  method every time a new feature comes out!
pub trait DockerContainerInitializer<T: Service + Send + Sync> {
    // Gets the Docker image that will be used for instantiating the Docker container
    fn get_docker_image(&self) -> &str;

    // Gets the "set" of ports that the Docker container running the service will listen on
    // This is in Docker port specification syntax, e.g. "80" (default TCP) or "80/udp"
    // It might even support ranges (e.g. "90:100/tcp"), though this is untested as of 2020-12-08
    fn get_used_ports(&self) -> HashSet<String>;

    /*
		Wrap the service context inside a user-defined service interface

        Returns:
            An instance of the user-defined service that this container initializer produces
    */
    fn get_service(&self, service_ctx: ServiceContext) -> Box<dyn Service>;

    /*
        This method is used to declare that the service will need a set of files in order to run. To do this, the developer
        declares a set of string keys that are meaningful to the developer, and Kurtosis will create one file per key. These newly-createed
        file objects will then be passed in to the `InitializeGeneratedFiles` and `GetStartCommand` functions below keyed on the
        strings that the developer passed in, so that the developer can initialize the contents of the files as they please.
        Kurtosis then guarantees that these files will be made available to the service at startup time.

        NOTE: The keys that the developer returns here are ONLY used for developer identification purposes; the actual
        filenames and filepaths of the file are implementation details handled by Kurtosis!

        Returns:
            A "set" of user-defined key strings identifying the files that the service will need, which is how files will be
                identified in `InitializeMountedFiles` and `GetStartCommand`
    */
    fn get_files_to_generate(&self) -> HashSet<String>;

    /*
        Initializes the contents of the files that the developer requested in `GetFilesToMount` with whatever
            contents the developer desires. This will be called before service startup.

        Args:
            mounted_files: A mapping of developer_key -> file_pointer, with developer_key corresponding to the keys declares in
                `GetFilesToMount`
    */
    fn initialize_generated_files(&self, generated_files: HashMap<String, File>) -> Result<()>;

    /*
        Allows the mounting of external files into a service container by mapping files artifacts (defined in your
        test's configuration) to mountpoints on the service container.

        NOTE: As of 2021-01-06, only GZ-compressed TAR artifacts are supported.

        Returns:
            A map of filesArtifactId -> serviceContainerMountpoint, where:
                1) The map key is the ID of the files artifact as defined in your TestConfiguration.
                2) The map value is the filepath inside of the service container where the
                    contents of the archive file should be mounted after decompression.
     */
    fn get_files_artifact_mountpoints(&self) -> HashMap<String, String>;


    /*
        Kurtosis mounts the files that the developer requested in `GetFilesToMount` via a Docker volume, but Kurtosis doesn't
        know anything about the Docker image backing the service so therefore doesn't know what filepath it can safely mount
        the volume on. This function uses the developer's knowledge of the Docker image running the service to inform
        Kurtosis of a filepath where the Docker volume can be safely mounted.

        Returns:
            A filepath on the Docker image backing this service that's safe to mount the test volume on
    */
    fn get_test_volume_mountpoint(&self) -> &'static str;

	/*
		Can optionally override the ENTRYPOINT and CMD Docker directives when starting the Docker container.

		Args:
			generated_file_filepaths: Mapping of developer_key -> generated_file_filepath where developer_key corresponds to the keys returned
				in the `GetFilesToGenerate` function, and generated_file_filepath is the path *on the Docker container* of where the
				file has been mounted. The files will have already been initialized via the `InitializeGeneratedFiles` function.
			ipAddr: The IP address of the service being started.

		Returns:
			entrypoint_args: If non-None, overrides the ENTRYPOINT directive of the Docker image with the given strings. If nil,
				the default ENTRYPOINT is used.
			cmd_args: If non-None, overrides the CMD directive of the Docker image of the Docker image with the given strings. If nil,
				the default CMD is used.
	*/
    fn get_start_command_overrides(
        &self,
        generated_file_filepaths: HashMap<String, PathBuf>,
        ip_addr: &str
    ) -> Result<(Option<Vec<String>>, Option<Vec<String>>)>;
}

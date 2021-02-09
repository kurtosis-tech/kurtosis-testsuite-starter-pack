use std::{env::current_dir, error::Error, fs::{read_dir, remove_file}, path::{Path, PathBuf}, process::exit};
use anyhow::{Context, Result};

const RUST_FILE_NAME_REGEX: &str = r"^.*\.rs$";

fn main() -> Result<()> {
    let current_dirpath = current_dir()
        .context("Couldn't get current directory")?;

    let root_dirpath = current_dirpath.join("../..");
    let root_dirpath = root_dirpath.canonicalize()
        .context(format!("An error occurred canonicalizing root dirpath {}", root_dirpath.display()))?;

    let protos_dirpath = root_dirpath.join("core-api");
    let protos_dirpath = protos_dirpath.canonicalize()
        .context(format!("An error occurred canonicalizing .proto dirpath '{}'", protos_dirpath.display()))?;

    let input_dir_entries = read_dir(&protos_dirpath)
        .context(format!("An error occurred reading the files in .proto dirpath '{}'", protos_dirpath.display()))?;
    let mut proto_filepaths: Vec<PathBuf> = Vec::new();
    for (idx, dir_entry_result) in input_dir_entries.enumerate() {
        let dir_entry = dir_entry_result
            .context(format!("An error occurred unwrapping .proto filepath at idx '{}'", idx))?;
        let metadata = dir_entry.metadata()
            .context(format!("Could not get metadata for dir entry '{}'", dir_entry.path().display()))?;
        if metadata.is_file() {
            println!("cargo:rerun-if-changed={}", dir_entry.path().display());
            proto_filepaths.push(dir_entry.path());
        }
    }

    let out_dirpath = current_dirpath.join("src/core_api_bindings");
    let out_dirpath = out_dirpath.canonicalize()
        .context(format!("An error occurred canonicalizing output dirpath '{}'", out_dirpath.display()))?;
    let output_dir_entries = read_dir(&out_dirpath)
        .context(format!("An error occurred reading the files in output dirpath '{}'", out_dirpath.display()))?;
    for (idx, dir_entry_result) in output_dir_entries.enumerate() {
        let dir_entry = dir_entry_result
            .context(format!("An error occurred unwrapping dir entry in output dirpath at idx '{}'", idx))?;
        let metadata = dir_entry.metadata()
            .context(format!("Could not get metadata for dir entry '{}'", dir_entry.path().display()))?;
        let is_rust_file = matches!(dir_entry.file_name(), RUST_FILE_NAME_PATTERN);
        if metadata.is_file() && is_rust_file {
            remove_file(dir_entry.path())
                .context(format!(
                    "An error occurred removing output directory file '{}'",
                    dir_entry.path().display(),
                ))?;
        }
    }

    let filepath = protos_dirpath.join("test_execution_service.proto");
    let filepath = filepath.canonicalize()
        .context(format!("An error occurred canonicalizing .proto filepath '{}'", filepath.display()))?;
    

    // NOTE: We have to do the Protobuf binding generation here, rather than in the "regenerate-protobufs" Bash script,
    // because we're using Tonic/Prost, which is only a Rust library (doesn't have a `protoc` plugin). The grpc-protobuf crate
    // does have a `protoc` plugin, but it's not maintained frequently so we don't use it
    tonic_build::configure()
        .out_dir(out_dirpath)
        .compile(&proto_filepaths, &[protos_dirpath])
        .context("An error occurred generating the .proto files")?;
    Ok(())
}
use std::{env::current_dir, error::Error, ffi::OsString, fs::{read_dir, remove_file, write}, io::empty, path::{Path, PathBuf}, process::exit};
use anyhow::{Context, Result};
use regex::Regex;

const RUST_FILE_NAME_REGEX: &str = r"^.*\.rs$";
const EMPTY_FILE_NEEDING_DELETION_FILENAME: &str = "google.protobuf.rs";
const MOD_FILENAME: &str = "mod.rs";

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
    let rust_file_pattern = Regex::new(RUST_FILE_NAME_REGEX)
        .context(format!("An error occurred compiling Rust filename pattern '{}'", RUST_FILE_NAME_REGEX))?;
    for (idx, dir_entry_result) in output_dir_entries.enumerate() {
        let dir_entry = dir_entry_result
            .context(format!("An error occurred unwrapping dir entry in output dirpath at idx '{}'", idx))?;
        let metadata = dir_entry.metadata()
            .context(format!("Could not get metadata for dir entry '{}'", dir_entry.path().display()))?;
        let filename = dir_entry.file_name();
        let filename_unicode = filename.to_str()
            .context(format!("Could not convert file name of path '{}' to a Unicode string", dir_entry.path().display()))?;
        let is_rust_file = rust_file_pattern.is_match(filename_unicode);
        if metadata.is_file() && is_rust_file {
            remove_file(dir_entry.path())
                .context(format!(
                    "An error occurred removing output directory file '{}'",
                    dir_entry.path().display(),
                ))?;
        }
    }

    // NOTE: We have to do the Protobuf binding generation here, rather than in the "regenerate-protobufs" Bash script,
    // because we're using Tonic/Prost, which is only a Rust library (doesn't have a `protoc` plugin). The grpc-protobuf crate
    // does have a `protoc` plugin, but it's not maintained frequently so we don't use it
    tonic_build::configure()
        .out_dir(&out_dirpath)
        .compile(&proto_filepaths, &[protos_dirpath])
        .context("An error occurred generating the .proto files")?;


    // Due to https://github.com/danburkert/prost/issues/228, we get an empty file named "google.protobuf.rs" that
    // needs to be deleted after every build
    let empty_file_filepath = out_dirpath.join(EMPTY_FILE_NEEDING_DELETION_FILENAME).canonicalize()
        .context("An error occurred canonicalizing the full path of empty file needing deletion")?;
    remove_file(&empty_file_filepath)
        .context(format!("Failed to remove empty file '{}'", empty_file_filepath.display()))?;

    // Get list of modules generated...
    let output_dir_entries = read_dir(&out_dirpath)
        .context(format!("An error occurred reading the files in output dirpath '{}'", out_dirpath.display()))?;
    let mut mod_file_lines: Vec<String> = Vec::new();
    for (idx, dir_entry_result) in output_dir_entries.enumerate() {
        let dir_entry = dir_entry_result
            .context(format!("An error occurred unwrapping generated dir entry in output dirpath at idx '{}'", idx))?;
        let dir_entry_path = dir_entry.path();
        let metadata = dir_entry.metadata()
            .context(format!("Could not get metadata for generated dir entry '{}'", dir_entry_path.display()))?;


        let filename = dir_entry.file_name();
        let filename_unicode = filename.to_str()
            .context(format!("Could not convert file name of path '{}' to a Unicode string", dir_entry_path.display()))?;
        let is_rust_file = rust_file_pattern.is_match(filename_unicode);
        if metadata.is_file() && is_rust_file {
            let filename_without_ext_osstr = dir_entry_path.file_stem()
                .context("Failed to get filename without extension for file ")?;
            let filename_without_ext_str = filename_without_ext_osstr.to_str()
                .context(format!("Could not convert OsString for file '{}' to regular string", dir_entry_path.display()))?;
            mod_file_lines.push(
                format!("pub mod {};", filename_without_ext_str)
            );
        }
    }

    // ...and write those modules to a mod.rs file
    let mod_filepath = out_dirpath.join(MOD_FILENAME);
    let data = mod_file_lines.join("\n");
    write(&mod_filepath, data)
        .context(format!("Failed to write modules to {} file", MOD_FILENAME))?;
    Ok(())
}
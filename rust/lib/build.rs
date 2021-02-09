use std::{env::current_dir, error::Error, fs::read_dir, process::exit};
use anyhow::{Context, Result};

fn main() -> Result<()> {
    let current_dirpath = current_dir()
        .context("Couldn't get current directory")?;

    let root_dirpath = current_dirpath.join("../..");
    let root_dirpath = root_dirpath.canonicalize()
        .context(format!("An error occurred canonicalizing root dirpath {}", root_dirpath.display()))?;

    let protos_dirpath = root_dirpath.join("core-api");
    let protos_dirpath = protos_dirpath.canonicalize()
        .context(format!("An error occurred canonicalizing .proto dirpath '{}'", protos_dirpath.display()))?;

    let dir_entries = read_dir(&protos_dirpath)
        .context(format!("An error occurred reading the files in .proto dirpath '{}'", protos_dirpath.display()))?;
    for (idx, dir_entry_result) in dir_entries.enumerate() {
        let dir_entry = dir_entry_result
            .context(format!("An error occurred unwrapping .proto filepath at idx '{}'", idx))?;
        let metadata = dir_entry.metadata()
            .context(format!("Could not get metadata for dir entry '{}'", dir_entry.path().display()))?;
        if metadata.is_file() {
            println!("cargo:rerun-if-changed={}", dir_entry.path().display());
        }
    }

    let out_dirpath = current_dirpath.join("src/core_api_bindings");
    let out_dirpath = out_dirpath.canonicalize()
        .context(format!("An error occurred canonicalizing output dirpath '{}'", out_dirpath.display()))?;

    

    let filepath = protos_dirpath.join("test_execution_service.proto");
    let filepath = filepath.canonicalize()
        .context(format!("An error occurred canonicalizing .proto filepath '{}'", filepath.display()))?;
    

    tonic_build::configure()
        .out_dir(out_dirpath)
        .compile(&[&filepath], &[&protos_dirpath])
        .context("An error occurred generating the .proto files")?;
    /*
    let mut compiler = prost_build::Config::default();
    compiler.out_dir(out_dirpath);
    compiler.compile_protos(&[&filepath], &[&protos_dirpath])?;
    */
    Ok(())
}
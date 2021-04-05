Kurtosis Versioning & Upgrading
===============================
There are two versions that you'll need to pay attention to inside your Kurtosis testsuite:

1. The version of Kurtosis Core, which represents the main machinery that powers testsuite execution
1. The version of the Kurtosis Lib, which contains language-specific bindings that let your testsuite interact with Kurtosis Core

These versions are stored in the following spots:

* **Kurtosis Core:** Stored inside the scripts in the `.kurtosis` directory in the root of your testsuite repo. Upgrading Kurtosis Core means replacing the contents of your `.kurtosis` directory with the scripts from the version of your choice, found [here](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=dist/).
* **Kurtosis Lib:** Stored as a version in the dependency manager of your testsuite's language. E.g. upgrading the Kurtosis lib dependency for a...
    * ...Java testsuite would mean updating the `build.gradle` or `pom.xml` file with the new Lib version
    * ...Go testsuite would mean updating the `go.mod` file with the new Lib version
    * ...Rust testsuite would mean updating the `Cargo.toml` file with thew new Lib version

Because the Kurtosis Lib really contains bindings for connecting to Kurtosis Core, the version of Kurtosis Lib used dictates which version of Kurtosis Core you'll need. You'll need to make sure that your version of Kurtosis Lib is compatible with the Kurtosis Core scripts inside your `.kurtosis` directory. To see which version of Kurtosis Core your Kurtosis Lib is compatible with, look for the "Breaking Changes" section of [the Kurtosis Lib changelog](./changelog.md), which will have a message like "Upgraded to Kurtosis Core v1.10". This indicates that you must replace the contents of your `.kurtosis` directory with [the scripts from Kurtosis Core v1.10](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=dist/).

_FINAL NOTE: we know this process isn't as smooth as it could be. We're investigating making the Kurtosis Core version entirely transparent to you, so that you only need to think about the Kurtosis Lib version._

---

[Back to index](https://docs.kurtosistech.com)

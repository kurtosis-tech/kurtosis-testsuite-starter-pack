Kurtosis Libs
==================
This repo contains:

1. Libraries in various languages for interacting with Kurtosis Core, which are used to write testsuites
1. Example implementations of testsuites in each langauge
1. Infrastructure for bootstrapping a new testsuite, that you can use to create your own customized testsuite

### Testsuite Quickstart
Prerequisites:
* `git` installed on your machine
* `docker` installed on your machine

<!-- TODO embed the supported-languages file in this Markdown -->
Steps:
1. Clone this repo's `master` branch: `git clone --single-branch --branch master THIS_REPO_URL`
1. View the `supported-languages.txt` file and choose the language you'd like your testsuite in
1. Run `bootstrap/bootstrap.sh YOUR_LANG /path/to/output/testsuite/dir`

### Developing Libraries
Each library needs to talk with Kurtosis Core, and the Kurtosis Core API is defined via Protobuf. Rather than storing the Protobufs in Git submodules (which add significant complexity), the `.proto` files are simply copied from the relevant version of Kurtosis Core. In the future, we can move to a more productized solution.

To regenerate the bindings corresponding to the Protobuf files, use the `scripts/regenerate-protobuf-output.sh` script.

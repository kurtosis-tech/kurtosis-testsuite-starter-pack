Kurtosis Testsuite Starter Pack
===============================
This repo contains:

1. Example testsuites in each language that Kurtosis supports
1. A `bootstrap.sh` script for creating your own testsuite in your language of choice

Testsuite Quickstart
--------------------
Prerequisites:
* A [Kurtosis user account](https://www.kurtosistech.com/sign-up)
* `git` installed on your machine
* `docker` installed on your machine

Quickstart steps:
1. Clone this repo's `master` branch: `git clone --single-branch --branch master git@github.com:kurtosis-tech/kurtosis-testsuite-starter-pack.git`
1. View [the supported languages](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/supported-languages.txt) and choose the language you'd like your testsuite in
1. Run `bootstrap/bootstrap.sh` and follow the helptext instructions to fill in the script arguments and bootstrap your repo
1. If you see error messages after running your new testsuite, check out [the guide for debugging failed tests](https://docs.kurtosistech.com/kurtosis-core/debugging-failed-tests) which contains solutions to common issues. If this still doesn't resolve your issue, feel free to ask for help in [the Kurtosis Discord server](https://discord.gg/6Jjp9c89z9)
1. If all tests are passing, you can [proceed to customizing your testsuite](https://docs.kurtosistech.com/kurtosis-core/testsuite-customization).

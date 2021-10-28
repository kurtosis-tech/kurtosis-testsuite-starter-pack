Kurtosis Testsuite Starter Pack
===============================
This repo contains:

1. Example testsuites in each language that Kurtosis supports
1. A `bootstrap.sh` script for creating your own testsuite in your language of choice

This repo is for users already familiar with Kurtosis. Brand-new users should go through [the onboarding experience](https://github.com/kurtosis-tech/kurtosis-onboarding-experience) rather than using this repo.

Bootstrapping A Testsuite
-------------------------
Prerequisites:

* The [Kurtosis CLI installed on your machine](https://docs.kurtosistech.com/installation.html)
* `git` installed on your machine
* `docker` installed & running on your machine

Bootstrap steps:

1. Clone this repo's `master` branch (the command can be copied by hovering and clicking the clipboard icon in the top-right): 
    ```
    repo_name="kurtosis-testsuite-starter-pack"
    destination="/tmp/${repo_name}"
    git clone --single-branch --branch master "https://github.com/kurtosis-tech/${repo_name}.git" "${destination}"
    cd "${destination}"
    ```
1. Run the following bootstrap command to see the helptext instructions that the bootstrapping script accepts:
    ```
    bootstrap/bootstrap.sh
    ````
1. Run the bootstrap script again with arguments appropriate to the testsuite you'd like to bootstrap, and follow the onscreen instructions
1. If you see error messages after running your new testsuite, check out [the guide for debugging failed tests](https://docs.kurtosistech.com/debugging-failed-tests) which contains solutions to common issues. If this still doesn't resolve your issue, feel free to ask for help in [the Kurtosis Discord server](https://discord.gg/6Jjp9c89z9)
1. If all tests are passing, you can [proceed to customizing your testsuite](https://docs.kurtosistech.com/testsuite-customization).

use std::collections::HashMap;

// ====================================================================================================
//                                    Config Object
// ====================================================================================================
// Docs available at https://docs.kurtosistech.com/kurtosis-libs/lib-documentation
pub struct ContainerRunConfig {
    entrypoint_override_args: Option<Vec<String>>,
    cmd_override_args: Option<Vec<String>>,
    environment_variable_overrides: HashMap<String, String>,
}

impl ContainerRunConfig {
    pub fn get_entrypoint_override_args(&self) -> &Option<Vec<String>> {
        return &self.entrypoint_override_args;
    }

    pub fn get_cmd_override_args(&self) -> &Option<Vec<String>> {
        return &self.cmd_override_args;
    }

    pub fn get_environment_variable_overrides(&self) -> &HashMap<String, String> {
        return &self.environment_variable_overrides;
    }
}



// ====================================================================================================
//                                      Builder
// ====================================================================================================
pub struct ContainerRunConfigBuilder {
    entrypoint_override_args: Option<Vec<String>>,
    cmd_override_args: Option<Vec<String>>,
    environment_variable_overrides: HashMap<String, String>,
}

impl ContainerRunConfigBuilder {
    pub fn new() -> ContainerRunConfigBuilder {
        return ContainerRunConfigBuilder{
            entrypoint_override_args: None,
            cmd_override_args: None,
            environment_variable_overrides: HashMap::new(),
        }
    }

    pub fn with_entrypoint_override(&mut self, args: Vec<String>) -> &mut ContainerRunConfigBuilder {
        self.entrypoint_override_args = Some(args);
        return self;
    }

    pub fn with_cmd_override(&mut self, args: Vec<String>) -> &mut ContainerRunConfigBuilder {
        self.cmd_override_args = Some(args);
        return self;
    }

    pub fn with_environment_variable_overrides(&mut self, env_vars: HashMap<String, String>) -> &mut ContainerRunConfigBuilder {
        self.environment_variable_overrides = env_vars;
        return self;
    }

    pub fn build(&self) -> ContainerRunConfig {
        return ContainerRunConfig{
            entrypoint_override_args: self.entrypoint_override_args.clone(),
            cmd_override_args: self.cmd_override_args.clone(),
            environment_variable_overrides: self.environment_variable_overrides.clone(),
        }
    }
}
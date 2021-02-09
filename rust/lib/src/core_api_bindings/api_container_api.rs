/// ==============================================================================================
///                                  Get Test Execution Info
/// ==============================================================================================
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TestExecutionInfo {
    /// Name of the test that the testsuite container should execute
    #[prost(string, tag = "1")]
    pub test_name: ::prost::alloc::string::String,
}
/// ==============================================================================================
///                                  Register Test Execution
/// ==============================================================================================
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RegisterTestExecutionArgs {
    /// TODO This should actually be unnecessary - we should pass in testsuite metadata at API container startup time,
    ///  so that registration just says "I'm starting" and the API container can look up the timeout
    #[prost(uint64, tag = "1")]
    pub timeout_seconds: u64,
}
/// ==============================================================================================
///                                     Register Service
/// ==============================================================================================
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RegisterServiceArgs {
    /// ID that will be used to identify the service going forward
    #[prost(string, tag = "1")]
    pub service_id: ::prost::alloc::string::String,
    /// If emptystring, the default partition ID will be used
    #[prost(string, tag = "2")]
    pub partition_id: ::prost::alloc::string::String,
    /// "Set" of files that the service needs and the API container should make available upon service start
    /// The key of the map is a user-meaningful identifier
    #[prost(map = "string, bool", tag = "3")]
    pub files_to_generate: ::std::collections::HashMap<::prost::alloc::string::String, bool>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RegisterServiceResponse {
    /// Mapping of user-created key in the request -> filepath (RELATIVE to the suite execution volume root!) where
    ///  the file was created
    #[prost(map = "string, string", tag = "1")]
    pub generated_files_relative_filepaths:
        ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    /// The IP address that the service will receive when it starts
    #[prost(string, tag = "2")]
    pub ip_addr: ::prost::alloc::string::String,
}
/// ==============================================================================================
///                                        Start Service
/// ==============================================================================================
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct StartServiceArgs {
    /// ID of the previously-registered service that should be started
    #[prost(string, tag = "1")]
    pub service_id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub docker_image: ::prost::alloc::string::String,
    /// "Set" of ports that the running service will listen on
    /// This is a string because it's Docker port specification syntax, e.g. "80" (default TCP) or "80/udp"
    #[prost(map = "string, bool", tag = "3")]
    pub used_ports: ::std::collections::HashMap<::prost::alloc::string::String, bool>,
    /// String array indicating the command that should be run inside the sevice's container on startup
    #[prost(string, repeated, tag = "4")]
    pub start_cmd_args: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// Docker environment variables that should be set in the service's container
    #[prost(map = "string, string", tag = "5")]
    pub docker_env_vars:
        ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    /// The full path where the API container should execute the suite execution volume on the service container
    #[prost(string, tag = "6")]
    pub suite_execution_vol_mnt_dirpath: ::prost::alloc::string::String,
    /// Mapping of artifact_url -> filepath_on_container_to_mount_artifact_contents
    #[prost(map = "string, string", tag = "7")]
    pub files_artifact_mount_dirpaths:
        ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
/// ==============================================================================================
///                                        Remove Service
/// ==============================================================================================
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RemoveServiceArgs {
    #[prost(string, tag = "1")]
    pub service_id: ::prost::alloc::string::String,
    /// How long to wait for the service to gracefully stop before hard killing it
    #[prost(uint64, tag = "2")]
    pub container_stop_timeout_seconds: u64,
}
/// ==============================================================================================
///                                          Repartition
/// ==============================================================================================
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RepartitionArgs {
    /// Definition of partitionId -> services that should be inside the partition after repartitioning
    #[prost(map = "string, message", tag = "1")]
    pub partition_services:
        ::std::collections::HashMap<::prost::alloc::string::String, PartitionServices>,
    /// Definition of partitionIdA -> partitionIdB -> information defining the connection between A <-> B
    #[prost(map = "string, message", tag = "2")]
    pub partition_connections:
        ::std::collections::HashMap<::prost::alloc::string::String, PartitionConnections>,
    /// Information about the default inter-partition connection to set up if one is not defined in the
    ///  partition connections map
    #[prost(message, optional, tag = "3")]
    pub default_connection: ::core::option::Option<PartitionConnectionInfo>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PartitionServices {
    /// "Set" of service IDs in partition
    #[prost(map = "string, bool", tag = "1")]
    pub service_id_set: ::std::collections::HashMap<::prost::alloc::string::String, bool>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PartitionConnections {
    #[prost(map = "string, message", tag = "1")]
    pub connection_info:
        ::std::collections::HashMap<::prost::alloc::string::String, PartitionConnectionInfo>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PartitionConnectionInfo {
    /// Whether network traffic is allowed between the two partitions
    #[prost(bool, tag = "1")]
    pub is_blocked: bool,
}
#[doc = r" Generated client implementations."]
pub mod test_execution_service_client {
    #![allow(unused_variables, dead_code, missing_docs)]
    use tonic::codegen::*;
    pub struct TestExecutionServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl TestExecutionServiceClient<tonic::transport::Channel> {
        #[doc = r" Attempt to create a new client by connecting to a given endpoint."]
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: std::convert::TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> TestExecutionServiceClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::BoxBody>,
        T::ResponseBody: Body + HttpBody + Send + 'static,
        T::Error: Into<StdError>,
        <T::ResponseBody as HttpBody>::Error: Into<StdError> + Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_interceptor(inner: T, interceptor: impl Into<tonic::Interceptor>) -> Self {
            let inner = tonic::client::Grpc::with_interceptor(inner, interceptor);
            Self { inner }
        }
        #[doc = " Returns detailed information to the testsuite about what it should do during test execution -"]
        #[doc = "  namely, what test it should run"]
        #[doc = " This method should be called first by the testsuite"]
        pub async fn get_test_execution_info(
            &mut self,
            request: impl tonic::IntoRequest<()>,
        ) -> Result<tonic::Response<super::TestExecutionInfo>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.TestExecutionService/GetTestExecutionInfo",
            );
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Registers that the testsuite is about to start executing test logic"]
        pub async fn register_test_execution(
            &mut self,
            request: impl tonic::IntoRequest<super::RegisterTestExecutionArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.TestExecutionService/RegisterTestExecution",
            );
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Registers a service with the API container but doesn't start the container for it"]
        pub async fn register_service(
            &mut self,
            request: impl tonic::IntoRequest<super::RegisterServiceArgs>,
        ) -> Result<tonic::Response<super::RegisterServiceResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.TestExecutionService/RegisterService",
            );
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Starts a previously-registered service by creating a Docker container for it"]
        pub async fn start_service(
            &mut self,
            request: impl tonic::IntoRequest<super::StartServiceArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.TestExecutionService/StartService",
            );
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Instructs the API container to remove the given service"]
        pub async fn remove_service(
            &mut self,
            request: impl tonic::IntoRequest<super::RemoveServiceArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.TestExecutionService/RemoveService",
            );
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Instructs the API container to repartition the test network"]
        pub async fn repartition(
            &mut self,
            request: impl tonic::IntoRequest<super::RepartitionArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/api_container_api.TestExecutionService/Repartition",
            );
            self.inner.unary(request.into_request(), path, codec).await
        }
    }
    impl<T: Clone> Clone for TestExecutionServiceClient<T> {
        fn clone(&self) -> Self {
            Self {
                inner: self.inner.clone(),
            }
        }
    }
    impl<T> std::fmt::Debug for TestExecutionServiceClient<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "TestExecutionServiceClient {{ ... }}")
        }
    }
}
#[doc = r" Generated server implementations."]
pub mod test_execution_service_server {
    #![allow(unused_variables, dead_code, missing_docs)]
    use tonic::codegen::*;
    #[doc = "Generated trait containing gRPC methods that should be implemented for use with TestExecutionServiceServer."]
    #[async_trait]
    pub trait TestExecutionService: Send + Sync + 'static {
        #[doc = " Returns detailed information to the testsuite about what it should do during test execution -"]
        #[doc = "  namely, what test it should run"]
        #[doc = " This method should be called first by the testsuite"]
        async fn get_test_execution_info(
            &self,
            request: tonic::Request<()>,
        ) -> Result<tonic::Response<super::TestExecutionInfo>, tonic::Status>;
        #[doc = " Registers that the testsuite is about to start executing test logic"]
        async fn register_test_execution(
            &self,
            request: tonic::Request<super::RegisterTestExecutionArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status>;
        #[doc = " Registers a service with the API container but doesn't start the container for it"]
        async fn register_service(
            &self,
            request: tonic::Request<super::RegisterServiceArgs>,
        ) -> Result<tonic::Response<super::RegisterServiceResponse>, tonic::Status>;
        #[doc = " Starts a previously-registered service by creating a Docker container for it"]
        async fn start_service(
            &self,
            request: tonic::Request<super::StartServiceArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status>;
        #[doc = " Instructs the API container to remove the given service"]
        async fn remove_service(
            &self,
            request: tonic::Request<super::RemoveServiceArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status>;
        #[doc = " Instructs the API container to repartition the test network"]
        async fn repartition(
            &self,
            request: tonic::Request<super::RepartitionArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status>;
    }
    #[derive(Debug)]
    pub struct TestExecutionServiceServer<T: TestExecutionService> {
        inner: _Inner<T>,
    }
    struct _Inner<T>(Arc<T>, Option<tonic::Interceptor>);
    impl<T: TestExecutionService> TestExecutionServiceServer<T> {
        pub fn new(inner: T) -> Self {
            let inner = Arc::new(inner);
            let inner = _Inner(inner, None);
            Self { inner }
        }
        pub fn with_interceptor(inner: T, interceptor: impl Into<tonic::Interceptor>) -> Self {
            let inner = Arc::new(inner);
            let inner = _Inner(inner, Some(interceptor.into()));
            Self { inner }
        }
    }
    impl<T, B> Service<http::Request<B>> for TestExecutionServiceServer<T>
    where
        T: TestExecutionService,
        B: HttpBody + Send + Sync + 'static,
        B::Error: Into<StdError> + Send + 'static,
    {
        type Response = http::Response<tonic::body::BoxBody>;
        type Error = Never;
        type Future = BoxFuture<Self::Response, Self::Error>;
        fn poll_ready(&mut self, _cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
            Poll::Ready(Ok(()))
        }
        fn call(&mut self, req: http::Request<B>) -> Self::Future {
            let inner = self.inner.clone();
            match req.uri().path() {
                "/api_container_api.TestExecutionService/GetTestExecutionInfo" => {
                    #[allow(non_camel_case_types)]
                    struct GetTestExecutionInfoSvc<T: TestExecutionService>(pub Arc<T>);
                    impl<T: TestExecutionService> tonic::server::UnaryService<()> for GetTestExecutionInfoSvc<T> {
                        type Response = super::TestExecutionInfo;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(&mut self, request: tonic::Request<()>) -> Self::Future {
                            let inner = self.0.clone();
                            let fut =
                                async move { (*inner).get_test_execution_info(request).await };
                            Box::pin(fut)
                        }
                    }
                    let inner = self.inner.clone();
                    let fut = async move {
                        let interceptor = inner.1.clone();
                        let inner = inner.0;
                        let method = GetTestExecutionInfoSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = if let Some(interceptor) = interceptor {
                            tonic::server::Grpc::with_interceptor(codec, interceptor)
                        } else {
                            tonic::server::Grpc::new(codec)
                        };
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.TestExecutionService/RegisterTestExecution" => {
                    #[allow(non_camel_case_types)]
                    struct RegisterTestExecutionSvc<T: TestExecutionService>(pub Arc<T>);
                    impl<T: TestExecutionService>
                        tonic::server::UnaryService<super::RegisterTestExecutionArgs>
                        for RegisterTestExecutionSvc<T>
                    {
                        type Response = ();
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RegisterTestExecutionArgs>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut =
                                async move { (*inner).register_test_execution(request).await };
                            Box::pin(fut)
                        }
                    }
                    let inner = self.inner.clone();
                    let fut = async move {
                        let interceptor = inner.1.clone();
                        let inner = inner.0;
                        let method = RegisterTestExecutionSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = if let Some(interceptor) = interceptor {
                            tonic::server::Grpc::with_interceptor(codec, interceptor)
                        } else {
                            tonic::server::Grpc::new(codec)
                        };
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.TestExecutionService/RegisterService" => {
                    #[allow(non_camel_case_types)]
                    struct RegisterServiceSvc<T: TestExecutionService>(pub Arc<T>);
                    impl<T: TestExecutionService>
                        tonic::server::UnaryService<super::RegisterServiceArgs>
                        for RegisterServiceSvc<T>
                    {
                        type Response = super::RegisterServiceResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RegisterServiceArgs>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).register_service(request).await };
                            Box::pin(fut)
                        }
                    }
                    let inner = self.inner.clone();
                    let fut = async move {
                        let interceptor = inner.1.clone();
                        let inner = inner.0;
                        let method = RegisterServiceSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = if let Some(interceptor) = interceptor {
                            tonic::server::Grpc::with_interceptor(codec, interceptor)
                        } else {
                            tonic::server::Grpc::new(codec)
                        };
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.TestExecutionService/StartService" => {
                    #[allow(non_camel_case_types)]
                    struct StartServiceSvc<T: TestExecutionService>(pub Arc<T>);
                    impl<T: TestExecutionService>
                        tonic::server::UnaryService<super::StartServiceArgs>
                        for StartServiceSvc<T>
                    {
                        type Response = ();
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::StartServiceArgs>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).start_service(request).await };
                            Box::pin(fut)
                        }
                    }
                    let inner = self.inner.clone();
                    let fut = async move {
                        let interceptor = inner.1.clone();
                        let inner = inner.0;
                        let method = StartServiceSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = if let Some(interceptor) = interceptor {
                            tonic::server::Grpc::with_interceptor(codec, interceptor)
                        } else {
                            tonic::server::Grpc::new(codec)
                        };
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.TestExecutionService/RemoveService" => {
                    #[allow(non_camel_case_types)]
                    struct RemoveServiceSvc<T: TestExecutionService>(pub Arc<T>);
                    impl<T: TestExecutionService>
                        tonic::server::UnaryService<super::RemoveServiceArgs>
                        for RemoveServiceSvc<T>
                    {
                        type Response = ();
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RemoveServiceArgs>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).remove_service(request).await };
                            Box::pin(fut)
                        }
                    }
                    let inner = self.inner.clone();
                    let fut = async move {
                        let interceptor = inner.1.clone();
                        let inner = inner.0;
                        let method = RemoveServiceSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = if let Some(interceptor) = interceptor {
                            tonic::server::Grpc::with_interceptor(codec, interceptor)
                        } else {
                            tonic::server::Grpc::new(codec)
                        };
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/api_container_api.TestExecutionService/Repartition" => {
                    #[allow(non_camel_case_types)]
                    struct RepartitionSvc<T: TestExecutionService>(pub Arc<T>);
                    impl<T: TestExecutionService>
                        tonic::server::UnaryService<super::RepartitionArgs> for RepartitionSvc<T>
                    {
                        type Response = ();
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RepartitionArgs>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).repartition(request).await };
                            Box::pin(fut)
                        }
                    }
                    let inner = self.inner.clone();
                    let fut = async move {
                        let interceptor = inner.1.clone();
                        let inner = inner.0;
                        let method = RepartitionSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = if let Some(interceptor) = interceptor {
                            tonic::server::Grpc::with_interceptor(codec, interceptor)
                        } else {
                            tonic::server::Grpc::new(codec)
                        };
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                _ => Box::pin(async move {
                    Ok(http::Response::builder()
                        .status(200)
                        .header("grpc-status", "12")
                        .header("content-type", "application/grpc")
                        .body(tonic::body::BoxBody::empty())
                        .unwrap())
                }),
            }
        }
    }
    impl<T: TestExecutionService> Clone for TestExecutionServiceServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self { inner }
        }
    }
    impl<T: TestExecutionService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(self.0.clone(), self.1.clone())
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: TestExecutionService> tonic::transport::NamedService for TestExecutionServiceServer<T> {
        const NAME: &'static str = "api_container_api.TestExecutionService";
    }
}

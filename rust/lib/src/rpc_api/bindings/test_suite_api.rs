/// ====================================================================================================
///                                       GetTestSuiteMetadata
/// ====================================================================================================
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TestSuiteMetadata {
    /// Mapping of testName -> testMetadata
    #[prost(map = "string, message", tag = "1")]
    pub test_metadata: ::std::collections::HashMap<::prost::alloc::string::String, TestMetadata>,
    #[prost(uint32, tag = "2")]
    pub network_width_bits: u32,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TestMetadata {
    #[prost(bool, tag = "1")]
    pub is_partitioning_enabled: bool,
    /// "Set" of artifact URLs used by the test
    #[prost(map = "string, bool", tag = "2")]
    pub used_artifact_urls: ::std::collections::HashMap<::prost::alloc::string::String, bool>,
    #[prost(uint32, tag = "3")]
    pub test_setup_timeout_in_seconds: u32,
    #[prost(uint32, tag = "4")]
    pub test_run_timeout_in_seconds: u32,
}
/// ====================================================================================================
///                                       SetupTest
/// ====================================================================================================
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SetupTestArgs {
    #[prost(string, tag = "1")]
    pub test_name: ::prost::alloc::string::String,
}
#[doc = r" Generated client implementations."]
pub mod test_suite_service_client {
    #![allow(unused_variables, dead_code, missing_docs)]
    use tonic::codegen::*;
    pub struct TestSuiteServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl TestSuiteServiceClient<tonic::transport::Channel> {
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
    impl<T> TestSuiteServiceClient<T>
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
        pub async fn get_test_suite_metadata(
            &mut self,
            request: impl tonic::IntoRequest<()>,
        ) -> Result<tonic::Response<super::TestSuiteMetadata>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/test_suite_api.TestSuiteService/GetTestSuiteMetadata",
            );
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn setup_test(
            &mut self,
            request: impl tonic::IntoRequest<super::SetupTestArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/test_suite_api.TestSuiteService/SetupTest");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " We don't need args dictating what test to run because SetupTest already indicates it (and it wouldn't make"]
        #[doc = "  sense to setup one test and run another)"]
        pub async fn run_test(
            &mut self,
            request: impl tonic::IntoRequest<()>,
        ) -> Result<tonic::Response<()>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/test_suite_api.TestSuiteService/RunTest");
            self.inner.unary(request.into_request(), path, codec).await
        }
    }
    impl<T: Clone> Clone for TestSuiteServiceClient<T> {
        fn clone(&self) -> Self {
            Self {
                inner: self.inner.clone(),
            }
        }
    }
    impl<T> std::fmt::Debug for TestSuiteServiceClient<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "TestSuiteServiceClient {{ ... }}")
        }
    }
}
#[doc = r" Generated server implementations."]
pub mod test_suite_service_server {
    #![allow(unused_variables, dead_code, missing_docs)]
    use tonic::codegen::*;
    #[doc = "Generated trait containing gRPC methods that should be implemented for use with TestSuiteServiceServer."]
    #[async_trait]
    pub trait TestSuiteService: Send + Sync + 'static {
        async fn get_test_suite_metadata(
            &self,
            request: tonic::Request<()>,
        ) -> Result<tonic::Response<super::TestSuiteMetadata>, tonic::Status>;
        async fn setup_test(
            &self,
            request: tonic::Request<super::SetupTestArgs>,
        ) -> Result<tonic::Response<()>, tonic::Status>;
        #[doc = " We don't need args dictating what test to run because SetupTest already indicates it (and it wouldn't make"]
        #[doc = "  sense to setup one test and run another)"]
        async fn run_test(
            &self,
            request: tonic::Request<()>,
        ) -> Result<tonic::Response<()>, tonic::Status>;
    }
    #[derive(Debug)]
    pub struct TestSuiteServiceServer<T: TestSuiteService> {
        inner: _Inner<T>,
    }
    struct _Inner<T>(Arc<T>, Option<tonic::Interceptor>);
    impl<T: TestSuiteService> TestSuiteServiceServer<T> {
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
    impl<T, B> Service<http::Request<B>> for TestSuiteServiceServer<T>
    where
        T: TestSuiteService,
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
                "/test_suite_api.TestSuiteService/GetTestSuiteMetadata" => {
                    #[allow(non_camel_case_types)]
                    struct GetTestSuiteMetadataSvc<T: TestSuiteService>(pub Arc<T>);
                    impl<T: TestSuiteService> tonic::server::UnaryService<()> for GetTestSuiteMetadataSvc<T> {
                        type Response = super::TestSuiteMetadata;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(&mut self, request: tonic::Request<()>) -> Self::Future {
                            let inner = self.0.clone();
                            let fut =
                                async move { (*inner).get_test_suite_metadata(request).await };
                            Box::pin(fut)
                        }
                    }
                    let inner = self.inner.clone();
                    let fut = async move {
                        let interceptor = inner.1.clone();
                        let inner = inner.0;
                        let method = GetTestSuiteMetadataSvc(inner);
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
                "/test_suite_api.TestSuiteService/SetupTest" => {
                    #[allow(non_camel_case_types)]
                    struct SetupTestSvc<T: TestSuiteService>(pub Arc<T>);
                    impl<T: TestSuiteService> tonic::server::UnaryService<super::SetupTestArgs> for SetupTestSvc<T> {
                        type Response = ();
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::SetupTestArgs>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).setup_test(request).await };
                            Box::pin(fut)
                        }
                    }
                    let inner = self.inner.clone();
                    let fut = async move {
                        let interceptor = inner.1.clone();
                        let inner = inner.0;
                        let method = SetupTestSvc(inner);
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
                "/test_suite_api.TestSuiteService/RunTest" => {
                    #[allow(non_camel_case_types)]
                    struct RunTestSvc<T: TestSuiteService>(pub Arc<T>);
                    impl<T: TestSuiteService> tonic::server::UnaryService<()> for RunTestSvc<T> {
                        type Response = ();
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(&mut self, request: tonic::Request<()>) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).run_test(request).await };
                            Box::pin(fut)
                        }
                    }
                    let inner = self.inner.clone();
                    let fut = async move {
                        let interceptor = inner.1.clone();
                        let inner = inner.0;
                        let method = RunTestSvc(inner);
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
    impl<T: TestSuiteService> Clone for TestSuiteServiceServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self { inner }
        }
    }
    impl<T: TestSuiteService> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(self.0.clone(), self.1.clone())
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: TestSuiteService> tonic::transport::NamedService for TestSuiteServiceServer<T> {
        const NAME: &'static str = "test_suite_api.TestSuiteService";
    }
}

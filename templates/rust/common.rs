
use cdrs::authenticators::StaticPasswordAuthenticator;
use cdrs::cluster::session::{new as new_session, Session};
use cdrs::cluster::{ClusterTcpConfig, NodeTcpConfigBuilder, TcpConnectionPool};
use cdrs::load_balancing::RoundRobin;
use cdrs::query::*;

use cdrs::frame::{IntoBytes, Frame};
use cdrs::types::from_cdrs::FromCDRSByName;
use cdrs::types::prelude::*;
use cdrs::frame::frame_error::CDRSError;
use cdrs::Error as DriverError;

// pub type CurrentSession = Session<RoundRobin<TcpConnectionPool<StaticPasswordAuthenticator>>>;

// Our simplified proxy session caller (cdrs has a very complex type system, with this
// trait the source code gets much more simplified without needing for complex generics
// all over place
pub trait FCQueryExecutor {
    /// Executes a query with bounded values (either with or without names).
    fn query_with_values<Q: ToString, V: Into<QueryValues>>(
        &self,
        query: Q,
        values: V,
    ) -> cdrs::error::Result<Frame>
        where
            Self: Sized;
}

#[derive(Debug, Clone)]
pub struct WhereClause {
    // pub condition: &'static str,
    pub condition: String,
    pub args: Value,
}

#[derive(Debug)]
pub enum CWError {
    Server(CDRSError),
    General,
    Driver(DriverError),
    InvalidCQL,
    NotFound,
}

impl From<DriverError> for CWError {
    fn from(err: Error) -> Self {
        match err {
            DriverError::Server(serr) => CWError::Server(serr),
            _ => CWError::Driver(err)
        }
    }
}


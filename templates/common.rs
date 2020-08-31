
use cdrs::authenticators::StaticPasswordAuthenticator;
use cdrs::cluster::session::{new as new_session, Session};
use cdrs::cluster::{ClusterTcpConfig, NodeTcpConfigBuilder, TcpConnectionPool};
use cdrs::load_balancing::RoundRobin;
use cdrs::query::*;

use cdrs::frame::IntoBytes;
use cdrs::types::from_cdrs::FromCDRSByName;
use cdrs::types::prelude::*;
use cdrs::frame::frame_error::CDRSError;
use cdrs::Error as DriverError;

pub type CurrentSession = Session<RoundRobin<TcpConnectionPool<StaticPasswordAuthenticator>>>;

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


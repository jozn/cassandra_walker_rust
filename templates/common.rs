
use cdrs::authenticators::StaticPasswordAuthenticator;
use cdrs::cluster::session::{new as new_session, Session};
use cdrs::cluster::{ClusterTcpConfig, NodeTcpConfigBuilder, TcpConnectionPool};
use cdrs::load_balancing::RoundRobin;
use cdrs::query::*;

use cdrs::frame::IntoBytes;
use cdrs::types::from_cdrs::FromCDRSByName;
use cdrs::types::prelude::*;

pub type CurrentSession = Session<RoundRobin<TcpConnectionPool<StaticPasswordAuthenticator>>>;

#[derive(Debug, Clone)]
pub struct WhereClause {
    // pub condition: &'static str,
    pub condition: String,
    pub args: Value,
}

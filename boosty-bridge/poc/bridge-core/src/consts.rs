use std::time::Duration;

/// Time to wait before retrying a failed connection.
pub const RETRY_TIMEOUT: Duration = Duration::from_secs(5);

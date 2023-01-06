use std::{
    sync::{
        atomic::{AtomicBool, Ordering},
        Arc,
    },
    time::Duration,
};

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use parking_lot::Mutex;
use tokio::sync::oneshot;

use crate::TimeSource;

pub struct RealTimeSource;

#[async_trait]
impl TimeSource for RealTimeSource {
    fn now(&self) -> DateTime<Utc> {
        Utc::now()
    }

    async fn sleep(&self, duration: Duration) {
        tokio::time::sleep(duration).await
    }

    async fn sleep_until(&self, until: DateTime<Utc>) {
        let duration = until - self.now();
        if duration > chrono::Duration::zero() {
            let duration = duration.to_std().expect("should be non-negative");

            tokio::time::sleep(duration).await;
        }
    }
}

#[derive(Clone)]
pub struct MockTimeSource {
    inner: Arc<MockTimeSourceInner>,
}

impl MockTimeSource {
    pub fn new(initial: DateTime<Utc>) -> Self {
        Self {
            inner: Arc::new(MockTimeSourceInner {
                current: Mutex::new(initial),
                pending: Default::default(),
                auto_advance: AtomicBool::new(true),
            }),
        }
    }

    async fn poll_timers(&self) {
        {
            let mut pending = self.inner.pending.lock();
            let now = self.now();

            pending.retain_mut(|item| {
                if item.deadline <= now {
                    let trigger = item.trigger.take().expect("trigger should be available");
                    trigger.send(()).ok();
                    false
                } else {
                    true
                }
            });
        }

        tokio::task::yield_now().await
    }

    pub async fn set(&self, time: DateTime<Utc>) {
        let current = self.now();
        assert!(current <= time, "cannot advance time backwards");

        *self.inner.current.lock() = time;

        self.poll_timers().await;
    }

    pub async fn advance(&self, duration: Duration) {
        let time = self.now() + chrono::Duration::from_std(duration).expect("invalid duration");
        self.set(time).await;
    }

    pub async fn advance_all(&self) {
        {
            let mut pending = self.inner.pending.lock();
            let mut max_time = self.now();
            for item in pending.drain(..) {
                max_time = max_time.max(item.deadline);
                item.trigger
                    .expect("trigger should be available")
                    .send(())
                    .ok();
            }

            *self.inner.current.lock() = max_time;
        }

        tokio::task::yield_now().await
    }

    pub fn set_auto_advance(&self, auto_advance: bool) {
        self.inner
            .auto_advance
            .store(auto_advance, Ordering::SeqCst);
    }
}

struct PendingTimer {
    deadline: DateTime<Utc>,
    trigger: Option<oneshot::Sender<()>>,
}

struct MockTimeSourceInner {
    current: Mutex<DateTime<Utc>>,
    pending: Mutex<Vec<PendingTimer>>,
    auto_advance: AtomicBool,
}

#[async_trait]
impl TimeSource for MockTimeSource {
    fn now(&self) -> DateTime<Utc> {
        *self.inner.current.lock()
    }

    async fn sleep(&self, duration: Duration) {
        let deadline = self.now() + chrono::Duration::from_std(duration).expect("invalid duration");

        if self.inner.auto_advance.load(Ordering::SeqCst) {
            self.set(deadline).await;

            return;
        }

        let (tx, rx) = oneshot::channel();

        let pending = PendingTimer {
            deadline,
            trigger: Some(tx),
        };

        self.inner.pending.lock().push(pending);

        rx.await.expect("timer handle dropped unexpectedly");
    }

    async fn sleep_until(&self, until: DateTime<Utc>) {
        if until <= self.now() {
            return;
        }

        let duration = until - self.now();

        self.sleep(duration.to_std().expect("invalid duration"))
            .await;
    }
}

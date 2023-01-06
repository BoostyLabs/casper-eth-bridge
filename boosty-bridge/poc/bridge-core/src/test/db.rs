use std::{path::PathBuf, time::Duration};

use pg_embed::{
    pg_fetch::{PgFetchSettings, PG_V13},
    postgres::{PgEmbed, PgSettings},
};

use crate::db::Database;

pub async fn init_pg(port: u16) -> PgEmbed {
    println!("starting pg on port {port}");

    let settings = PgSettings {
        database_dir: PathBuf::from("data/db"),
        port: port as i16,
        user: "postgres".into(),
        password: "password".into(),
        auth_method: pg_embed::pg_enums::PgAuthMethod::Plain,
        persistent: false,
        timeout: Some(Duration::from_secs(15)),
        migration_dir: None,
    };

    let fetch_settings = PgFetchSettings {
        version: PG_V13,
        ..Default::default()
    };

    let mut pg = PgEmbed::new(settings, fetch_settings).await.unwrap();

    pg.setup().await.unwrap();
    pg.start_db().await.unwrap();
    pg.create_database("golden_gate").await.unwrap();

    pg
}

pub async fn init_db(port: u16) -> Database {
    Database::connect(db_config(port)).await.unwrap()
}

pub fn db_config(port: u16) -> crate::db::Config {
    crate::db::Config {
        host: "127.0.0.1".to_string(),
        port,
        user: "postgres".to_string(),
        pass: "password".to_string(),
        dbname: "golden_gate".to_string(),
    }
}

use std::time::Duration;

use primitive_types::U256;

use crate::types::BridgeEvent;

use super::TestContext;

#[tokio::test]
async fn single_transfer() -> anyhow::Result<()> {
    let ctx = TestContext::create().await;
    ctx.time_source.set_auto_advance(true);
    let token = ctx.create_test_token().await;

    ctx.bridge.load_tokens().await.unwrap();

    let evm_addr = ctx.evm.generate_address();
    let cspr_addr = ctx.casper.generate_address();

    let evm_addr_str = ctx.bridge.stringify_address(&evm_addr).unwrap();
    let cspr_addr_str = ctx.bridge.stringify_address(&cspr_addr).unwrap();

    ctx.evm
        .bridge_in(
            &evm_addr,
            &cspr_addr_str,
            &token.casper,
            U256::one() * 1_000_000_000 * 1_000_000_000,
        )
        .await;

    tokio::time::sleep(Duration::from_millis(1000)).await;

    let casper_event = if let BridgeEvent::TokenTransferOut(event) =
        ctx.casper.view_events().first().cloned().unwrap()
    {
        event
    } else {
        panic!("invalid variant")
    };

    assert_eq!(casper_event.amount, U256::one() * 1_000_000_000);
    assert_eq!(casper_event.from, evm_addr_str);
    assert_eq!(casper_event.to, cspr_addr);

    Ok(())
}

#[tokio::test]
async fn test_restore_transaction() -> anyhow::Result<()> {
    let mut ctx = TestContext::create().await;
    ctx.time_source.set_auto_advance(true);
    let token = ctx.create_test_token().await;

    ctx.bridge.load_tokens().await.unwrap();

    let evm_addr = ctx.evm.generate_address();
    let cspr_addr = ctx.casper.generate_address();

    let evm_addr_str = ctx.bridge.stringify_address(&evm_addr).unwrap();
    let cspr_addr_str = ctx.bridge.stringify_address(&cspr_addr).unwrap();

    ctx.casper.switch_failing();

    ctx.evm
        .bridge_in(
            &evm_addr,
            &cspr_addr_str,
            &token.casper,
            U256::one() * 1_000_000_000 * 1_000_000_000,
        )
        .await;

    tokio::time::sleep(Duration::from_millis(1000)).await;

    // Transaction should hang because connector is in failure state right now.
    let events = ctx.casper.view_events();
    assert_eq!(events.len(), 0);

    // We are restarting bridge to continue processing.
    ctx.casper.switch_failing(); // Enable casper back.
    ctx.restart_bridge().await;

    tokio::time::sleep(Duration::from_millis(1000)).await;

    let casper_event = if let BridgeEvent::TokenTransferOut(event) =
        ctx.casper.view_events().first().cloned().unwrap()
    {
        event
    } else {
        panic!("invalid variant")
    };

    assert_eq!(casper_event.amount, U256::one() * 1_000_000_000);
    assert_eq!(casper_event.from, evm_addr_str);
    assert_eq!(casper_event.to, cspr_addr);

    Ok(())
}

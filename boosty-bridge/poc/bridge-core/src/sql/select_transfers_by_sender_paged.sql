SELECT tt.id, tt.token_id, tt.amount, tt.status, tt.sender_network_id, tt.sender_address, tt.recipient_network_id, tt.recipient_address, 
    txt.network_id as triggering_tx_nid, txt.txhash as triggering_tx_hash, txt.seen_at,
    txo.network_id as outbound_tx_nid, txo.txhash as outbound_tx_hash
    FROM token_transfers as tt 
    LEFT JOIN transactions as txt ON tt.triggering_tx = txt.id
    LEFT JOIN transactions as txo ON tt.outbound_tx = txo.id
    WHERE tt.sender_network_id = $1 AND tt.sender_address = $2
    ORDER BY txt.seen_at DESC
    LIMIT $3 OFFSET $4
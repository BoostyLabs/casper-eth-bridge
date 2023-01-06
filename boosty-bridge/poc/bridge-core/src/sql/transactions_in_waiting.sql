SELECT tt.id, tt.token_id, tt.amount, tt.sender_network_id, tt.sender_address, tt.recipient_network_id, tt.recipient_address, txt.seen_at
    FROM token_transfers as tt 
    LEFT JOIN transactions as txt ON tt.triggering_tx = txt.id
    WHERE tt.status = 'WAITING'
    ORDER BY txt.seen_at ASC
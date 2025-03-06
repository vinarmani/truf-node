/*
    Migration 002: Primitive Actions
    This file contains actions related to the primitive events stream.
*/

CREATE ACTION insert_record(
    $stream_id TEXT,
    $ts INT8,
    $value NUMERIC(36,18)
) PUBLIC {
    -- Get the caller's address as the data provider
    $data_provider TEXT := @caller;

    -- Ensure the wallet is allowed to write
    if is_wallet_allowed_to_write($data_provider, 1, 0, 'created_at DESC') == false {
        ERROR('wallet not allowed to write');
    }

    -- Ensure that the stream/contract is initiated
    if is_initiated() == false {
        ERROR('contract must be initiated');
    }

    $current_block INT := @height;

    -- Insert the new record into the primitive_events table
    INSERT INTO primitive_events (stream_id, data_provider, ts, value, created_at)
    VALUES ($stream_id, $data_provider, $ts, $value, $current_block);
};

CREATE ACTION is_initiated() PUBLIC view returns (result bool) {
    -- Check if the contract/stream has been initiated by verifying metadata "type"
    for $row in SELECT * FROM metadata WHERE metadata_key = 'type' LIMIT 1 {
         return true;
    }
    return false;
};

-- This action wraps metadata selection with pagination parameters.
-- It supports ordering only by created_at ascending or descending.
CREATE ACTION get_metadata(
    $key TEXT,
    $only_latest BOOL,
    $ref TEXT,
    $limit INT,
    $offset INT,
    $order_by TEXT
) PUBLIC view returns table(
    row_id uuid,
    value_i int,
    value_f NUMERIC(36,18),
    value_b bool,
    value_s TEXT,
    value_ref TEXT,
    created_at INT
) {
    -- Set default values if parameters are null
    if $limit IS NULL{
        $limit := 100;
    }
    if $offset IS NULL{
        $offset := 0;
    }
    if $order_by IS NULL{
        $order_by := 'created_at DESC';
    }

    RETURN SELECT row_id,
                  value_i,
                  value_f,
                  value_b,
                  value_s,
                  value_ref,
                  created_at
           FROM metadata
           WHERE metadata_key = $key
             AND disabled_at IS NULL
             AND ($ref IS NULL OR LOWER(value_ref) = LOWER($ref))
           ORDER BY
               CASE WHEN $order_by = 'created_at DESC' THEN created_at END DESC,
               CASE WHEN $order_by = 'created_at ASC' THEN created_at END ASC
           LIMIT $limit OFFSET $offset;
};

CREATE ACTION is_wallet_allowed_to_write(
    $wallet TEXT,
    $limit INT,
    $offset INT,
    $order_by TEXT
) PUBLIC view returns (result bool) {
    -- Set default pagination parameters if not provided
    if $limit IS NULL{
       $limit := 1;
    }
    if $offset IS NULL{
       $offset := 0;
    }
    if $order_by IS NULL{
       $order_by := 'created_at DESC';
    }

    -- Check if the wallet is the stream owner
    for $row in SELECT * FROM metadata
                 WHERE metadata_key = 'stream_owner'
                   AND value_ref = LOWER($wallet)
                 LIMIT $limit OFFSET $offset {
         return true;
    }

    -- Check if the wallet is explicitly allowed to write via metadata permissions
    for $row in get_metadata('allow_write_wallet', false, $wallet, $limit, $offset, $order_by) {
         return true;
    }

    return false;
};
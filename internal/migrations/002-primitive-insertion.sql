/*
    Migration 002: Primitive Actions
    This file contains actions related to the primitive events stream.
*/

CREATE OR REPLACE ACTION insert_record(
    $stream_id TEXT,
    $date_ts INT8,
    $value NUMERIC(36,18)
) PUBLIC {
    -- Get the caller's address as the data provider
    $data_provider TEXT := @caller;

    -- Ensure the wallet is allowed to write
    if is_wallet_allowed_to_write($data_provider) == false {
        ERROR('wallet not allowed to write');
    }

    -- Ensure that the stream/contract is existent
    if is_existent($data_provider, $stream_id) == false {
        ERROR('contract must be initiated');
    }

    $current_block INT := @height;

    -- Insert the new record into the primitive_events table
    INSERT INTO primitive_events (stream_id, data_provider, date_ts, value, created_at)
    VALUES ($stream_id, $data_provider, $date_ts, $value, $current_block);
};

-- is_existent checks if the stream is existent
CREATE OR REPLACE ACTION is_existent($data_provider TEXT, $stream_id TEXT) PUBLIC view returns (result bool) {
    -- Check if the stream is initiated
    for $row in SELECT * FROM metadata WHERE metadata_key = 'type' AND stream_id = $stream_id AND data_provider = $data_provider LIMIT 1 {
         return true;
    }
    return false;
};


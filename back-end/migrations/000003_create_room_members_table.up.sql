CREATE TABLE room_members (
    room_id   UUID        REFERENCES rooms(id) ON DELETE CASCADE,
    user_id   UUID        REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (room_id, user_id)
);

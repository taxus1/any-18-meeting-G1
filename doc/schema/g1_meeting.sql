-- meeting G1 · 会议室 + 预约（模型看到的最终表结构，照此写代码，不新建/不改表）
CREATE TABLE IF NOT EXISTS t_room (
  id          BIGINT       NOT NULL AUTO_INCREMENT,
  room_no     VARCHAR(32)  NOT NULL,
  name        VARCHAR(128) NOT NULL,
  capacity    INT          NULL,
  status      VARCHAR(16)  NOT NULL DEFAULT 'ACTIVE',
  create_by   VARCHAR(64)  NULL,
  create_time DATETIME     NULL,
  update_by   VARCHAR(64)  NULL,
  update_time DATETIME     NULL,
  deleted_at  DATETIME(3)  NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_room_no (room_no),
  KEY idx_room_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS t_booking (
  id          BIGINT       NOT NULL AUTO_INCREMENT,
  booking_no  VARCHAR(32)  NOT NULL,
  room_id     BIGINT       NOT NULL,
  booker      VARCHAR(64)  NOT NULL,
  start_at    DATETIME     NOT NULL,
  end_at      DATETIME     NOT NULL,
  status      VARCHAR(16)  NOT NULL DEFAULT 'BOOKED',
  create_by   VARCHAR(64)  NULL,
  create_time DATETIME     NULL,
  update_by   VARCHAR(64)  NULL,
  update_time DATETIME     NULL,
  deleted_at  DATETIME(3)  NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_booking_no (booking_no),
  KEY idx_booking_room_time (room_id, start_at, end_at),
  KEY idx_booking_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

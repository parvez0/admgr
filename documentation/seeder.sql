CREATE DATABASE IF NOT EXISTS `admanager` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `admanager`.`slots` (
                                                   id         VARCHAR(36)                         NOT NULL COMMENT 'Unique id of the slot row',
                                                   start_date DATE                                NOT NULL COMMENT 'Slot available start date',
                                                   end_date   DATE                                NOT NULL COMMENT 'Slot available end date, should be greater than or equal to start date',
                                                   position   INT                                 NOT NULL COMMENT 'Position of the slot',
                                                   cost       DECIMAL(10, 2)                      NULL     COMMENT 'Cost of the slot',
                                                   status     VARCHAR(20)                         NOT NULL COMMENT 'Status of the availability of the slot ENUM(OPEN, CLOSE, ON_HOLD)',
                                                   created    DATETIME DEFAULT CURRENT_TIMESTAMP NULL     COMMENT 'Row creation timestamp',
                                                   modified   DATETIME DEFAULT NULL               NULL ON UPDATE CURRENT_TIMESTAMP COMMENT 'Row modification timestamp',
                                                   CONSTRAINT slots_pk PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS `slots_start_date_end_date_index`
    ON `admanager`.`slots` (start_date, end_date);

CREATE INDEX IF NOT EXISTS `slots_status_index`
    ON `admanager`.`slots` (status DESC);

DELIMITER //
CREATE TRIGGER IF NOT EXISTS `admanager`.`before_slots_insert`
    BEFORE INSERT ON `admanager`.`slots`
    FOR EACH ROW
BEGIN
    SET NEW.id = UUID();
END //
DELIMITER ;

CREATE TABLE IF NOT EXISTS `admanager`.`transactions`
(
    id       VARCHAR(36)                          NOT NULL COMMENT 'Unique id of the transaction row'
        PRIMARY KEY,
    slot_id  VARCHAR(36)                          NOT NULL COMMENT 'Slot id referencing slots row for foreign key',
    status   VARCHAR(20)                          NOT NULL COMMENT 'Status of the transaction ENUM(FAILED, SUCCESS)',
    created  DATETIME DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    modified DATETIME                             NULL ON UPDATE CURRENT_TIMESTAMP(),
    CONSTRAINT transactions_slots_id_fk
        FOREIGN KEY (slot_id) REFERENCES `admanager`.`slots` (id)
)
    COMMENT 'Stores information about the slots transactions';

CREATE INDEX IF NOT EXISTS transactions_status_index
    ON `admanager`.`transactions` (status);

DELIMITER //
CREATE TRIGGER IF NOT EXISTS `admanager`.`before_transaction_insert`
    BEFORE INSERT ON `admanager`.`transactions`
    FOR EACH ROW
BEGIN
    SET NEW.id = UUID();
END //
DELIMITER ;

/*
 // Dummy insert queries for slots and transactions table

 INSERT INTO `admanager`.`slots` (start_date, end_date, position, cost, status)
        VALUES ('2023-05-01', '2023-05-07', 1, 50.00, 'OPEN');

 INSERT INTO `admanager`.`transactions` (`slot_id`, `status`) VALUES
('a50e2a85-1e9f-4e75-a7d1-645c62f7e243', 'SUCCESS'),
('2bfa09e6-d62b-4a7e-8568-07d7b06f64af', 'FAILED');
 */
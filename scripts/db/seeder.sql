-- MySQL Script generated by MySQL Workbench
-- Thu May  4 23:38:07 2023
-- Model: New Model    Version: 1.0
-- MySQL Workbench Forward Engineering

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION';

-- -----------------------------------------------------
-- Schema admanager
-- -----------------------------------------------------

-- -----------------------------------------------------
-- Table `slots`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `slots` (
  `date` DATE NOT NULL COMMENT 'Slot available start date',
  `position` INT NOT NULL COMMENT 'Position of the slot',
  `cost` DECIMAL(10,2) NOT NULL COMMENT 'Cost of the slots decimal value update 3 numbers',
  `status` VARCHAR(45) NOT NULL COMMENT 'Status of the availability of the slot ENUM(OPEN, BOOKED, ONHOLD)',
  `created` DATETIME NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Row creation timestamp',
  `modified` DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT 'Row modification timestamp',
  `booked_date` DATETIME NULL,
  `booked_by` VARCHAR(36) NULL,
  PRIMARY KEY (`date`, `position`))
ENGINE = InnoDB;

CREATE INDEX `BY_DATE` USING BTREE ON `slots` (`date`) VISIBLE;


-- -----------------------------------------------------
-- Table `transactions`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `transactions` (
  `txnid` VARCHAR(36) NOT NULL COMMENT 'Unique Id identifying the slots',
  `created` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Record created timestap',
  `date` DATE NOT NULL,
  `position` INT NOT NULL,
  PRIMARY KEY (`date`, `position`),
  CONSTRAINT `slot_foreign_key`
    FOREIGN KEY (`date` , `position`)
    REFERENCES `slots` (`date` , `position`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB;

CREATE INDEX `slot_foreign_key_idx` ON `transactions` (`date` ASC, `position` ASC) VISIBLE;

CREATE UNIQUE INDEX `txnid_UNIQUE` ON `transactions` (`txnid` ASC) VISIBLE;


DELIMITER $$
CREATE DEFINER = CURRENT_USER TRIGGER `transactions_id_BEFORE_INSERT` BEFORE INSERT ON `transactions` FOR EACH ROW
BEGIN
	IF NEW.txnid IS NULL THEN
		SET NEW.txnid = uuid();
	END IF;
END$$


DELIMITER ;

SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;

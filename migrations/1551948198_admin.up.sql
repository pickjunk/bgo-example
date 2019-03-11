CREATE TABLE `bgo`.`admin` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(45) NOT NULL,
  `passwd` VARCHAR(60) NOT NULL,
  `ctime` DATETIME NOT NULL,
  `mtime` DATETIME NOT NULL,
  `ltime` DATETIME NULL,
  PRIMARY KEY (`id`));
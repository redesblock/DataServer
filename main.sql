/*
 Navicat Premium Data Transfer

 Source Server         : mopdataserver
 Source Server Type    : MySQL
 Source Server Version : 80100 (8.1.0)
 Source Host           : 103.201.26.35:23306
 Source Schema         : dataserver

 Target Server Type    : MySQL
 Target Server Version : 80100 (8.1.0)
 File Encoding         : 65001

 Date: 10/09/2023 19:27:55
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for bucket_objects
-- ----------------------------
DROP TABLE IF EXISTS `bucket_objects`;
CREATE TABLE `bucket_objects` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `c_id` longtext,
  `size` bigint unsigned DEFAULT NULL,
  `status` bigint DEFAULT NULL,
  `asset_id` longtext,
  `parent_id` bigint unsigned DEFAULT NULL,
  `bucket_id` bigint unsigned DEFAULT NULL,
  `downloaded` bigint unsigned DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `uplink_progress` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_bucket_objects_deleted_at` (`deleted_at`),
  KEY `fk_bucket_objects_user` (`user_id`),
  CONSTRAINT `fk_bucket_objects_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for buckets
-- ----------------------------
DROP TABLE IF EXISTS `buckets`;
CREATE TABLE `buckets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `access` tinyint(1) DEFAULT NULL,
  `network` longtext,
  `area` longtext,
  `user_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_buckets_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for coupons
-- ----------------------------
DROP TABLE IF EXISTS `coupons`;
CREATE TABLE `coupons` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `coupon_type` bigint unsigned DEFAULT NULL,
  `p_type` bigint unsigned DEFAULT NULL,
  `discount` longtext,
  `storage_quantity_min` bigint unsigned DEFAULT NULL,
  `storage_quantity_max` bigint unsigned DEFAULT NULL,
  `traffic_quantity_min` bigint unsigned DEFAULT NULL,
  `traffic_quantity_max` bigint unsigned DEFAULT NULL,
  `start_time` datetime(3) DEFAULT NULL,
  `end_time` datetime(3) DEFAULT NULL,
  `sold` bigint unsigned DEFAULT NULL,
  `reserve` bigint unsigned DEFAULT NULL,
  `max_claim` bigint unsigned DEFAULT NULL,
  `status` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_coupons_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for currencies
-- ----------------------------
DROP TABLE IF EXISTS `currencies`;
CREATE TABLE `currencies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `symbol` varchar(191) DEFAULT NULL,
  `rate` longtext,
  `base` tinyint(1) DEFAULT '0',
  `payment` bigint unsigned DEFAULT NULL,
  `receiptor` longtext,
  PRIMARY KEY (`id`),
  UNIQUE KEY `symbol` (`symbol`),
  KEY `idx_currencies_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for nodes
-- ----------------------------
DROP TABLE IF EXISTS `nodes`;
CREATE TABLE `nodes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `ip` longtext,
  `port` bigint DEFAULT NULL,
  `zone` longtext,
  `voucher_id` varchar(191) DEFAULT NULL,
  `usable` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `voucher_id` (`voucher_id`),
  KEY `idx_nodes_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for orders
-- ----------------------------
DROP TABLE IF EXISTS `orders`;
CREATE TABLE `orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `order_id` longtext,
  `p_type` bigint unsigned DEFAULT NULL,
  `quantity` bigint unsigned DEFAULT NULL,
  `price` longtext,
  `payment_id` longtext,
  `payment` bigint unsigned DEFAULT NULL,
  `payment_account` longtext,
  `receive_account` longtext,
  `payment_amount` longtext,
  `payment_time` datetime(3) DEFAULT NULL,
  `status` bigint unsigned DEFAULT NULL,
  `hash` longtext,
  `discount` longtext,
  `user_id` bigint unsigned DEFAULT NULL,
  `currency_id` bigint unsigned DEFAULT NULL,
  `payment_time_str` longtext,
  `user_coupon_id` bigint unsigned DEFAULT NULL,
  `coupon_id` bigint unsigned DEFAULT NULL, 
  `discount1` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_orders_deleted_at` (`deleted_at`),
  KEY `fk_orders_user` (`user_id`),
  KEY `fk_orders_currency` (`currency_id`),
  CONSTRAINT `fk_orders_currency` FOREIGN KEY (`currency_id`) REFERENCES `currencies` (`id`),
  CONSTRAINT `fk_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for products
-- ----------------------------
DROP TABLE IF EXISTS `products`;
CREATE TABLE `products` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `p_type` bigint unsigned DEFAULT NULL,
  `quantity` bigint unsigned DEFAULT NULL,
  `price` longtext,
  `currency_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `p_type` (`p_type`),
  KEY `idx_products_deleted_at` (`deleted_at`),
  KEY `fk_products_currency` (`currency_id`),
  CONSTRAINT `fk_products_currency` FOREIGN KEY (`currency_id`) REFERENCES `currencies` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for report_traffics
-- ----------------------------
DROP TABLE IF EXISTS `report_traffics`;
CREATE TABLE `report_traffics` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `token` longtext,
  `uploaded` bigint DEFAULT NULL,
  `uploaded_cnt` bigint DEFAULT NULL,
  `downloaded` bigint DEFAULT NULL,
  `downloaded_cnt` bigint DEFAULT NULL,
  `timestamp` bigint DEFAULT NULL,
  `nat_addr` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_report_traffics_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for sign_ins
-- ----------------------------
DROP TABLE IF EXISTS `sign_ins`;
CREATE TABLE `sign_ins` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `p_type` bigint unsigned DEFAULT NULL,
  `quantity` bigint unsigned DEFAULT NULL,
  `period` bigint unsigned DEFAULT NULL,
  `enable` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `p_type` (`p_type`),
  KEY `idx_sign_ins_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for special_products
-- ----------------------------
DROP TABLE IF EXISTS `special_products`;
CREATE TABLE `special_products` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `p_type` bigint unsigned DEFAULT NULL,
  `quantity` bigint unsigned DEFAULT NULL,
  `discount` longtext,
  `sold` bigint unsigned DEFAULT NULL,
  `reserve` bigint unsigned DEFAULT NULL,
  `product_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_special_products_deleted_at` (`deleted_at`),
  KEY `fk_special_products_product` (`product_id`),
  CONSTRAINT `fk_special_products_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for used_storages
-- ----------------------------
DROP TABLE IF EXISTS `used_storages`;
CREATE TABLE `used_storages` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `num` bigint unsigned DEFAULT NULL,
  `time` longtext,
  `user_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for used_traffics
-- ----------------------------
DROP TABLE IF EXISTS `used_traffics`;
CREATE TABLE `used_traffics` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `num` bigint unsigned DEFAULT NULL,
  `time` longtext,
  `user_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for user_actions
-- ----------------------------
DROP TABLE IF EXISTS `user_actions`;
CREATE TABLE `user_actions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `action_type` bigint unsigned DEFAULT NULL,
  `email` varchar(191) DEFAULT NULL,
  `ip` longtext,
  `user_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_actions_deleted_at` (`deleted_at`),
  KEY `idx_user_actions_email` (`email`),
  KEY `fk_user_actions_user` (`user_id`),
  CONSTRAINT `fk_user_actions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for user_coupons
-- ----------------------------
DROP TABLE IF EXISTS `user_coupons`;
CREATE TABLE `user_coupons` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `status` bigint unsigned DEFAULT NULL,
  `end_time` datetime(3) DEFAULT NULL,
  `p_type` bigint unsigned DEFAULT NULL,
  `coupon_id` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_coupons_deleted_at` (`deleted_at`),
  KEY `fk_user_coupons_user` (`user_id`),
  KEY `fk_user_coupons_coupon` (`coupon_id`),
  CONSTRAINT `fk_user_coupons_coupon` FOREIGN KEY (`coupon_id`) REFERENCES `coupons` (`id`),
  CONSTRAINT `fk_user_coupons_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `email` varchar(191) DEFAULT NULL,
  `password` longtext,
  `first_name` longtext,
  `last_name` longtext,
  `role` bigint unsigned DEFAULT '0',
  `status` bigint unsigned DEFAULT '0',
  `total_storage` bigint unsigned DEFAULT NULL,
  `total_traffic` bigint unsigned DEFAULT NULL,
  `signed_in` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

SET FOREIGN_KEY_CHECKS = 1;

CREATE TABLE jobresult (
  job_id binary(16) NOT NULL,
  metadata JSON NOT NULL,
  error TEXT NOT NULL,
  PRIMARY KEY (`job_id`),
  CONSTRAINT `fk_job_id` FOREIGN KEY (`job_id`) REFERENCES `job` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

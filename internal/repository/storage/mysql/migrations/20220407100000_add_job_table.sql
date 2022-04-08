SET SESSION sql_mode = 'NO_ZERO_DATE';
CREATE TABLE job (
  id binary(16) NOT NULL,
  name varchar(255) NOT NULL,
  task_name varchar(255) NOT NULL,
  task_params JSON NOT NULL,
  timeout INT NOT NULL,
  description varchar(255) NOT NULL DEFAULT '',
  status VARCHAR (50) NOT NULL DEFAULT '',
  failure_reason TEXT NOT NULL,
  run_at timestamp NULL ,
  scheduled_at timestamp NULL,
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  started_at timestamp NULL,
  completed_at timestamp NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

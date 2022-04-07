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
  run_at timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  scheduled_at timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  created_at timestamp NOT NULL DEFAULT current_timestamp(),
  started_at timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  completed_at timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

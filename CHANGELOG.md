
<a name="v0.8.2"></a>
## [v0.8.2](https://github.com/svaloumas/valet/compare/v0.8.1...v0.8.2)

> 2022-04-27

### Chore

* bump version to v0.8.2
* add go report badge
* update badges

### Doc

* update README secrets

### Feat

* merge consumer into scheduler
* **memorydb:** update check health func

### Test

* add dispatch pipeline test case
* minor refactor tests


<a name="v0.8.1"></a>
## [v0.8.1](https://github.com/svaloumas/valet/compare/v0.8.0...v0.8.1)

> 2022-04-25

### Chore

* bump version to v0.8.1

### Doc

* update CHANGELOG
* update README
* update README

### Feat

* add valet stop method and test run

### Test

* add config load default values test cases, add worksrv send merged jobs test
* **taskrepo:** add get task repository test
* **tasksrv:** add get task func test case


<a name="v0.8.0"></a>
## [v0.8.0](https://github.com/svaloumas/valet/compare/v0.6.0...v0.8.0)

> 2022-04-24

### Chore

* bump version to v0.8.0
* update codecov coverage in CI
* update codecov coverage in CI
* add codecov coverage in CI
* remove coverage from CI
* **Makefile:** remove go tool from make report ([#8](https://github.com/svaloumas/valet/issues/8))
* **Makefile:** update make report
* **ci:** use current path in codecov step
* **ci:** fix codecov step ([#7](https://github.com/svaloumas/valet/issues/7))
* **ci:** remove codecov token
* **protos:** add jobresult proto files

### Doc

* update CHANGELOG
* update README
* add pipelines endpoints in swagger files
* add CHANGELOG
* add go report badge in README

### Feat

* add pipeline gRPC endpoints
* add rest of the pipeline CRUD http endpoints
* process pipeline works and support previous results manipulation
* add create pipeline service and http endpoint
* prohibit deletion for job that belongs to a pipeline
* add pipeline branch for logging in consume func
* remove panic branch from job result wait()
* add pipeline CRUD storage methods
* **domain:** add pipeline resource

### Fix

* **worksrv:** properly close work result channels on early exit cases


<a name="v0.6.0"></a>
## [v0.6.0](https://github.com/svaloumas/valet/compare/v0.5.1...v0.6.0)

> 2022-04-16

### Chore

* update coverage badge

### Doc

* update README

### Feat

* add get tasks endpoint


<a name="v0.5.1"></a>
## [v0.5.1](https://github.com/svaloumas/valet/compare/v0.5.0...v0.5.1)

> 2022-04-16

### Chore

* retract versions


<a name="v0.5.0"></a>
## v0.5.0

> 2022-04-16

### Chore

* update README
* rename module
* bump version
* delete bin
* rename module
* bump version
* update coverage badge
* update coverage badge
* rename postgres dsn secret in compose
* update coverage badge
* remove ldflags
* update coverage badge
* update coverage badge
* update coverage badge
* add test vars in Makefile for report
* remove no zero date leftover from Makefile
* update coverage badge
* remove no zero date from mysql
* add proto files
* rephrase comment
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* add depends_on on valet service
* update coverage badge
* update coverage badge
* remove build target from compose override
* remove build dependency from CI
* remove prod target from Dockerfile
* remove build step from CI
* update coverage badge
* update github CI
* update github CI
* update github CI
* update compose files
* update coverage badge
* add LICENSE
* bump version to 0.8.0
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* remove codecov  from CI
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update coverage badge
* update Makefile and CI yml
* update coverage badge
* update CI codecov step syntax
* change badge order
* update CI name
* add CI badge in README
* update commit message for coverage badge in CI
* Updated coverage badge.
* add go coverage badge in CI
* add go coverage badge in CI
* fix upload to codecov step
* fix upload to codecov step
* add upload to codecov step
* remove codecov token from CI
* add codecov token in CI
* add test coverage in CI
* update CI to run on push to every branch
* update CI
* add github ci script
* update and add comments
* update go mod
* install gomock
* install google uuid pkg
* add Makefile
* .gitignore
* install gin and xerrors
* add go.mod

### Doc

* update README
* fix swagger get jobs response
* add LISENCE badge in README
* update secrets in README
* update README
* update config section in README
* update README
* update README secrets section
* add secrets section in README
* update README config section
* update README
* update README
* update README
* update config section in README
* update config section in README
* use smaller headlines for components in README
* update README arch and config section
* update README overview section
* update config section in README
* update swagger files and README
* update swagger files
* update snippet in README
* update snippet in README
* change indentation in README
* change in code snippet in README
* change in code snippet in README
* fix typo in README
* minor changes in README
* move subtitle on top in README
* minor change in README
* fix typo in README
* update ToC in README
* add ToC in README
* fix indentation in README
* fix typo in README
* add usage and configuration section in README
* replace development with installation section in README
* fix minor syntantic errors in  README
* update README overview section
* update README

### Docs

* update README

### Feat

* place valet in root level
* convert valet to library
* add get jobs endpoint
* remove UUID helper functions and use native mysql UUID support
* add PostgreSQL adapter ([#2](https://github.com/svaloumas/valet/issues/2))
* add Redis adapter ([#1](https://github.com/svaloumas/valet/issues/1))
* enable gRPC reflection
* hardwire gRPC and add server factory
* add gRPC handlers
* service delete checks for resource existence before operate
* parameterize rabbitmq consume and publish params
* add server config section
* return error instead of bool in job queue push
* add RabbitMQ adapter
* add job result mysql CRUD
* add job duration
* add scheduled job status
* add get due jobs in mysql pkg
* add update and delete job in mysql pkg
* add get job in mysql
* add mysql CreateJob
* add custom MapInterface type to implement sql driver and store task_params
* add mysql uuid functions migration script
* add mysql migration scripts
* update mysql config
* add env pkg
* update mock ports
* add mysl pkg and merge job and job result repos under storage
* add jogrus and configurable logging format, text and JSON
* add scheduler
* add run_at field for job
* make timeout unit configurable
* work service send blocks until worker pool queue has space
* add full queue error response in create jobs
* add TaskFunc as a Work field
* add consumer service and remove transmitter
* containerize valet
* add status endpoint
* add yaml config
* add CreateJobItem to workerpool interface
* add task repository
* add job timeout
* make valid tasks configurable parameter and refactor tests
* add new error response structure and job handler tests
* add CORS allow all origins and update swagger files
* add compile time proof for concrete error implementations
* add task per job functionality
* add recovery pattern to job callback
* add configurable task timeout
* add the rest endpoints swagger doc
* add patch job endpoint to router
* add swagger doc
* add futureresult pkg and resultsrv tests
* add port mocks
* wait and create job results when they are available
* execute jobs from job service
* add result domain and repository
* add task types map to make task eligible for configuration
* add compile time proof for concrete implementations
* remove transmitter interface
* enable task callback injection
* return no content in update endpoint and return dto instead of job
* add multi task and task metadata support (dummytask)
* add simple logging and ticker at transmit method
* update job queue and transmitter methods
* update mock ports
* wrap up task, transmitter and worker pool in valet
* add workerpool pkg
* add job queue concrete implementation
* embed job queue in job service
* add job queue, worker pool, task and transmitter ports
* update mock ports
* add job update end to end plus minor fixes
* use DI in job service for time and add service tests
* add jobstatus unit tests
* add job unit tests
* add uuidgen to job service
* convert job status string literals to constants
* add uuidgen pkg
* add valetd http server
* add repositories custom errors
* add job http handler
* add job repository
* add struct tags to job domain
* add comments to ports and services
* add jobsrv service
* add JobRepository and JobService ports
* add Job and JobStatus domains

### Fix

* add format operator in string
* remove postponed status leftover from job status
* correct order in NotFoundError fmt args

### Refactor

* use google.protobuf.Value for result metadata field
* move work to its own pkg and rename serve method
* remove mysql migrate pkg and mysql client dependency
* call config setters in config.Load
* rename map string interface custom type
* change task_params type to map[string]interface{}
* move memorydb under its own pkg
* init loggers inside the service constructors
* move mysql under storage pkg and implement repository factory
* replace env config with logging_format
* rename job metadata to task params
* add work service and remove worker pool
* rename JobItem to Work
* rename leftovers in consumersrv
* move task under root dir and rename taskrepo
* move transmitter under its own package
* pass jobItem in workerpool send instead of job
* rename task type to task name
* rename timeout type to timeout unit
* replace loc with utc method in time pkg
* replace xerros pkg with error type assertion
* resouce validation error, still not working
* change init ordering in main
* change base URL to /api in swagger spec
* change base URL to /api
* move router config in a separate file under main
* move errors under pkg/apperrors
* TaskFunc return interface type
* rename result queue chan to result
* rename result queue chan to result
* move jobitem and jobresult under core domain
* move transmitter under valetd and workerpool in repository
* pass only job's metadata to task run
* rename pkgs to singular
* time and uuigen minor changes, use DI for time
* change return types in job service
* change return types in ports

### Test

* add test for dummytask
* add error cases in config load test
* add config load test
* add test main in mysql test pkg
* split array deepequal
* add get due job, schedule and consumer error cases tests
* add full queue error case for post jobs
* add task repository tests
* add sub test report
* add jobdb and resultdb tests
* add fifoqueue pop and close test
* add fifoqueue push test
* remove used validTasks var from test
* add tests for dto funcs and minor refactoring
* disable debug logs in tests
* add handler get job test cases
* add post jobs test cases
* add job service exec test cases
* rename mock pkg

### Reverts

* refactor: init loggers inside the service constructors
* feat: add configurable task timeout


const { sh, cli, help  } = require('tasksfile')

let IS_CI = process.env.IS_CI || false

// Disable entirely the use of docker
let MM_NO_DOCKER = false

// Build Flags
let BUILD_NUMBER = process.env.BUILD_NUMBER || 'dev'
let BUILD_DATE = new Date().toISOString().slice(0,10)
let BUILD_HASH = sh('git rev-parse HEAD', { silent: true }).replace('\n', '')

let BUILD_HASH_ENTERPRISE = 'none'
let BUILD_ENTERPRISE_READY = 'false'

// these variables are used by QA to override location of InProduct Notices
let NOTICES_JSON_URL = process.env.NOTICES_JSON_URL || 'https://hungknow.com/notices'
let NOTICES_FETCH_SECS = process.env.NOTICES_FETCH_SECS || 3600
let NOTICES_SKIP_CACHE = process.env.NOTICES_SKIP_CACHE || false

// Go Flags
let GOFLAGS = process.env.GOFLAGS || ''

// We need to export GOBIN to allow it to be set for processes spawned from the Makefile
let GOBIN = `${__dirname}/bin`
let GO = 'go'

let LDFLAGS = process.env.LDFLAGS || ''
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/model.BuildNumber=${BUILD_NUMBER}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/model.BuildDate=${BUILD_DATE}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/model.BuildHash=${BUILD_HASH}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/model.BuildHashEnterprise=${BUILD_HASH_ENTERPRISE}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/model.BuildEnterpriseReady=${BUILD_ENTERPRISE_READY}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/app.NOTICES_JSON_URL=${NOTICES_JSON_URL}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/app.NOTICES_JSON_FETCH_FREQUENCY_SECONDS=${NOTICES_FETCH_SECS}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/app.NOTICES_SKIP_CACHE=${NOTICES_SKIP_CACHE}'`

let PLATFORM_FILES = "./cmd/hser/main.go"

// Possible options: mysql, postgres, minio, inbucket, openldap, dejavu,
let ENABLED_DOCKER_SERVICES = 'mysql postgres inbucket'
// ifeq (,$(findstring minio,$(ENABLED_DOCKER_SERVICES)))
//   TEMP_DOCKER_SERVICES:=$(TEMP_DOCKER_SERVICES) minio
// endif
// ifeq ($(BUILD_ENTERPRISE_READY),true)
//   ifeq (,$(findstring openldap,$(ENABLED_DOCKER_SERVICES)))
//     TEMP_DOCKER_SERVICES:=$(TEMP_DOCKER_SERVICES) openldap
//   endif
//   ifeq (,$(findstring elasticsearch,$(ENABLED_DOCKER_SERVICES)))
//     TEMP_DOCKER_SERVICES:=$(TEMP_DOCKER_SERVICES) elasticsearch
//   endif
// endif
// let ENABLED_DOCKER_SERVICES = `${ENABLED_DOCKER_SERVICES} ${TEMP_DOCKER_SERVICES}`

function start_docker() {
  if (IS_CI == false) {
    console.log('CI Build: skipping docker start')
  } if (MM_NO_DOCKER == true) {
    console.log('No Docker Enabled: skipping docker start')
  } else {
    console.log('Starting docker containers')
  }
  //TODO: Open
  sh(`${GO} run ./build/docker-compose-generator/main.go $(ENABLED_DOCKER_SERVICES) | docker-compose -f docker-compose.makefile.yml -f /dev/stdin run --rm start_dependencies`, { nopipe: true })

  // if ($(findstring openldap,$(ENABLED_DOCKER_SERVICES))) {
    // sh(`cat tests/${LDAP_DATA}-data.ldif | docker-compose -f docker-compose.makefile.yml exec -T openldap bash -c 'ldapadd -x -D "cn=admin,dc=mm,dc=test,dc=com" -w mostest || true'`, { nopipe: true })
  // }
}
help(start_docker, 'Start necessary services in docker')

// Add test data to the local instance
function test_data() {
  start_docker()

	sh(`${GO} run ${GOFLAGS} -ldflags "${LDFLAGS}" ${PLATFORM_FILES} config set TeamSettings.MaxUsersPerTeam 100`, { nopipe: true })
	sh(`${GO} run ${GOFLAGS} -ldflags "${LDFLAGS}" ${PLATFORM_FILES} sampledata -w 4 -u 60`, { nopipe: true })

	console.log('You may need to restart the HungKnow server before using the following')
	console.log('========================================================================')
	console.log('Login with a system admin account username=sysadmin password=Sys@dmin-sample1')
	console.log('Login with a regular account username=user-1 password=SampleUs@r-1')
  console.log('========================================================================')
}
help(test_data, 'Add test data to the local instance')

function start_server() {
  sh(`go run ./cmd/hser/main.go`, { nopipe: true })
}
help(start_server, 'Start server instance')

cli({
  start_docker,
  start_server,
  test_data,
})

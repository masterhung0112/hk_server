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
let GOBIN = process.env.GOBIN || `${process.env.GOPATH}/bin`
let GO =  process.env.GO || 'go'

let LDFLAGS = process.env.LDFLAGS || ''
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/v5/model.BuildNumber=${BUILD_NUMBER}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/v5/model.BuildDate=${BUILD_DATE}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/v5/model.BuildHash=${BUILD_HASH}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/v5/model.BuildHashEnterprise=${BUILD_HASH_ENTERPRISE}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/v5/model.BuildEnterpriseReady=${BUILD_ENTERPRISE_READY}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/v5/app.NOTICES_JSON_URL=${NOTICES_JSON_URL}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/v5/app.NOTICES_JSON_FETCH_FREQUENCY_SECONDS=${NOTICES_FETCH_SECS}'`
LDFLAGS += ` -X 'github.com/masterhung0112/hk_server/v5/app.NOTICES_SKIP_CACHE=${NOTICES_SKIP_CACHE}'`

let PLATFORM_FILES = "./cmd/hkserver/main.go"

// Possible options: mysql, postgres, minio, inbucket, openldap, dejavu,
let ENABLED_DOCKER_SERVICES = 'mysql postgres inbucket minio'
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

  sh(`${GO} run ./build/docker-compose-generator/main.go ${ENABLED_DOCKER_SERVICES} > enabled_services.yml`, { nopipe: true })
  sh(`docker-compose -f docker-compose.makefile.yml -f enabled_services.yml run --rm start_dependencies`, { nopipe: true })

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
  sh(`${GO} version`, { nopipe: true })
  sh(`${GO} run ./cmd/hkserver/main.go`, { nopipe: true })
}
help(start_server, 'Start server instance')

function store_mocks() {
  sh(`${GO} get -modfile=go.tools.mod github.com/vektra/mockery/...`, { nopipe: true })
  sh(`${GOBIN}/mockery -dir store -all -output store/storetest/mocks -note 'Regenerate this file using \`npm run task store_mocks\`.'`, { nopipe: true })
  sh(`rm store/storetest/mocks/StoreTestBaseSuite.go`)
}
help(store_mocks, 'Creates mock files for stores')

function einterfaces_mocks() {
  sh(`${GO} get -modfile=go.tools.mod github.com/vektra/mockery/...`, { nopipe: true })
  sh(`${GOBIN}/mockery -dir einterfaces -all -output einterfaces/mocks -note 'Regenerate this file using \`npm run task einterfaces_mocks\`.'`, { nopipe: true })
}
help(einterfaces_mocks, 'Creates mock files for einterfaces')

function app_layers() {
  sh(`${GO} get -modfile=go.tools.mod github.com/reflog/struct2interface`, { nopipe: true })
  sh(`${GOBIN}/struct2interface -f "app" -o "app/app_iface.go" -p "app" -s "App" -i "AppIface" -t ./app/layer_generators/app_iface.go.tmpl`, { nopipe: true })
  sh(`${GO} run ./app/layer_generators -in ./app/app_iface.go -out ./app/opentracing/opentracing_layer.go -template ./app/layer_generators/opentracing_layer.go.tmpl`, { nopipe: true })
}
help(app_layers, 'Extract interface from App struct')

function store_layers() {
  sh(`${GO} generate ${GOFLAGS} ./store`, { nopipe: true })

}
help(store_layers, 'Generate layers for the store')

function test_folder(_, package_name) {
  sh(`${GO} test -timeout 60m github.com/masterhung0112/hk_server/v5/${package_name} > test_log.txt`, { nopipe: true })
}

cli({
  start_docker,
  start_server,
  test_data,
  test_folder,

  store_mocks,
  einterfaces_mocks,
  app_layers,
  store_layers,
})

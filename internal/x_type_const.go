package x

type CacheDataType struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Parent  string `json:"parent"`
	State   string `json:"state"`
	LastDay string `json:"lastday"`
	User    string `json:"user"`
	Token   string `json:"token"`
	HttpErr int    `json:"httperr"`
}

type WebResponseType struct {
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
	Type     string   `json:"type"`
	Name     string   `json:"name"`
	Parent   string   `json:"parent"`
	State    string   `json:"state"`
	LastDay  string   `json:"lastday"`
	User     string   `json:"user"`
}

type StatsType struct {
	ValidEnvs   []string `json:"validenvs"`
	LockedEnvs  []string `json:"lockedenvs"`
	MaintEnvs   []string `json:"maintenvs"`
	TermdEnvs   []string `json:"termdenvs"`
	LockedHosts []string `json:"lockedhosts"`
}

const (
	C_ADMIN     string = "admin"
	C_ENV_LIST  string = "envlist"
	C_TYPE_ENV  string = "env"
	C_TYPE_HOST string = "host"
	C_PARENT    string = "parent"

	C_SUCCESS string = "✅"
	C_FAILED  string = "❌"
	C_STARTED string = "🌐"

	C_STATE_VALID       string = "valid"
	C_STATE_LOCKED      string = "locked"
	C_STATE_TERMINATED  string = "termnd"
	C_STATE_MAINTENANCE string = "maint"

	C_RespHeader string = "application/json"
	C_Secret     string = "XXXXXXX"

	ERR_JsonConvertData      string = "ERR: Error converting LockData to JSON."
	ERR_NoNameSpecified      string = "ERR: No 'name' parameter specified."
	ERR_NoTypeSpecified      string = "ERR: No 'type' parameter specified."
	ERR_NoUserSpecified      string = "ERR: No 'user' parameter specified."
	ERR_NoTokenSpecified     string = "ERR: No 'token' parameter specified."
	ERR_WrongTypeSpecified   string = "ERR: Wrong 'type' specified, must be 'env' or 'host'."
	ERR_IllegalUser          string = "ERR: Illegal user."
	ERR_CannotDeleteUser     string = "ERR: User setup failed."
	ERR_EnvLockFail          string = "ERR: Environment lock unsuccesful."
	ERR_EnvCreationFail      string = "ERR: Creating a new enviromnent failed."
	ERR_EnvUnlockFail        string = "ERR: Environment unlock failed."
	ERR_EnvSetMaintFailFail  string = "ERR: Environment maintenance set failed."
	ERR_EnvSetTermFail       string = "ERR: Environment termination failed."
	ERR_HostLockFail         string = "ERR: Host lock unsuccesful."
	ERR_ParentEnvLockFail    string = "ERR: Parent environment is locked, cannot lock host."
	ERR_HostUnlockFail       string = "ERR: Host unlock failed."
	ERR_InvalidDateSpecified string = "ERR: Invalid 'lastday' specified, format is: YYYYMMDD."
	ERR_NoAdminPresent       string = "ERR: No 'admin' user present, cannot continue."
	ERR_LockedHostsInEnv     string = "ERR: Locked hosts in envm it cannot be locked."
	ERR_UserExists           string = "ERR: User already exists."
	ERR_UserSetupFailed      string = "ERR: Cannot setup user."

	OK_UserPurged          string = "OK: User purged."
	OK_EnvCreated          string = "OK: Environment created."
	OK_EnvUnlocked         string = "OK: Environment unlocked."
	OK_EnvLocked           string = "OK: Environment locked successfully."
	OK_EnvSetToMaintenance string = "OK: Environment is in maintenance mode now."
	OK_EnvSetToTerminate   string = "OK: Environment terminated."
	OK_HostUnlocked        string = "OK: Host has been unlocked succesfully."
	OK_HostLocked          string = "OK: Host has been locked succesfully."

	C_HTTP_OK          = 0    // default no-error state
	C_TLS_ENABLED bool = true // serve TLS with self-signed cert?

)

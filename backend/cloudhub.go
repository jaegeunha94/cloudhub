package cloudhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

// General errors.
const (
	ErrUpstreamTimeout                 = Error("request to backend timed out")
	ErrSourceNotFound                  = Error("source not found")
	ErrServerNotFound                  = Error("server not found")
	ErrLayoutNotFound                  = Error("layout not found")
	ErrProtoboardNotFound              = Error("protoboard not found")
	ErrDashboardNotFound               = Error("dashboard not found")
	ErrUserNotFound                    = Error("user not found")
	ErrLayoutInvalid                   = Error("layout is invalid")
	ErrProtoboardInvalid               = Error("protoboard is invalid")
	ErrDashboardInvalid                = Error("dashboard is invalid")
	ErrSourceInvalid                   = Error("source is invalid")
	ErrServerInvalid                   = Error("server is invalid")
	ErrAlertNotFound                   = Error("alert not found")
	ErrAuthentication                  = Error("user not authenticated")
	ErrUninitialized                   = Error("client uninitialized. Call Open() method")
	ErrInvalidAxis                     = Error("Unexpected axis in cell. Valid axes are 'x', 'y', and 'y2'")
	ErrInvalidColorType                = Error("Invalid color type. Valid color types are 'min', 'max', 'threshold', 'text', and 'background'")
	ErrInvalidColor                    = Error("Invalid color. Accepted color format is #RRGGBB")
	ErrInvalidLegend                   = Error("Invalid legend. Orientation must be set")
	ErrInvalidLegendType               = Error("Invalid legend type. Valid legend type must be 'static'")
	ErrInvalidLegendOrient             = Error("Invalid orientation type. Valid orientation types are 'top', 'bottom', 'right', 'left'")
	ErrUserAlreadyExists               = Error("user already exists")
	ErrOrganizationNotFound            = Error("organization not found")
	ErrMappingNotFound                 = Error("mapping not found")
	ErrOrganizationAlreadyExists       = Error("organization already exists")
	ErrCannotDeleteDefaultOrganization = Error("cannot delete default organization")
	ErrConfigNotFound                  = Error("cannot find configuration")
	ErrAnnotationNotFound              = Error("annotation not found")
	ErrInvalidCellOptionsText          = Error("invalid text wrapping option. Valid wrappings are 'truncate', 'wrap', and 'single line'")
	ErrInvalidCellOptionsSort          = Error("cell options sortby cannot be empty'")
	ErrInvalidCellOptionsColumns       = Error("cell options columns cannot be empty'")
	ErrOrganizationConfigNotFound      = Error("could not find organization config")
	ErrInvalidCellQueryType            = Error("invalid cell query type: must be 'flux' or 'influxql'")
	ErrVsphereNotFound                 = Error("vsphere not found")
	ErrTopologyNotFound                = Error("topology not found")
	ErrTopologyAlreadyExists           = Error("topology already exists")
	ErrCSPNotFound                     = Error("CSP not found")
	ErrCSPAlreadyExists                = Error("CSP already exists")
	ErrDeviceNotFound                  = Error("Network Device not found")
	ErrDeviceOrgNotFound               = Error("Network Device organization not found")
	ErrTemplatesInvalid                = Error("template is invalid")
	ErrTemplateInvalid                 = Error("invalid template")
	ErrTemplateNotFound                = Error("template not found")
	ErrMLNxRstNotFound                 = Error("MLNxRet not found")
	ErrDLNxRstNotFound                 = Error("DLNxRet not found")
)

// Error is a domain error encountered while processing CloudHub requests
type Error string

func (e Error) Error() string {
	return string(e)
}

// Logger represents an abstracted structured logging implementation. It
// provides methods to trigger log messages at various alert levels and a
// WithField method to set keys for a structured log message.
type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Error(...interface{})

	WithField(string, interface{}) Logger

	// Logger can be transformed into an io.Writer.
	// That writer is the end of an io.Pipe and it is your responsibility to close it.
	Writer() *io.PipeWriter
}

// Router is an abstracted Router based on the API provided by the
// julienschmidt/httprouter package.
type Router interface {
	http.Handler
	GET(string, http.HandlerFunc)
	PATCH(string, http.HandlerFunc)
	POST(string, http.HandlerFunc)
	DELETE(string, http.HandlerFunc)
	PUT(string, http.HandlerFunc)

	Handler(string, string, http.Handler)
}

// Assets returns a handler to serve the website.
type Assets interface {
	Handler() http.Handler
}

// Supported time-series databases
const (
	// InfluxDB is the open-source time-series database
	InfluxDB = "influx"
	// InfluxEnteprise is the clustered HA time-series database
	InfluxEnterprise = "influx-enterprise"
	// InfluxRelay is the basic HA layer over InfluxDB
	InfluxRelay = "influx-relay"
	// InfluxDBv2 is Influx DB 2.x with Token authentication
	InfluxDBv2 = "influx-v2"
)

// TSDBStatus represents the current status of a time series database
type TSDBStatus interface {
	// Connect will connect to the time series using the information in `Source`.
	Connect(ctx context.Context, src *Source) error
	// Ping returns version and TSDB type of time series database if reachable.
	Ping(context.Context) error
	// Version returns the version of the TSDB database
	Version(context.Context) (string, error)
	// Type returns the type of the TSDB database
	Type(context.Context) (string, error)
}

// Point is a field set in a series
type Point struct {
	Database        string
	RetentionPolicy string
	Measurement     string
	Time            int64
	Tags            map[string]string
	Fields          map[string]interface{}
}

// TimeSeries represents a queryable time series database.
type TimeSeries interface {
	// Connect will connect to the time series using the information in `Source`.
	Connect(context.Context, *Source) error
	// Query retrieves time series data from the database.
	Query(context.Context, Query) (Response, error)
	// Write records points into a series
	Write(context.Context, []Point) error
	// UsersStore represents the user accounts within the TimeSeries database
	Users(context.Context) UsersStore
	// Permissions returns all valid names permissions in this database
	Permissions(context.Context) Permissions
	// Roles represents the roles associated with this TimesSeriesDatabase
	Roles(context.Context) (RolesStore, error)
}

// Role is a restricted set of permissions assigned to a set of users.
type Role struct {
	Name         string      `json:"name"`
	Permissions  Permissions `json:"permissions,omitempty"`
	Users        []User      `json:"users,omitempty"`
	Organization string      `json:"organization,omitempty"`
}

// RolesStore is the Storage and retrieval of authentication information
type RolesStore interface {
	// All lists all roles from the RolesStore
	All(context.Context) ([]Role, error)
	// Create a new Role in the RolesStore
	Add(context.Context, *Role) (*Role, error)
	// Delete the Role from the RolesStore
	Delete(context.Context, *Role) error
	// Get retrieves a role if name exists.
	Get(ctx context.Context, name string) (*Role, error)
	// Update the roles' users or permissions
	Update(context.Context, *Role) error
}

// Range represents an upper and lower bound for data
type Range struct {
	Upper int64 `json:"upper"` // Upper is the upper bound
	Lower int64 `json:"lower"` // Lower is the lower bound
}

// TemplateValue is a value use to replace a template in an InfluxQL query
type TemplateValue struct {
	Value    string `json:"value"`         // Value is the specific value used to replace a template in an InfluxQL query
	Type     string `json:"type"`          // Type can be tagKey, tagValue, fieldKey, csv, map, measurement, database, constant, influxql
	Selected bool   `json:"selected"`      // Selected states that this variable has been picked to use for replacement
	Key      string `json:"key,omitempty"` // Key is the key for the Value if the Template Type is 'map'
}

// TemplateVar is a named variable within an InfluxQL query to be replaced with Values
type TemplateVar struct {
	Var    string          `json:"tempVar"` // Var is the string to replace within InfluxQL
	Values []TemplateValue `json:"values"`  // Values are the replacement values within InfluxQL
}

// TemplateID is the unique ID used to identify a template
type TemplateID string

// Template represents a series of choices to replace TemplateVars within InfluxQL
type Template struct {
	TemplateVar
	ID    TemplateID     `json:"id"`              // ID is the unique ID associated with this template
	Type  string         `json:"type"`            // Type can be fieldKeys, tagKeys, tagValues, csv, constant, measurements, databases, map, influxql, text
	Label string         `json:"label"`           // Label is a user-facing description of the Template
	Query *TemplateQuery `json:"query,omitempty"` // Query is used to generate the choices for a template
}

// Query retrieves a Response from a TimeSeries.
type Query struct {
	Command  string   `json:"query"`              // Command is the query itself
	DB       string   `json:"db,omitempty"`       // DB is optional and if empty will not be used.
	RP       string   `json:"rp,omitempty"`       // RP is a retention policy and optional; if empty will not be used.
	Epoch    string   `json:"epoch,omitempty"`    // Epoch is the time format for the return results
	Wheres   []string `json:"wheres,omitempty"`   // Wheres restricts the query to certain attributes
	GroupBys []string `json:"groupbys,omitempty"` // GroupBys collate the query by these tags
	Label    string   `json:"label,omitempty"`    // Label is the Y-Axis label for the data
	Range    *Range   `json:"range,omitempty"`    // Range is the default Y-Axis range for the data
	UUID     string   `json:"uuid,omitempty"`     // Indentifier from client to be added to the result
}

// DashboardQuery includes state for the query builder.  This is a transition
// struct while we move to the full InfluxQL AST
type DashboardQuery struct {
	Command     string      `json:"query"`                 // Command is the query itself
	Label       string      `json:"label,omitempty"`       // Label is the Y-Axis label for the data
	Range       *Range      `json:"range,omitempty"`       // Range is the default Y-Axis range for the data
	QueryConfig QueryConfig `json:"queryConfig,omitempty"` // QueryConfig represents the query state that is understood by the data explorer
	Source      string      `json:"source"`                // Source is the optional URI to the data source for this queryConfig
	Shifts      []TimeShift `json:"-"`                     // Shifts represents shifts to apply to an influxql query's time range.  Clients expect the shift to be in the generated QueryConfig
	Type        string      `json:"type"`                  // Type represents the language the query is in (flux or influxql)
}

// TemplateQuery is used to retrieve choices for template replacement
type TemplateQuery struct {
	Command     string `json:"influxql"`       // Command is the query itself
	Flux        string `json:"flux,omitempty"` // flux is the flux query, if available
	DB          string `json:"db,omitempty"`   // DB is optional and if empty will not be used.
	RP          string `json:"rp,omitempty"`   // RP is a retention policy and optional; if empty will not be used.
	Measurement string `json:"measurement"`    // Measurement is the optionally selected measurement for the query
	TagKey      string `json:"tagKey"`         // TagKey is the optionally selected tag key for the query
	FieldKey    string `json:"fieldKey"`       // FieldKey is the optionally selected field key for the query
}

// Response is the result of a query against a TimeSeries
type Response interface {
	MarshalJSON() ([]byte, error)
}

// Source is connection information to a time-series data store.
type Source struct {
	ID                 int    `json:"id,string"`                    // ID is the unique ID of the source
	Name               string `json:"name"`                         // Name is the user-defined name for the source
	Type               string `json:"type,omitempty"`               // Type specifies which kinds of source (enterprise vs oss)
	Username           string `json:"username,omitempty"`           // Username is the username to connect to the source
	Password           string `json:"password,omitempty"`           // Password is in CLEARTEXT
	SharedSecret       string `json:"sharedSecret,omitempty"`       // ShareSecret is the optional signing secret for Influx JWT authorization
	URL                string `json:"url"`                          // URL are the connections to the source
	MetaURL            string `json:"metaUrl,omitempty"`            // MetaURL is the url for the meta node
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty"` // InsecureSkipVerify as true means any certificate presented by the source is accepted.
	Default            bool   `json:"default"`                      // Default specifies the default source for the application
	Telegraf           string `json:"telegraf"`                     // Telegraf is the db telegraf is written to.  By default it is "telegraf"
	Organization       string `json:"organization"`                 // Organization is the organization ID that resource belongs to
	Role               string `json:"role,omitempty"`               // Not Currently Used. Role is the name of the minimum role that a user must possess to access the resource.
	DefaultRP          string `json:"defaultRP"`                    // DefaultRP is the default retention policy used in database queries to this source
	Version            string `json:"version,omitempty"`            // Version of influxdb
}

// SourcesStore stores connection information for a `TimeSeries`
type SourcesStore interface {
	// All returns all sources in the store
	All(context.Context) ([]Source, error)
	// Add creates a new source in the SourcesStore and returns Source with ID
	Add(context.Context, Source) (Source, error)
	// Delete the Source from the store
	Delete(context.Context, Source) error
	// Get retrieves Source if `ID` exists
	Get(ctx context.Context, ID int) (Source, error)
	// Update the Source in the store.
	Update(context.Context, Source) error
}

// DBRP represents a database and retention policy for a time series source
type DBRP struct {
	DB string `json:"db"`
	RP string `json:"rp"`
}

// AlertRule represents rules for building a tickscript alerting task
type AlertRule struct {
	ID            string        `json:"id,omitempty"`           // ID is the unique ID of the alert
	TICKScript    TICKScript    `json:"tickscript"`             // TICKScript is the raw tickscript associated with this Alert
	Query         *QueryConfig  `json:"query"`                  // Query is the filter of data for the alert.
	Every         string        `json:"every"`                  // Every how often to check for the alerting criteria
	AlertNodes    AlertNodes    `json:"alertNodes"`             // AlertNodes defines the destinations for the alert
	Message       string        `json:"message"`                // Message included with alert
	Details       string        `json:"details"`                // Details is generally used for the Email alert.  If empty will not be added.
	Trigger       string        `json:"trigger"`                // Trigger is a type that defines when to trigger the alert
	TriggerValues TriggerValues `json:"values"`                 // Defines the values that cause the alert to trigger
	Name          string        `json:"name"`                   // Name is the user-defined name for the alert
	Type          string        `json:"type"`                   // Represents the task type where stream is data streamed to kapacitor and batch is queried by kapacitor
	DBRPs         []DBRP        `json:"dbrps"`                  // List of database retention policy pairs the task is allowed to access
	Status        string        `json:"status"`                 // Represents if this rule is enabled or disabled in kapacitor
	Executing     bool          `json:"executing"`              // Whether the task is currently executing
	Error         string        `json:"error"`                  // Any error encountered when kapacitor executes the task
	Created       time.Time     `json:"created"`                // Date the task was first created
	Modified      time.Time     `json:"modified"`               // Date the task was last modified
	LastEnabled   time.Time     `json:"last-enabled,omitempty"` // Date the task was last set to status enabled
}

// TemplateFieldType represents the type of template field
type TemplateFieldType string

const (
	// LearnScriptPrefix TickScript ID Prefix
	LearnScriptPrefix = "learn-"
	// PredictScriptPrefix TickScript ID Prefix
	PredictScriptPrefix = "predict-"
)

// AutoGeneratePredictionRule extends Prediction AlertRule with an additional TaskTemplate field for automatic rule registration.
type AutoGeneratePredictionRule struct {
	AlertRule
	TaskTemplate         TemplateFieldType `json:"task_template,omitempty"` // TaskTemplate is the template string for the task.
	Organization         string            `json:"organization"`
	OrganizationName     string            `json:"organization_name"`
	PredictMode          string            `json:"predict_mode"`
	PredictModeCondition string            `json:"predict_mode_condition"`
}

// AutoGenerateLearnRule extends Learning Rule with an additional TaskTemplate field for automatic rule registration.
type AutoGenerateLearnRule struct {
	AlertRule
	TaskTemplate     TemplateFieldType `json:"task_template,omitempty"` // TaskTemplate is the template string for the task.
	Organization     string            `json:"organization"`
	OrganizationName string            `json:"organization_name"`
	LearningCron     string            `json:"learning_cron"`
	LoadModule       string            `json:"load_module,omitempty"`
	MLFunction       string            `json:"ml_function"`
	RetentionPolicy  string            `json:"retention_policy"`
	InfluxOrigin     string            `json:"influxdb_origin"`
	InfluxDBPort     string            `json:"influxdb_port"`
	InfluxDBUsername string            `json:"influxdb_username"`
	InfluxDBPassword string            `json:"influxdb_password"`
	EtcdOrigin       string            `json:"etcd_origin"`
	EtcdPort         string            `json:"etcd_port"`
	ProcCnt          int               `json:"process_count"`
}

// TICKScript task to be used by kapacitor
type TICKScript string

// TemplateParamsMap is TemplateParams Params
type TemplateParamsMap map[string]interface{}

// TemplateBlock is TickScript Template Params
type TemplateBlock struct {
	Name   string
	Params TemplateParamsMap
}

// LoadTemplateConfig Load file info
type LoadTemplateConfig struct {
	Field          TemplateFieldType
	TemplateString string
}

// Ticker generates tickscript tasks for kapacitor
type Ticker interface {
	// Generate will create the tickscript to be used as a kapacitor task
	Generate(AlertRule) (TICKScript, error)
}

// TriggerValues specifies the alerting logic for a specific trigger type
type TriggerValues struct {
	Change     string `json:"change,omitempty"`   // Change specifies if the change is a percent or absolute
	Period     string `json:"period,omitempty"`   // Period length of time before deadman is alerted
	Shift      string `json:"shift,omitempty"`    // Shift is the amount of time to look into the past for the alert to compare to the present
	Operator   string `json:"operator,omitempty"` // Operator for alert comparison
	Value      string `json:"value,omitempty"`    // Value is the boundary value when alert goes critical
	RangeValue string `json:"rangeValue"`         // RangeValue is an optional value for range comparisons
}

// Field represent influxql fields and functions from the UI
type Field struct {
	Value   interface{} `json:"value"`
	Type    string      `json:"type"`
	Alias   string      `json:"alias"`
	Args    []Field     `json:"args,omitempty"`
	SubFunc string      `json:"subFunc,omitempty"`
}

// GroupBy represents influxql group by tags from the UI
type GroupBy struct {
	Time string   `json:"time"`
	Tags []string `json:"tags"`
}

// DurationRange represents the lower and upper durations of the query config
type DurationRange struct {
	Upper string `json:"upper"`
	Lower string `json:"lower"`
}

// TimeShift represents a shift to apply to an influxql query's time range
type TimeShift struct {
	Label    string `json:"label"`    // Label user facing description
	Unit     string `json:"unit"`     // Unit influxql time unit representation i.e. ms, s, m, h, d
	Quantity string `json:"quantity"` // Quantity number of units
}

// QueryConfig represents UI query from the data explorer
type QueryConfig struct {
	ID              string              `json:"id,omitempty"`
	Database        string              `json:"database"`
	Measurement     string              `json:"measurement"`
	RetentionPolicy string              `json:"retentionPolicy"`
	Fields          []Field             `json:"fields"`
	Tags            map[string][]string `json:"tags"`
	GroupBy         GroupBy             `json:"groupBy"`
	AreTagsAccepted bool                `json:"areTagsAccepted"`
	Fill            string              `json:"fill,omitempty"`
	RawText         *string             `json:"rawText"`
	Range           *DurationRange      `json:"range"`
	Shifts          []TimeShift         `json:"shifts"`
}

// KapacitorNode adds arguments and properties to an alert
type KapacitorNode struct {
	Name       string              `json:"name"`
	Args       []string            `json:"args"`
	Properties []KapacitorProperty `json:"properties"`
	// In the future we could add chaining methods here.
}

// KapacitorProperty modifies the node they are called on
type KapacitorProperty struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

// Server represents a proxy connection to an HTTP server
type Server struct {
	ID                 int                    `json:"id,string"`          // ID is the unique ID of the server
	SrcID              int                    `json:"srcId,string"`       // SrcID of the data source
	Name               string                 `json:"name"`               // Name is the user-defined name for the server
	Username           string                 `json:"username"`           // Username is the username to connect to the server
	Password           string                 `json:"password"`           // Password is in CLEARTEXT
	URL                string                 `json:"url"`                // URL are the connections to the server
	InsecureSkipVerify bool                   `json:"insecureSkipVerify"` // InsecureSkipVerify as true means any certificate presented by the server is accepted.
	Active             bool                   `json:"active"`             // Is this the active server for the source?
	Organization       string                 `json:"organization"`       // Organization is the organization ID that resource belongs to
	Type               string                 `json:"type"`               // Type is the kind of service (e.g. kapacitor or flux)
	Metadata           map[string]interface{} `json:"metadata"`           // Metadata is any other data that the frontend wants to store about this service
}

// ServersStore stores connection information for a `Server`
type ServersStore interface {
	// All returns all servers in the store
	All(context.Context) ([]Server, error)
	// Add creates a new source in the ServersStore and returns Server with ID
	Add(context.Context, Server) (Server, error)
	// Delete the Server from the store
	Delete(context.Context, Server) error
	// Get retrieves Server if `ID` exists
	Get(ctx context.Context, ID int) (Server, error)
	// Update the Server in the store.
	Update(context.Context, Server) error
}

// ID creates uniq ID string
type ID interface {
	// Generate creates a unique ID string
	Generate() (string, error)
}

const (
	// AllScope grants permission for all databases.
	AllScope Scope = "all"
	// DBScope grants permissions for a specific database
	DBScope Scope = "database"
)

// Permission is a specific allowance for User or Role bound to a
// scope of the data source
type Permission struct {
	Scope   Scope      `json:"scope"`
	Name    string     `json:"name,omitempty"`
	Allowed Allowances `json:"allowed"`
}

// Permissions represent the entire set of permissions a User or Role may have
type Permissions []Permission

// Allowances defines what actions a user can have on a scoped permission
type Allowances []string

// Scope defines the location of access of a permission
type Scope string

// User represents an authenticated user.
type User struct {
	ID                 uint64      `json:"id,string,omitempty"`
	Name               string      `json:"name"`
	Passwd             string      `json:"password,omitempty"`
	Permissions        Permissions `json:"permissions,omitempty"`
	Roles              []Role      `json:"roles"`
	Provider           string      `json:"provider,omitempty"`
	Scheme             string      `json:"scheme,omitempty"`
	SuperAdmin         bool        `json:"superAdmin,omitempty"`
	PasswordUpdateDate string      `json:"passwordUpdateDate,omitempty"`
	PasswordResetFlag  string      `json:"passwordResetFlag,omitempty"`
	Email              string      `json:"email,omitempty"`
	RetryCount         int32       `json:"retryCount,omitempty"`
	LockedTime         string      `json:"lockedTime,omitempty"`
	Locked             bool        `json:"locked,omitempty"`
}

// UserQuery represents the attributes that a user may be retrieved by.
// It is predominantly used in the UsersStore.Get method.
//
// It is expected that only one of ID or Name, Provider, and Scheme will be
// specified, but all are provided UserStores should prefer ID.
type UserQuery struct {
	ID       *uint64
	Name     *string
	Provider *string
	Scheme   *string
}

// UsersStore is the Storage and retrieval of authentication information
//
// While not necessary for the app to function correctly, it is
// expected that Implementors of the UsersStore will take
// care to guarantee that the combinartion of a  users Name, Provider,
// and Scheme are unique.
type UsersStore interface {
	// All lists all users from the UsersStore
	All(context.Context) ([]User, error)
	// Create a new User in the UsersStore
	Add(context.Context, *User) (*User, error)
	// Delete the User from the UsersStore
	Delete(context.Context, *User) error
	// Get retrieves a user if name exists.
	Get(ctx context.Context, q UserQuery) (*User, error)
	// Update the user's permissions or roles
	Update(context.Context, *User) error
	// Num returns the number of users in the UsersStore
	Num(context.Context) (int, error)
}

// Database represents a database in a time series source
type Database struct {
	Name          string `json:"name"`                    // a unique string identifier for the database
	Duration      string `json:"duration,omitempty"`      // the duration (when creating a default retention policy)
	Replication   int32  `json:"replication,omitempty"`   // the replication factor (when creating a default retention policy)
	ShardDuration string `json:"shardDuration,omitempty"` // the shard duration (when creating a default retention policy)
}

// RetentionPolicy represents a retention policy in a time series source
type RetentionPolicy struct {
	Name          string `json:"name"`                    // a unique string identifier for the retention policy
	Duration      string `json:"duration,omitempty"`      // the duration
	Replication   int32  `json:"replication,omitempty"`   // the replication factor
	ShardDuration string `json:"shardDuration,omitempty"` // the shard duration
	Default       bool   `json:"isDefault,omitempty"`     // whether the RP should be the default
}

// Measurement represents a measurement in a time series source
type Measurement struct {
	Name string `json:"name"` // a unique string identifier for the measurement
}

// Databases represents a databases in a time series source
type Databases interface {
	// AllDB lists all databases in the current data source
	AllDB(context.Context) ([]Database, error)
	// Connect connects to a database in the current data source
	Connect(context.Context, *Source) error
	// CreateDB creates a database in the current data source
	CreateDB(context.Context, *Database) (*Database, error)
	// DropDB drops a database in the current data source
	DropDB(context.Context, string) error

	// AllRP lists all retention policies in the current data source
	AllRP(context.Context, string) ([]RetentionPolicy, error)
	// CreateRP creates a retention policy in the current data source
	CreateRP(context.Context, string, *RetentionPolicy) (*RetentionPolicy, error)
	// UpdateRP updates a retention policy in the current data source
	UpdateRP(context.Context, string, string, *RetentionPolicy) (*RetentionPolicy, error)
	// DropRP drops a retention policy in the current data source
	DropRP(context.Context, string, string) error

	// GetMeasurements lists measurements in the current data source
	GetMeasurements(ctx context.Context, db string, limit, offset int) ([]Measurement, error)
}

// AnnotationTags describes a set of user-defined tags associated with an Annotation
type AnnotationTags map[string]string

var annotationTagsBlacklist = map[string]bool{
	"deleted":          true,
	"start_time":       true,
	"startTime":        true,
	"end_time":         true,
	"endTime":          true,
	"modified_time_ns": true,
	"text":             true,
	"type":             true,
	"id":               true,
}

// ValidateAnnotationTagKey checks whether a user supplied tag can be stored
// in Annotation.Tags
func ValidateAnnotationTagKey(tagKey string) error {
	if _, prs := annotationTagsBlacklist[tagKey]; prs {
		return fmt.Errorf("Cannot use %q as tag key", tagKey)
	}

	return nil
}

// Valid returns an error if any key of the AnnotationTags is invalid
func (t AnnotationTags) Valid() error {
	for k := range t {
		if err := ValidateAnnotationTagKey(k); err != nil {
			return err
		}
	}

	return nil
}

// Annotation represents a time-based metadata associated with a source
type Annotation struct {
	ID        string         // ID is the unique annotation identifier
	StartTime time.Time      // StartTime starts the annotation
	EndTime   time.Time      // EndTime ends the annotation
	Text      string         // Text is the associated user-facing text describing the annotation
	Tags      AnnotationTags // Tags is a collection of user defined key/value pairs that contextualize the annotation
}

// AnnotationTagFilter describes data used to filter a collection of Annotations by their Tags
type AnnotationTagFilter struct {
	Key        string
	Value      string
	Comparator string // Either '=', '==', '!=', '=~', or '!~'
}

func (f *AnnotationTagFilter) String() string {
	return fmt.Sprintf("%s %s %s", f.Key, f.Comparator, f.Value)
}

// AnnotationStore represents storage and retrieval of annotations
type AnnotationStore interface {
	All(ctx context.Context, start, stop time.Time, filters []*AnnotationTagFilter) ([]Annotation, error) // All lists all Annotations between start and stop
	Add(context.Context, *Annotation) (*Annotation, error)                                                // Add creates a new annotation in the store
	Delete(ctx context.Context, id string) error                                                          // Delete removes the annotation from the store
	Get(ctx context.Context, id string) (*Annotation, error)                                              // Get retrieves an annotation
	Update(context.Context, *Annotation) error                                                            // Update replaces annotation
}

// DashboardID is the dashboard ID
type DashboardID int

// Dashboard represents all visual and query data for a dashboard
type Dashboard struct {
	ID           DashboardID     `json:"id"`
	Cells        []DashboardCell `json:"cells"`
	Templates    []Template      `json:"templates"`
	Name         string          `json:"name"`
	Organization string          `json:"organization"` // Organization is the organization ID that resource belongs to
}

// UnmarshalJSON unmarshals a string ID into a DashboardID (int).
func (d *Dashboard) UnmarshalJSON(data []byte) error {
	type Alias Dashboard

	aux := &struct {
		ID interface{} `json:"id,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.ID != nil {
		// Allows backwards compatibility with filestore `.dashboard` files.
		switch id := aux.ID.(type) {
		case int:
			d.ID = DashboardID(id)
		case int32:
			d.ID = DashboardID(id)
		case int64:
			d.ID = DashboardID(id)
		case float32:
			d.ID = DashboardID(id)
		case float64:
			d.ID = DashboardID(id)
		case string:
			ID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return err
			}
			d.ID = DashboardID(ID)
		default:
			return fmt.Errorf("invalid id type %T", id)
		}
	}

	return nil
}

// Axis represents the visible extents of a visualization
type Axis struct {
	Bounds       []string `json:"bounds"` // bounds are an arbitrary list of client-defined strings that specify the viewport for a cell
	LegacyBounds [2]int64 `json:"-"`      // legacy bounds are for testing a migration from an earlier version of axis
	Label        string   `json:"label"`  // label is a description of this Axis
	Prefix       string   `json:"prefix"` // Prefix represents a label prefix for formatting axis values
	Suffix       string   `json:"suffix"` // Suffix represents a label suffix for formatting axis values
	Base         string   `json:"base"`   // Base represents the radix for formatting axis values
	Scale        string   `json:"scale"`  // Scale is the axis formatting scale. Supported: "log", "linear"
}

// CellColor represents the encoding of data into visualizations
type CellColor struct {
	ID    string `json:"id"`    // ID is the unique id of the cell color
	Type  string `json:"type"`  // Type is how the color is used. Accepted (min,max,threshold)
	Hex   string `json:"hex"`   // Hex is the hex number of the color
	Name  string `json:"name"`  // Name is the user-facing name of the hex color
	Value string `json:"value"` // Value is the data value mapped to this color
}

// Legend represents the encoding of data into a legend
type Legend struct {
	Type        string `json:"type,omitempty"`
	Orientation string `json:"orientation,omitempty"`
}

// DashboardCell holds visual and query information for a cell
type DashboardCell struct {
	ID             string           `json:"i"`
	X              int32            `json:"x"`
	Y              int32            `json:"y"`
	W              int32            `json:"w"`
	H              int32            `json:"h"`
	MinW           int32            `json:"minW"`
	MinH           int32            `json:"minH"`
	Name           string           `json:"name"`
	Queries        []DashboardQuery `json:"queries"`
	Axes           map[string]Axis  `json:"axes"`
	Type           string           `json:"type"`
	CellColors     []CellColor      `json:"colors"`
	Legend         Legend           `json:"legend"`
	TableOptions   TableOptions     `json:"tableOptions,omitempty"`
	FieldOptions   []RenamableField `json:"fieldOptions"`
	TimeFormat     string           `json:"timeFormat"`
	DecimalPlaces  DecimalPlaces    `json:"decimalPlaces"`
	Note           string           `json:"note"`
	NoteVisibility string           `json:"noteVisibility"`
	GraphOptions   GraphOptions     `json:"graphOptions"`
}

// RenamableField is a column/row field in a DashboardCell of type Table
type RenamableField struct {
	InternalName string `json:"internalName"`
	DisplayName  string `json:"displayName"`
	Visible      bool   `json:"visible"`
	Direction    string `json:"direction"`
	TempVar      string `json:"tempVar"`
}

// GraphOptions is a type of options for a DashboardCell for graph
type GraphOptions struct {
	FillArea         bool   `json:"fillArea"`
	ShowLine         bool   `json:"showLine"`
	ShowPoint        bool   `json:"showPoint"`
	ShowTempVarCount string `json:"showTempVarCount"`
}

// TableOptions is a type of options for a DashboardCell with type Table
type TableOptions struct {
	VerticalTimeAxis bool           `json:"verticalTimeAxis"`
	SortBy           RenamableField `json:"sortBy"`
	Wrapping         string         `json:"wrapping"`
	FixFirstColumn   bool           `json:"fixFirstColumn"`
}

// DecimalPlaces indicates whether decimal places should be enforced, and how many digits it should show.
type DecimalPlaces struct {
	IsEnforced bool  `json:"isEnforced"`
	Digits     int32 `json:"digits"`
}

// DashboardsStore is the storage and retrieval of dashboards
type DashboardsStore interface {
	// All lists all dashboards from the DashboardStore
	All(context.Context) ([]Dashboard, error)
	// Create a new Dashboard in the DashboardStore
	Add(context.Context, Dashboard) (Dashboard, error)
	// Delete the Dashboard from the DashboardStore if `ID` exists.
	Delete(context.Context, Dashboard) error
	// Get retrieves a dashboard if `ID` exists.
	Get(ctx context.Context, id DashboardID) (Dashboard, error)
	// Update replaces the dashboard information
	Update(context.Context, Dashboard) error
}

// Cell is a rectangle and multiple time series queries to visualize.
type Cell struct {
	X            int32            `json:"x"`
	Y            int32            `json:"y"`
	W            int32            `json:"w"`
	H            int32            `json:"h"`
	I            string           `json:"i"`
	Name         string           `json:"name"`
	Queries      []Query          `json:"queries"`
	Axes         map[string]Axis  `json:"axes"`
	Type         string           `json:"type"`
	CellColors   []CellColor      `json:"colors"`
	GraphOptions GraphOptions     `json:"graphOptions"`
	TableOptions TableOptions     `json:"tableOptions,omitempty"`
	FieldOptions []RenamableField `json:"fieldOptions,omitempty"`
}

// Layout is a collection of Cells for visualization
type Layout struct {
	ID          string `json:"id"`
	Application string `json:"app"`
	Measurement string `json:"measurement"`
	Autoflow    bool   `json:"autoflow"`
	Cells       []Cell `json:"cells"`
}

// UnmarshalJSON GraphOptions setting default values for missing fields.
func (c *Cell) UnmarshalJSON(data []byte) error {
	type Alias Cell
	cell := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &cell); err != nil {
		return err
	}

	if cell.GraphOptions == (GraphOptions{}) {
		c.GraphOptions = GraphOptions{
			FillArea:         false,
			ShowLine:         true,
			ShowPoint:        false,
			ShowTempVarCount: "",
		}
	}

	return nil
}

// LayoutsStore stores dashboards and associated Cells
type LayoutsStore interface {
	// All returns all dashboards in the store
	All(context.Context) ([]Layout, error)
	// Get retrieves Layout if `ID` exists
	Get(ctx context.Context, ID string) (Layout, error)
}

// ProtoboardMeta is the metadata of a Protoboard
type ProtoboardMeta struct {
	Name             string   `json:"name"`
	Icon             string   `json:"icon,omitempty"`
	Version          string   `json:"version"`
	Measurements     []string `json:"measurements"`
	DashboardVersion string   `json:"dashboardVersion"`
	Description      string   `json:"description,omitempty"`
	Author           string   `json:"author,omitempty"`
	License          string   `json:"license,omitempty"`
	URL              string   `json:"url,omitempty"`
}

// ProtoboardCell holds visual and query information for a cell
type ProtoboardCell struct {
	X              int32            `json:"x"`
	Y              int32            `json:"y"`
	W              int32            `json:"w"`
	H              int32            `json:"h"`
	MinW           int32            `json:"minW"`
	MinH           int32            `json:"minH"`
	Name           string           `json:"name"`
	Queries        []DashboardQuery `json:"queries"`
	Axes           map[string]Axis  `json:"axes"`
	Type           string           `json:"type"`
	CellColors     []CellColor      `json:"colors"`
	Legend         Legend           `json:"legend"`
	TableOptions   TableOptions     `json:"tableOptions,omitempty"`
	FieldOptions   []RenamableField `json:"fieldOptions"`
	TimeFormat     string           `json:"timeFormat"`
	DecimalPlaces  DecimalPlaces    `json:"decimalPlaces"`
	Note           string           `json:"note"`
	NoteVisibility string           `json:"noteVisibility"`
	GraphOptions   GraphOptions     `json:"graphOptions"`
}

// ProtoboardData is the data of a Protoboard that can be instantiated into a dashboard, including a collection of cells
type ProtoboardData struct {
	Cells     []ProtoboardCell `json:"cells"`
	Templates []Template       `json:"templates"`
}

// Protoboard is a prototype of a dashboard that can be instantiated
type Protoboard struct {
	ID   string         `json:"id"`
	Meta ProtoboardMeta `json:"meta"`
	Data ProtoboardData `json:"data"`
}

// ProtoboardsStore stores protoboards that can be instantiated into dashboards
type ProtoboardsStore interface {
	// All returns all protoboards in the store
	All(context.Context) ([]Protoboard, error)
	// Get returns the specified protoboard from the store
	Get(ctx context.Context, ID string) (Protoboard, error)
}

// MappingWildcard is the wildcard value for mappings
const MappingWildcard string = "*"

// A Mapping is the structure that is used to determine a users
// role within an organization. The high level idea is to grant
// certain roles to certain users without them having to be given
// explicit role within the organization.
//
// One can think of a mapping like so:
//
//	Provider:Scheme:Group -> Organization
//	github:oauth2:cloudhub -> Happy
//	beyondcorp:ldap:cloudhub -> TheBillHilliettas
//
// Any of Provider, Scheme, or Group may be provided as a wildcard *
//
//	github:oauth2:* -> MyOrg
//	*:*:* -> AllOrg
type Mapping struct {
	ID                   string `json:"id"`
	Organization         string `json:"organizationId"`
	Provider             string `json:"provider"`
	Scheme               string `json:"scheme"`
	ProviderOrganization string `json:"providerOrganization"`
}

// MappingsStore is the storage and retrieval of Mappings
type MappingsStore interface {
	// Add creates a new Mapping.
	// The Created mapping is returned back to the user with the
	// ID field populated.
	Add(context.Context, *Mapping) (*Mapping, error)
	// All lists all Mapping in the MappingsStore
	All(context.Context) ([]Mapping, error)
	// Delete removes an Mapping from the MappingsStore
	Delete(context.Context, *Mapping) error
	// Get retrieves an Mapping from the MappingsStore
	Get(context.Context, string) (*Mapping, error)
	// Update updates an Mapping in the MappingsStore
	Update(context.Context, *Mapping) error
}

// Organization is a group of resources under a common name
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// DefaultRole is the name of the role that is the default for any users added to the organization
	DefaultRole string `json:"defaultRole,omitempty"`
}

// OrganizationQuery represents the attributes that a organization may be retrieved by.
// It is predominantly used in the OrganizationsStore.Get method.
// It is expected that only one of ID or Name will be specified, but will prefer ID over Name if both are specified.
type OrganizationQuery struct {
	// If an ID is provided in the query, the lookup time for an organization will be O(1).
	ID *string
	// If Name is provided, the lookup time will be O(n).
	Name *string
}

// OrganizationsStore is the storage and retrieval of Organizations
//
// While not necessary for the app to function correctly, it is
// expected that Implementors of the OrganizationsStore will take
// care to guarantee that the Organization.Name is unqiue. Allowing
// for duplicate names creates a confusing UX experience for the User.
type OrganizationsStore interface {
	// Add creates a new Organization.
	// The Created organization is returned back to the user with the
	// ID field populated.
	Add(context.Context, *Organization) (*Organization, error)
	// All lists all Organizations in the OrganizationsStore
	All(context.Context) ([]Organization, error)
	// Delete removes an Organization from the OrganizationsStore
	Delete(context.Context, *Organization) error
	// Get retrieves an Organization from the OrganizationsStore
	Get(context.Context, OrganizationQuery) (*Organization, error)
	// Update updates an Organization in the OrganizationsStore
	Update(context.Context, *Organization) error
	// CreateDefault creates the default organization
	CreateDefault(ctx context.Context) error
	// DefaultOrganization returns the DefaultOrganization
	DefaultOrganization(ctx context.Context) (*Organization, error)
}

// Config is the global application Config for parameters that can be set via
// API, with different sections, such as Auth
type Config struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig is the global application config section for auth parameters
type AuthConfig struct {
	// SuperAdminNewUsers configuration option that specifies which users will auto become super admin
	SuperAdminNewUsers bool `json:"superAdminNewUsers"`
}

// ConfigStore is the storage and retrieval of global application Config
type ConfigStore interface {
	// All lists all Configs in the ConfigStore
	All(context.Context) ([]Config, error)
	// Get retrieves the whole Config from the ConfigStore
	Get(context.Context) (*Config, error)
	// Update updates the whole Config in the ConfigStore
	Update(context.Context, *Config) error
}

// OrganizationConfig is the organization config for parameters that can
// be set via API, with different sections, such as LogViewer
type OrganizationConfig struct {
	OrganizationID string          `json:"organization"`
	LogViewer      LogViewerConfig `json:"logViewer"`
}

// LogViewerConfig is the configuration settings for the Log Viewer UI
type LogViewerConfig struct {
	Columns []LogViewerColumn `json:"columns"`
}

// LogViewerColumn is a specific column of the Log Viewer UI
type LogViewerColumn struct {
	Name      string           `json:"name"`
	Position  int32            `json:"position"`
	Encodings []ColumnEncoding `json:"encodings"`
}

// ColumnEncoding is the settings for a specific column of the Log Viewer UI
type ColumnEncoding struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Name  string `json:"name,omitempty"`
}

// OrganizationConfigStore is the storage and retrieval of organization Configs
type OrganizationConfigStore interface {
	// All lists all org configs in the OrganizationConfigStore
	All(context.Context) ([]OrganizationConfig, error)
	// FindOrCreate gets an existing OrganizationConfig and creates one if none exists
	FindOrCreate(ctx context.Context, orgID string) (*OrganizationConfig, error)
	// Put replaces the whole organization config in the OrganizationConfigStore
	Put(context.Context, *OrganizationConfig) error
}

// BuildInfo is sent to the usage client to track versions and commits
type BuildInfo struct {
	Version string
	Commit  string
}

// BuildStore is the storage and retrieval of CloudHub build information
type BuildStore interface {
	Get(context.Context) (BuildInfo, error)
	Update(context.Context, BuildInfo) error
}

// Vsphere represents an vsphere
type Vsphere struct {
	ID           string `json:"id,string,omitempty"`
	Host         string `json:"host,string"`
	UserName     string `json:"username,string"`
	Password     string `json:"password"`
	Protocol     string `json:"protocol,omitempty"`
	Port         int    `json:"port,omitempty"`
	Interval     int    `json:"interval"`
	Minion       string `json:"minion"`
	Organization string `json:"organization"`
	DataSource   string `json:"datasource"`
}

// VspheresStore is the Storage and retrieval of information
type VspheresStore interface {
	// All lists all vSpheres from the VspheresStore
	All(context.Context) ([]Vsphere, error)
	// Create a new vSphere in the VspheresStore
	Add(context.Context, Vsphere) (Vsphere, error)
	// Delete the Vsphere from the VspheresStore
	Delete(context.Context, Vsphere) error
	// Get retrieves a vSphere if `ID` exists.
	Get(context.Context, string) (Vsphere, error)
	// Update replaces the vSphere information
	Update(context.Context, Vsphere) error
}

// Environment is the set of front-end exposed environment variables
// that were set on the server
type Environment struct {
	TelegrafSystemInterval time.Duration `json:"telegrafSystemInterval"`
	CustomAutoRefresh      string        `json:"customAutoRefresh,omitempty"`
}

// The InternalEnvironment variable is an internally shared environment variable within the server.
type InternalEnvironment struct {
	EtcdEndpoints    []string
	TemplatesPath    string
	TemplatesManager TemplatesManager
	AIConfig         AIConfig
}

// Topology is represents represents an topology
type Topology struct {
	ID              string          `json:"id,string,omitempty"`
	Organization    string          `json:"organization,omitempty"`    // Organization is the organization ID that resource belongs to
	Diagram         string          `json:"diagram,string,omitempty"`  // diagram xml
	Preferences     []string        `json:"preferences,omitempty"`     // User preferences
	TopologyOptions TopologyOptions `json:"topologyOptions,omitempty"` // Configuration options for the topology, defined in TopologyOptions
}

// TopologyOptions represents various settings for displaying elements of the topology.
// Each field controls the visibility of specific icons or features within the topology.
type TopologyOptions struct {
	MinimapVisible    bool `json:"minimapVisible"`    // Controls whether the minimap is visible in the mxgraph
	HostStatusVisible bool `json:"hostStatusVisible"` // Controls whether the host status is visible
	IPMIVisible       bool `json:"ipmiVisible"`       // Controls whether the IPMI icon is visible
	LinkVisible       bool `json:"linkVisible"`       // Controls whether the dashboard link icon is visible
}

// TopologyQuery represents the attributes that a topology may be retrieved by.
// It is predominantly used in the TopologiesStore.Get method.
//
// It is expected that only one of ID or Organization will be
// specified, but all are provided TopologiesStore should prefer ID.
type TopologyQuery struct {
	ID           *string
	Organization *string
}

// TopologiesStore is the Storage and retrieval of information
type TopologiesStore interface {
	// All lists all topologies from the TopologiesStore
	All(context.Context) ([]Topology, error)
	// Create a new topology in the TopologiesStore
	Add(context.Context, *Topology) (*Topology, error)
	// Delete the topology from the TopologiesStore
	Delete(context.Context, *Topology) error
	// Get retrieves a topology if `ID` exists.
	Get(ctx context.Context, q TopologyQuery) (*Topology, error)
	// Update replaces the topology information
	Update(context.Context, *Topology) error
}

// The kinds of CSP.
const (
	AWS   = "aws"
	AZURE = "azure"
	GCP   = "gcp"
	OSP   = "osp"
	OCP   = "ocp"
)

// CSPQuery represents the attributes that a CSP may be retrieved by.
// It is predominantly used in the CSPStore.Get method.
//
// It is expected that only one of ID or Organization will be
// specified, but all are provided CSPStore should prefer ID.
type CSPQuery struct {
	ID           *string
	Organization *string
}

// CSP is CSP connection information
type CSP struct {
	ID       string `json:"id,string,omitempty"`
	Provider string `json:"provider,string"` // aws, gcp, azure, osp, ocp, and so on.
	// if provider=aws, Region
	// if provider=gcp, Project ID
	// if provider=osp, project_name
	NameSpace string `json:"namespace,string"`
	// if provider=aws, Access Key
	// if provider=osp, username
	AccessKey string `json:"accesskey,string"`
	// if provider=aws, Secret Key
	// if provider=osp, password
	SecretKey    string `json:"secretkey,string"`
	Organization string `json:"organization"`
	Minion       string `json:"minion,string"`
}

// CSPStore is the Storage and retrieval of information
type CSPStore interface {
	// All lists all CSP from the CSPStore
	All(context.Context) ([]CSP, error)
	// Create a new CSP in the CSPStore
	Add(context.Context, *CSP) (*CSP, error)
	// Delete the CSP from the CSPStore
	Delete(context.Context, *CSP) error
	// Get retrieves a CSP if `ID` exists.
	Get(ctx context.Context, q CSPQuery) (*CSP, error)
	// Update replaces the CSP information
	Update(context.Context, *CSP) error
}

// KVClient defines what each kv store should be capable of.
type KVClient interface {
	// ConfigStore returns the kv's ConfigStore type.
	ConfigStore() ConfigStore
	// DashboardsStore returns the kv's DashboardsStore type.
	DashboardsStore() DashboardsStore
	// MappingsStore returns the kv's MappingsStore type.
	MappingsStore() MappingsStore
	// OrganizationConfigStore returns the kv's OrganizationConfigStore type.
	OrganizationConfigStore() OrganizationConfigStore
	// OrganizationsStore returns the kv's OrganizationsStore type.
	OrganizationsStore() OrganizationsStore
	// ServersStore returns the kv's ServersStore type.
	ServersStore() ServersStore
	// SourcesStore returns the kv's SourcesStore type.
	SourcesStore() SourcesStore
	// UsersStore returns the kv's UsersStore type.
	UsersStore() UsersStore
	// VspheresStore returns the kv's VspheresStore type.
	VspheresStore() VspheresStore
	// TopologiesStore returns the kv's TopologiesStore type.
	TopologiesStore() TopologiesStore
	// CSPStore returns the kv's CSPStore type.
	CSPStore() CSPStore
	// NetworkDeviceStore returns the kv's NetworkDeviceStore type.
	NetworkDeviceStore() NetworkDeviceStore
	// NetworkDeviceOrg returns the kv's NetworkDeviceOrg type.
	NetworkDeviceOrgStore() NetworkDeviceOrgStore
	// MLNxRstStore returns the kv's MLNxRstStore type.
	MLNxRstStore() MLNxRstStore
}

// NetworkDeviceOrgQuery represents the attributes that a networkDeviceOrg may be retrieved by.
// It is predominantly used in the networkDeviceOrgStore.Get method.
//
// It is expected that only one of Organization ID will be
// specified, but all are provided networkDeviceOrgStore should prefer ID.
type NetworkDeviceOrgQuery struct {
	ID *string
}

// WorkerLimit controls the number of concurrent goroutines using a semaphore
const (
	WorkerLimit = 10
)

// AITemplates is a config Template for Cloudhub AI
type AITemplates struct {
	ID          string `json:"id"`
	Application string `json:"app"`
}

// AIKapacitor represents the information for Kapacitor login
type AIKapacitor struct {
	SrcID              int    `json:"srcId,string"`  // SrcID of the data source
	KapaID             int    `json:"kapaId,string"` // KapaID of the Kapacitor ID
	KapaURL            string `json:"url"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

// NetworkDeviceOrg represents the information of a network device group
type NetworkDeviceOrg struct {
	ID                  string      `json:"organization"`
	LoadModule          string      `json:"load_module"`
	MLFunction          string      `json:"ml_function"`
	DataDuration        int         `json:"data_duration"`
	LearnedDevicesIDs   []string    `json:"learned_devices_ids"`
	CollectorServer     string      `json:"collector_server"`
	CollectedDevicesIDs []string    `json:"collected_devices_ids"`
	AIKapacitor         AIKapacitor `json:"ai_kapacitor"`
	LearningCron        string      `json:"learning_cron"`
	ProcCnt             int         `json:"process_count"`
}

// NetworkDeviceOrgStore is the Storage and retrieval of information
type NetworkDeviceOrgStore interface {
	All(context.Context) ([]NetworkDeviceOrg, error)

	Add(context.Context, *NetworkDeviceOrg) (*NetworkDeviceOrg, error)

	Delete(context.Context, *NetworkDeviceOrg) error

	Get(ctx context.Context, q NetworkDeviceOrgQuery) (*NetworkDeviceOrg, error)

	Update(context.Context, *NetworkDeviceOrg) error
}

// DeviceCategoryMap maps device category keys to their corresponding category names.
var DeviceCategoryMap = map[string]string{
	"server":  "server",
	"network": "network",
}

// NetworkDeviceQuery represents the attributes that a NetworkDevice may be retrieved by.
// It is predominantly used in the NetworkDeviceStore.Get method.
//
// It is expected that only one of ID or Organization will be
// specified, but all are provided NetworkDeviceStore should prefer ID.
type NetworkDeviceQuery struct {
	ID           *string
	Organization *string
}

// SSHConfig is Connection Config
type SSHConfig struct {
	UserID     string `json:"user_id"`
	Password   string `json:"password"`
	EnPassword string `json:"en_password"`
	Port       int    `json:"port"`
}

// SNMPConfig is Connection Config
type SNMPConfig struct {
	Community     string `json:"community"`
	Version       string `json:"version"`
	Port          int    `json:"port"`
	Protocol      string `json:"protocol"`
	SecurityName  string `json:"security_name,omitempty"`
	AuthProtocol  string `json:"auth_protocol,omitempty"` // auth protocol one of ["md5", "sha", "sha2", "hmac128sha224", "hmac192sha256", "hmac256sha384", "hmac384sha512"]
	AuthPass      string `json:"auth_pass,omitempty"`
	PrivProtocol  string `json:"priv_protocol,omitempty"` // priv_protocol one of ["des", "aes", "aes128", "aes192", "aes256"]
	PrivPass      string `json:"priv_pass,omitempty"`
	SecurityLevel string `json:"security_level,omitempty"` // security_level one of ["noAuthNoPriv", "authNoPriv", "authPriv"]
}

// NetworkDevice represents the information of a network device
type NetworkDevice struct {
	ID                     string     `json:"id,omitempty"`
	Organization           string     `json:"organization"`
	DeviceIP               string     `json:"device_ip"`
	Hostname               string     `json:"hostname"`
	DeviceType             string     `json:"device_type"`
	DeviceCategory         string     `json:"device_category"`
	DeviceOS               string     `json:"device_os"`
	IsCollectingCfgWritten bool       `json:"is_collecting_cfg_written"`
	SSHConfig              SSHConfig  `json:"ssh_config"`
	SNMPConfig             SNMPConfig `json:"snmp_config"`
	Sensitivity            float32    `json:"sensitivity"`
	DeviceVendor           string     `json:"device_vendor"`
	LearningState          string     `json:"learning_state"`
	LearningBeginDatetime  string     `json:"learning_begin_datetime"`
	LearningFinishDatetime string     `json:"learning_finish_datetime"`
	IsLearning             bool       `json:"is_learning"`
}

// NetworkDeviceStore is the Storage and retrieval of information
type NetworkDeviceStore interface {
	All(context.Context) ([]NetworkDevice, error)

	Add(context.Context, *NetworkDevice) (*NetworkDevice, error)

	Delete(context.Context, *NetworkDevice) error

	Get(ctx context.Context, q NetworkDeviceQuery) (*NetworkDevice, error)

	Update(context.Context, *NetworkDevice) error
}

// ConfigTemplate represents a configuration template for a task
type ConfigTemplate struct {
	ID       string `json:"id"`
	App      string `json:"app"`
	Template string `json:"template"`
}

// ConfigTemplatesStore stores configuration templates
type ConfigTemplatesStore interface {
	// All returns all templates in the store
	All(ctx context.Context) ([]ConfigTemplate, error)
	// Get retrieves ConfigTemplate if `ID` exists
	Get(ctx context.Context, ID string) (ConfigTemplate, error)
}

// TemplatesManager is Template config Manager
type TemplatesManager interface {
	All(ctx context.Context) ([]ConfigTemplate, error)
	Get(ctx context.Context, id string) (ConfigTemplate, error)
}

// TemplateLoader is Template Loader for CLoudhub AI Config
type TemplateLoader interface {
	LoadTemplate(config LoadTemplateConfig) (*template.Template, error)
}

// AIConfig is to The Information to access to cloudhub AI
type AIConfig struct {
	DockerPath      string `json:"docker-path"`
	DockerCmd       string `json:"docker-cmd"`
	LogstashPath    string `json:"logstash-path"`
	PredictionRegex string `json:"prediction-regex"`
}

// MLNxRstQuery represents the attributes that a MLNxRst may be retrieved by.
// It is predominantly used in the MLNxRstStore.Get method.
type MLNxRstQuery struct {
	ID *string
}

// MLNxRstStore is the Storage and retrieval of information
type MLNxRstStore interface {
	All(context.Context) ([]MLNxRst, error)

	Add(context.Context, *MLNxRst) (*MLNxRst, error)

	Delete(context.Context, *MLNxRst) error

	Get(ctx context.Context, q MLNxRstQuery) (*MLNxRst, error)

	Update(context.Context, *MLNxRst) error
}

// MLNxRst represents the result of a Machine Learning (ML) process
type MLNxRst struct {
	Device                 string    `json:"device_ip"`                // IP address or ID of the device
	LearningFinishDatetime string    `json:"learning_finish_datetime"` // TZ=UTC, Format=RFC3339
	Epsilon                float64   `json:"epsilon"`                  // ML Result value
	MeanMatrix             string    `json:"mean_matrix"`              // 1x2 mean numpy matrix
	CovarianceMatrix       string    `json:"covariance_matrix"`        // 2x2 covariance numpy matrix
	K                      float32   `json:"k"`                        // Decision coefficient for determination of threshold
	Mean                   float32   `json:"mean"`                     // Mean value by whole elements
	MDThreshold            float32   `json:"md_threshold"`             // MDThreshold = mean * K * Sensitivity
	MDArray                []float32 `json:"md_array"`                 // Mahalanobis distance array
	CPUArray               []float32 `json:"cpu_array"`                // Use Gaussian Graph
	TrafficArray           []float32 `json:"traffic_array"`            // Use Gaussian Graph
	GaussianArray          []float32 `json:"gaussian_array"`           // Use Gaussian Graph
}

// DLNxRstQuery represents the attributes that a DLNxRst may be retrieved by.
// It is predominantly used in the DLNxRstStore.Get method.
type DLNxRstQuery struct {
	ID *string
}

// DLNxRstStore is the Storage and retrieval of information
type DLNxRstStore interface {
	All(context.Context) ([]DLNxRst, error)

	Add(context.Context, *DLNxRst) (*DLNxRst, error)

	Delete(context.Context, *DLNxRst) error

	Get(ctx context.Context, q DLNxRstQuery) (*DLNxRst, error)

	Update(context.Context, *DLNxRst) error
}

// DLNxRst represents the result of a deep learning process
type DLNxRst struct {
	Device                 string    `json:"device"`                   // IP address of the device
	LearningFinishDatetime string    `json:"learning_finish_datetime"` // TZ=UTC, Format=RFC3339
	DLThreshold            float32   `json:"dl_threshold"`             // DL Threshold value
	TrainLoss              []float32 `json:"train_loss"`               // Use Loss Graph
	ValidLoss              []float32 `json:"valid_loss"`               // Use Loss Graph
	MSE                    []float32 `json:"mse"`                      // Use Mean Squared Error Graph
}

// DLNxRstStgQuery represents the attributes that a DLNxRst may be retrieved by.
// It is predominantly used in the DLNxRstStgStore.Get method.
type DLNxRstStgQuery struct {
	ID *string
}

// DLNxRstStgStore is the Storage and retrieval of information
type DLNxRstStgStore interface {
	Delete(ctx context.Context, q DLNxRstStgQuery) error
}

// DLNxRstStg represents the result of a deep learning process
type DLNxRstStg struct {
	Device                 string  `json:"device"`                   // IP address of the device
	LearningFinishDatetime string  `json:"learning_finish_datetime"` // TZ=UTC, Format=RFC3339
	Scaler                 []byte  `json:"scaler"`
	Model                  []byte  `json:"model"`
	DLThreshold            float32 `json:"dl_threshold"` // DL Threshold value
}

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	_ "net/http/pprof" // required for /debug/pprof endpoint

	"github.com/NYTimes/gziphandler"
	basicAuth "github.com/abbot/go-http-auth"
	"github.com/bouk/httprouter"
	cloudhub "github.com/snetsystems/cloudhub/backend"
	"github.com/snetsystems/cloudhub/backend/oauth2"
	"github.com/snetsystems/cloudhub/backend/roles"
)

const (
	// JSONType the mimetype for a json request
	JSONType = "application/json"
)

// MuxOpts are the options for the router.  Mostly related to auth.
type MuxOpts struct {
	Logger                cloudhub.Logger
	Develop               bool                 // Develop loads assets from filesystem instead of bindata
	Basepath              string               // URL path prefix under which all cloudhub routes will be mounted
	UseAuth               bool                 // UseAuth turns on Github OAuth and JWT
	Auth                  oauth2.Authenticator // Auth is used to authenticate and authorize
	ProviderFuncs         []func(func(oauth2.Provider, oauth2.Mux))
	StatusFeedURL         string               // JSON Feed URL for the client Status page News Feed
	CustomLinks           []CustomLink         // Any custom external links for client's User menu
	PprofEnabled          bool                 // Mount pprof routes for profiling
	DisableGZip           bool                 // Optionally disable gzip.
	BasicAuth             *basicAuth.BasicAuth // HTTP basic authentication provider
	PasswordPolicy        string               // Password validity rules
	PasswordPolicyMessage string               // Password validity rule description
}

// NewMux attaches all the route handlers; handler returned servers cloudhub.
func NewMux(opts MuxOpts, service Service) http.Handler {
	hr := httprouter.New()

	// /* React Application */
	assets := Assets(AssetsOpts{
		Develop: opts.Develop,
		Logger:  opts.Logger,
	})

	if opts.Basepath != "" {
		basePath := fmt.Sprintf("%s/", opts.Basepath)
		// Prefix any URLs found in the React assets with any configured basepath
		assets = NewDefaultURLPrefixer(basePath, assets, opts.Logger)
	}

	// Compress the assets with gzip if an accepted encoding
	if !opts.DisableGZip {
		assets = gziphandler.GzipHandler(assets)
	}

	// The react application handles all the routing if the server does not
	// know about the route.  This means that we never have unknown routes on
	// the server.
	hr.NotFound = assets

	var router cloudhub.Router = hr

	// Set route prefix for all routes if basepath is present
	if opts.Basepath != "" {
		router = &MountableRouter{
			Prefix:   opts.Basepath,
			Delegate: hr,
		}

		//The assets handler is always unaware of basepaths, so the
		// basepath needs to always be removed before sending requests to it
		hr.NotFound = http.StripPrefix(opts.Basepath, hr.NotFound)
	}

	EnsureMember := func(next http.HandlerFunc) http.HandlerFunc {
		return AuthorizedUser(
			service.Store,
			opts.UseAuth,
			roles.MemberRoleName,
			opts.Logger,
			next,
		)
	}
	_ = EnsureMember

	EnsureViewer := func(next http.HandlerFunc) http.HandlerFunc {
		return AuthorizedUser(
			service.Store,
			opts.UseAuth,
			roles.ViewerRoleName,
			opts.Logger,
			next,
		)
	}

	EnsureEditor := func(next http.HandlerFunc) http.HandlerFunc {
		return AuthorizedUser(
			service.Store,
			opts.UseAuth,
			roles.EditorRoleName,
			opts.Logger,
			next,
		)
	}

	EnsureAdmin := func(next http.HandlerFunc) http.HandlerFunc {
		return AuthorizedUser(
			service.Store,
			opts.UseAuth,
			roles.AdminRoleName,
			opts.Logger,
			next,
		)
	}

	EnsureSuperAdmin := func(next http.HandlerFunc) http.HandlerFunc {
		return AuthorizedUser(
			service.Store,
			opts.UseAuth,
			roles.SuperAdminStatus,
			opts.Logger,
			next,
		)
	}

	rawStoreAccess := func(next http.HandlerFunc) http.HandlerFunc {
		return RawStoreAccess(opts.Logger, next)
	}

	ensureOrgMatches := func(next http.HandlerFunc) http.HandlerFunc {
		return RouteMatchesPrincipal(
			service.Store,
			opts.UseAuth,
			opts.Logger,
			next,
		)
	}

	if opts.PprofEnabled {
		// add profiling routes
		router.GET("/debug/pprof/:thing", http.DefaultServeMux.ServeHTTP)
	}

	/* Documentation */
	router.GET("/swagger.json", Spec())
	router.GET("/docs", Redoc("/swagger.json"))

	// websocket
	router.GET("/cloudhub/v1/WebTerminalHandler", EnsureAdmin(service.WebTerminalHandler))

	/* Health */
	router.GET("/ping", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	/* API (Provider=cloudhub, Scheme=basic)  */
	// Login, Logout
	router.POST("/basic/login", service.Login(opts.Auth, opts.Basepath))
	router.GET("/basic/logout", service.Logout(opts.Auth, opts.Basepath))

	// User sign up
	router.POST("/basic/users", service.NewBasicUser)

	// User password change
	router.PATCH("/basic/password", service.UserPassword)

	// User password change
	router.PATCH("/cloudhub/v1/basic/password", EnsureAdmin(service.UserPassword))

	// User password reset
	router.GET("/basic/password/reset", service.UserPwdReset)
	router.GET("/cloudhub/v1/password/reset", EnsureAdmin(service.UserPwdAdminReset))

	/* API */
	// Organizations
	router.GET("/cloudhub/v1/organizations", EnsureViewer(service.Organizations))
	router.POST("/cloudhub/v1/organizations", EnsureSuperAdmin(service.NewOrganization))

	router.GET("/cloudhub/v1/organizations/:oid", EnsureAdmin(service.OrganizationID))
	router.PATCH("/cloudhub/v1/organizations/:oid", EnsureSuperAdmin(service.UpdateOrganization))
	router.DELETE("/cloudhub/v1/organizations/:oid", EnsureSuperAdmin(service.RemoveOrganization))

	// Mappings
	router.GET("/cloudhub/v1/mappings", EnsureSuperAdmin(service.Mappings))
	router.POST("/cloudhub/v1/mappings", EnsureSuperAdmin(service.NewMapping))

	router.PUT("/cloudhub/v1/mappings/:id", EnsureSuperAdmin(service.UpdateMapping))
	router.DELETE("/cloudhub/v1/mappings/:id", EnsureSuperAdmin(service.RemoveMapping))

	// Sources
	router.GET("/cloudhub/v1/sources", EnsureViewer(service.Sources))
	router.POST("/cloudhub/v1/sources", EnsureEditor(service.NewSource))

	router.GET("/cloudhub/v1/sources/:id", EnsureViewer(service.SourcesID))
	router.PATCH("/cloudhub/v1/sources/:id", EnsureEditor(service.UpdateSource))
	router.DELETE("/cloudhub/v1/sources/:id", EnsureEditor(service.RemoveSource))
	router.GET("/cloudhub/v1/sources/:id/health", EnsureViewer(service.SourceHealth))

	// Flux
	router.GET("/cloudhub/v1/flux", EnsureViewer(service.Flux))
	router.POST("/cloudhub/v1/flux/ast", EnsureViewer(service.FluxAST))
	router.GET("/cloudhub/v1/flux/suggestions", EnsureViewer(service.FluxSuggestions))
	router.GET("/cloudhub/v1/flux/suggestions/:name", EnsureViewer(service.FluxSuggestion))

	// Source Proxy to Influx; Has gzip compression around the handler
	influx := gziphandler.GzipHandler(EnsureViewer(service.Influx))
	router.Handler("POST", "/cloudhub/v1/sources/:id/proxy", influx)

	// Source Proxy to Influx's flux endpoint; compression because the responses from
	// flux could be large.
	router.POST("/cloudhub/v1/sources/:id/proxy/flux", EnsureViewer(service.ProxyFlux))

	// Write proxies line protocol write requests to InfluxDB
	router.POST("/cloudhub/v1/sources/:id/write", EnsureViewer(service.Write))

	// Queries is used to analyze a specific queries and does not create any
	// resources. It's a POST because Queries are POSTed to InfluxDB, but this
	// only modifies InfluxDB resources with certain metaqueries, e.g. DROP DATABASE.
	//
	// Admins should ensure that the InfluxDB source as the proper permissions
	// intended for CloudHub Users with the Viewer Role type.
	router.POST("/cloudhub/v1/sources/:id/queries", EnsureViewer(service.Queries))

	// Annotations are user-defined events associated with this source
	router.GET("/cloudhub/v1/sources/:id/annotations", EnsureViewer(service.Annotations))
	router.POST("/cloudhub/v1/sources/:id/annotations", EnsureEditor(service.NewAnnotation))
	router.GET("/cloudhub/v1/sources/:id/annotations/:aid", EnsureViewer(service.Annotation))
	router.DELETE("/cloudhub/v1/sources/:id/annotations/:aid", EnsureEditor(service.RemoveAnnotation))
	router.PATCH("/cloudhub/v1/sources/:id/annotations/:aid", EnsureEditor(service.UpdateAnnotation))

	// All possible permissions for users in this source
	router.GET("/cloudhub/v1/sources/:id/permissions", EnsureViewer(service.Permissions))

	// Users associated with the data source
	router.GET("/cloudhub/v1/sources/:id/users", EnsureAdmin(service.SourceUsers))
	router.POST("/cloudhub/v1/sources/:id/users", EnsureAdmin(service.NewSourceUser))

	router.GET("/cloudhub/v1/sources/:id/users/:uid", EnsureAdmin(service.SourceUserID))
	router.DELETE("/cloudhub/v1/sources/:id/users/:uid", EnsureAdmin(service.RemoveSourceUser))
	router.PATCH("/cloudhub/v1/sources/:id/users/:uid", EnsureAdmin(service.UpdateSourceUser))

	// Roles associated with the data source
	router.GET("/cloudhub/v1/sources/:id/roles", EnsureViewer(service.SourceRoles))
	router.POST("/cloudhub/v1/sources/:id/roles", EnsureEditor(service.NewSourceRole))

	router.GET("/cloudhub/v1/sources/:id/roles/:rid", EnsureViewer(service.SourceRoleID))
	router.DELETE("/cloudhub/v1/sources/:id/roles/:rid", EnsureEditor(service.RemoveSourceRole))
	router.PATCH("/cloudhub/v1/sources/:id/roles/:rid", EnsureEditor(service.UpdateSourceRole))

	// Services are resources that cloudhub proxies to
	router.GET("/cloudhub/v1/sources/:id/services", EnsureViewer(service.Services))
	router.POST("/cloudhub/v1/sources/:id/services", EnsureEditor(service.NewService))
	router.GET("/cloudhub/v1/sources/:id/services/:kid", EnsureViewer(service.ServiceID))
	router.PATCH("/cloudhub/v1/sources/:id/services/:kid", EnsureEditor(service.UpdateService))
	router.DELETE("/cloudhub/v1/sources/:id/services/:kid", EnsureEditor(service.RemoveService))

	// Service Proxy
	router.GET("/cloudhub/v1/sources/:id/services/:kid/proxy", EnsureViewer(service.ProxyGet))
	router.POST("/cloudhub/v1/sources/:id/services/:kid/proxy", EnsureEditor(service.ProxyPost))
	router.PATCH("/cloudhub/v1/sources/:id/services/:kid/proxy", EnsureEditor(service.ProxyPatch))
	router.DELETE("/cloudhub/v1/sources/:id/services/:kid/proxy", EnsureEditor(service.ProxyDelete))

	// Salt Proxy
	router.POST("/cloudhub/v1/proxy/salt", EnsureViewer(service.SaltProxyPost))

	// Kapacitor
	router.GET("/cloudhub/v1/sources/:id/kapacitors", EnsureViewer(service.Kapacitors))
	router.POST("/cloudhub/v1/sources/:id/kapacitors", EnsureEditor(service.NewKapacitor))

	router.GET("/cloudhub/v1/sources/:id/kapacitors/:kid", EnsureViewer(service.KapacitorsID))
	router.PATCH("/cloudhub/v1/sources/:id/kapacitors/:kid", EnsureEditor(service.UpdateKapacitor))
	router.DELETE("/cloudhub/v1/sources/:id/kapacitors/:kid", EnsureEditor(service.RemoveKapacitor))

	// Kapacitor rules
	router.GET("/cloudhub/v1/sources/:id/kapacitors/:kid/rules", EnsureViewer(service.KapacitorRulesGet))
	router.POST("/cloudhub/v1/sources/:id/kapacitors/:kid/rules", EnsureEditor(service.KapacitorRulesPost))

	router.GET("/cloudhub/v1/sources/:id/kapacitors/:kid/rules/:tid", EnsureViewer(service.KapacitorRulesID))
	router.PUT("/cloudhub/v1/sources/:id/kapacitors/:kid/rules/:tid", EnsureEditor(service.KapacitorRulesPut))
	router.PATCH("/cloudhub/v1/sources/:id/kapacitors/:kid/rules/:tid", EnsureEditor(service.KapacitorRulesStatus))
	router.DELETE("/cloudhub/v1/sources/:id/kapacitors/:kid/rules/:tid", EnsureEditor(service.KapacitorRulesDelete))

	// Kapacitor Proxy
	router.GET("/cloudhub/v1/sources/:id/kapacitors/:kid/proxy", EnsureViewer(service.ProxyGet))
	router.POST("/cloudhub/v1/sources/:id/kapacitors/:kid/proxy", EnsureEditor(service.ProxyPost))
	router.PATCH("/cloudhub/v1/sources/:id/kapacitors/:kid/proxy", EnsureEditor(service.ProxyPatch))
	router.DELETE("/cloudhub/v1/sources/:id/kapacitors/:kid/proxy", EnsureEditor(service.ProxyDelete))

	// Layouts
	router.GET("/cloudhub/v1/layouts", EnsureViewer(service.Layouts))
	router.GET("/cloudhub/v1/layouts/:id", EnsureViewer(service.LayoutsID))

	// Protoboards
	router.GET("/cloudhub/v1/protoboards", EnsureViewer(service.Protoboards))
	router.GET("/cloudhub/v1/protoboards/:id", EnsureViewer(service.ProtoboardsID))

	// Users associated with CloudHub
	router.GET("/cloudhub/v1/me", service.Me)

	// Set current cloudhub organization the user is logged into
	router.PUT("/cloudhub/v1/me", service.UpdateMe(opts.Auth))

	// TODO: what to do about admin's being able to set superadmin
	router.GET("/cloudhub/v1/organizations/:oid/users", EnsureAdmin(ensureOrgMatches(service.Users)))
	router.POST("/cloudhub/v1/organizations/:oid/users", EnsureAdmin(ensureOrgMatches(service.OrganizationNewUser)))

	router.GET("/cloudhub/v1/organizations/:oid/users/:id", EnsureAdmin(ensureOrgMatches(service.UserID)))
	router.DELETE("/cloudhub/v1/organizations/:oid/users/:id", EnsureAdmin(ensureOrgMatches(service.OrganizationRemoveUser)))
	router.PATCH("/cloudhub/v1/organizations/:oid/users/:id", EnsureAdmin(ensureOrgMatches(service.OrganizationUpdateUser)))

	router.GET("/cloudhub/v1/users", EnsureSuperAdmin(rawStoreAccess(service.Users)))
	router.POST("/cloudhub/v1/users", EnsureSuperAdmin(rawStoreAccess(service.NewUser)))

	router.GET("/cloudhub/v1/users/:id", EnsureViewer(service.UserID))
	router.DELETE("/cloudhub/v1/users/:id", EnsureSuperAdmin(rawStoreAccess(service.RemoveUser)))
	router.PATCH("/cloudhub/v1/users/:id", EnsureViewer(service.UpdateUser))

	// Dashboards
	router.GET("/cloudhub/v1/dashboards", EnsureViewer(service.Dashboards))
	router.POST("/cloudhub/v1/dashboards", EnsureEditor(service.NewDashboard))

	router.GET("/cloudhub/v1/dashboards/:id", EnsureViewer(service.DashboardID))
	router.DELETE("/cloudhub/v1/dashboards/:id", EnsureEditor(service.RemoveDashboard))
	router.PUT("/cloudhub/v1/dashboards/:id", EnsureEditor(service.ReplaceDashboard))
	router.PATCH("/cloudhub/v1/dashboards/:id", EnsureEditor(service.UpdateDashboard))

	// Dashboard Cells
	router.GET("/cloudhub/v1/dashboards/:id/cells", EnsureViewer(service.DashboardCells))
	router.POST("/cloudhub/v1/dashboards/:id/cells", EnsureEditor(service.NewDashboardCell))

	router.GET("/cloudhub/v1/dashboards/:id/cells/:cid", EnsureViewer(service.DashboardCellID))
	router.DELETE("/cloudhub/v1/dashboards/:id/cells/:cid", EnsureEditor(service.RemoveDashboardCell))
	router.PUT("/cloudhub/v1/dashboards/:id/cells/:cid", EnsureEditor(service.ReplaceDashboardCell))

	// Dashboard Templates
	router.GET("/cloudhub/v1/dashboards/:id/templates", EnsureViewer(service.Templates))
	router.POST("/cloudhub/v1/dashboards/:id/templates", EnsureEditor(service.NewTemplate))

	router.GET("/cloudhub/v1/dashboards/:id/templates/:tid", EnsureViewer(service.TemplateID))
	router.DELETE("/cloudhub/v1/dashboards/:id/templates/:tid", EnsureEditor(service.RemoveTemplate))
	router.PUT("/cloudhub/v1/dashboards/:id/templates/:tid", EnsureEditor(service.ReplaceTemplate))

	// Databases
	router.GET("/cloudhub/v1/sources/:id/dbs", EnsureViewer(service.GetDatabases))
	router.POST("/cloudhub/v1/sources/:id/dbs", EnsureEditor(service.NewDatabase))

	router.DELETE("/cloudhub/v1/sources/:id/dbs/:db", EnsureEditor(service.DropDatabase))

	// Retention Policies
	router.GET("/cloudhub/v1/sources/:id/dbs/:db/rps", EnsureViewer(service.RetentionPolicies))
	router.POST("/cloudhub/v1/sources/:id/dbs/:db/rps", EnsureEditor(service.NewRetentionPolicy))

	router.PUT("/cloudhub/v1/sources/:id/dbs/:db/rps/:rp", EnsureEditor(service.UpdateRetentionPolicy))
	router.DELETE("/cloudhub/v1/sources/:id/dbs/:db/rps/:rp", EnsureEditor(service.DropRetentionPolicy))

	// Measurements
	router.GET("/cloudhub/v1/sources/:id/dbs/:db/measurements", EnsureViewer(service.Measurements))

	// Global application config for CloudHub
	router.GET("/cloudhub/v1/config", EnsureSuperAdmin(service.Config))
	router.GET("/cloudhub/v1/config/auth", EnsureSuperAdmin(service.AuthConfig))
	router.PUT("/cloudhub/v1/config/auth", EnsureSuperAdmin(service.ReplaceAuthConfig))

	// Organization config settings for CloudHub
	router.GET("/cloudhub/v1/org_config", EnsureViewer(service.OrganizationConfig))
	router.GET("/cloudhub/v1/org_config/logviewer", EnsureViewer(service.OrganizationLogViewerConfig))
	router.PUT("/cloudhub/v1/org_config/logviewer", EnsureEditor(service.ReplaceOrganizationLogViewerConfig))

	router.GET("/cloudhub/v1/env", EnsureViewer(service.Environment))

	// vSpheres
	router.GET("/cloudhub/v1/vspheres", EnsureViewer(service.Vspheres))
	router.GET("/cloudhub/v1/vspheres/:id", EnsureViewer(service.VsphereID))
	router.POST("/cloudhub/v1/vspheres", EnsureAdmin(service.NewVsphere))
	router.DELETE("/cloudhub/v1/vspheres/:id", EnsureAdmin(service.RemoveVsphere))
	router.PATCH("/cloudhub/v1/vspheres/:id", EnsureAdmin(service.UpdateVsphere))

	// topologies
	router.GET("/cloudhub/v1/topologies", EnsureViewer(service.Topology))
	router.POST("/cloudhub/v1/topologies", EnsureViewer(service.NewTopology))
	router.DELETE("/cloudhub/v1/topologies/:id", EnsureViewer(service.RemoveTopology))
	router.PATCH("/cloudhub/v1/topologies/:id", EnsureViewer(service.UpdateTopology))

	// Cloud Solution Provider
	router.GET("/cloudhub/v1/csp", EnsureViewer(service.CSP))
	router.GET("/cloudhub/v1/csp/:id", EnsureViewer(service.CSPID))
	router.POST("/cloudhub/v1/csp", EnsureAdmin(service.NewCSP))
	router.DELETE("/cloudhub/v1/csp/:id", EnsureAdmin(service.RemoveCSP))
	router.PATCH("/cloudhub/v1/csp/:id", EnsureAdmin(service.UpdateCSP))

	// Device Management
	router.GET("/cloudhub/v1/ai/network/managements/devices", EnsureViewer(service.AllDevices))
	router.GET("/cloudhub/v1/ai/network/managements/devices/:id", EnsureViewer(service.DeviceID))
	router.POST("/cloudhub/v1/ai/network/managements/devices", EnsureAdmin(service.NewDevices))
	router.DELETE("/cloudhub/v1/ai/network/managements/devices", EnsureAdmin(service.RemoveDevices))
	router.PATCH("/cloudhub/v1/ai/network/managements/devices/:id", EnsureAdmin(service.UpdateNetworkDevice))

	// Device Management Monitoring
	router.POST("/cloudhub/v1/ai/network/managements/monitoring/config", EnsureAdmin(service.MonitoringConfigManagement))

	// Device Management Learning
	router.POST("/cloudhub/v1/ai/network/managements/learning/config", EnsureAdmin(service.LearningDeviceManagement))

	// Device Orgs Management
	router.GET("/cloudhub/v1/ai/network/managements/orgs", EnsureViewer(service.AllDevicesOrg))
	router.GET("/cloudhub/v1/ai/network/managements/orgs/:id", EnsureViewer(service.NetworkDeviceOrgID))
	router.POST("/cloudhub/v1/ai/network/managements/orgs/", EnsureAdmin(service.AddNetworkDeviceOrg))
	router.PATCH("/cloudhub/v1/ai/network/managements/orgs/:id", EnsureAdmin(service.UpdateNetworkDeviceOrg))
	router.DELETE("/cloudhub/v1/ai/network/managements/orgs/:id", EnsureAdmin(service.RemoveNetworkDeviceOrg))
	// SNMP Management
	router.POST("/cloudhub/v1/snmp/validation", EnsureViewer(service.SNMPConnTestBulk))

	// http logging
	router.POST("/cloudhub/v1/logging", EnsureViewer(service.HTTPLogging))

	// login locked
	router.PATCH("/cloudhub/v1/login/locked", EnsureAdmin(service.LockedUser))

	// Validates go templates for the js client
	router.POST("/cloudhub/v1/validate_text_templates", EnsureViewer(service.ValidateTextTemplate))

	allRoutes := &AllRoutes{
		Logger:                opts.Logger,
		StatusFeed:            opts.StatusFeedURL,
		CustomLinks:           opts.CustomLinks,
		AddonTokens:           service.AddonTokens,
		PasswordPolicy:        opts.PasswordPolicy,
		PasswordPolicyMessage: opts.PasswordPolicyMessage,
		BasicRoute: BasicAuthRoute{
			Name:   "cloudhub",
			Login:  "/basic/login",
			Logout: "/basic/logout",
		},
		BasicLogoutLink:        "/basic/logout",
		LoginAuthType:          service.LoginAuthType,
		BasicPasswordResetType: service.BasicPasswordResetType,
		RetryPolicys:           service.RetryPolicy,
		AddonURLs:              service.AddonURLs,
		OSP:                    service.OSP,
	}

	getPrincipal := func(r *http.Request) oauth2.Principal {
		p, _ := HasAuthorizedToken(opts.Auth, r)
		return p
	}
	allRoutes.GetPrincipal = getPrincipal
	router.Handler("GET", "/cloudhub/v1/", allRoutes)

	var out http.Handler

	/* Authentication */
	if opts.UseAuth {
		// Encapsulate the router with OAuth2
		var auth http.Handler
		auth, allRoutes.AuthRoutes = AuthAPI(opts, router)
		allRoutes.LogoutLink = path.Join(opts.Basepath, "/oauth/logout")

		// Create middleware that redirects to the appropriate provider logout
		router.GET("/oauth/logout", logout("/", opts.Basepath, allRoutes.AuthRoutes))
		out = auth
	} else if opts.BasicAuth != nil {
		out = BasicAuthWrapper(router, opts.BasicAuth)
	} else {
		out = router
	}
	out = Logger(opts.Logger, FlushingHandler(out))

	return out
}

// AuthAPI adds the OAuth routes if auth is enabled.
func AuthAPI(opts MuxOpts, router cloudhub.Router) (http.Handler, AuthRoutes) {
	routes := AuthRoutes{}
	for _, pf := range opts.ProviderFuncs {
		pf(func(p oauth2.Provider, m oauth2.Mux) {
			urlName := url.PathEscape(strings.ToLower(p.Name()))

			loginPath := path.Join("/oauth", urlName, "login")
			logoutPath := path.Join("/oauth", urlName, "logout")
			callbackPath := path.Join("/oauth", urlName, "callback")

			router.Handler("GET", loginPath, m.Login())
			router.Handler("GET", logoutPath, m.Logout())
			router.Handler("GET", callbackPath, m.Callback())
			routes = append(routes, AuthRoute{
				Name:  p.Name(),
				Label: strings.Title(p.Name()),
				// AuthRoutes are content served to the page. When Basepath is set, it
				// says that all content served to the page will be prefixed with the
				// basepath. Since these routes are consumed by JS, it will need the
				// basepath set to traverse a proxy correctly
				Login:    path.Join(opts.Basepath, loginPath),
				Logout:   path.Join(opts.Basepath, logoutPath),
				Callback: path.Join(opts.Basepath, callbackPath),
			})
		})
	}

	rootPath := path.Join(opts.Basepath, "/cloudhub/v1")
	logoutPath := path.Join(opts.Basepath, "/oauth/logout")

	tokenMiddleware := AuthorizedToken(opts.Auth, opts.Logger, router)
	// Wrap the API with token validation middleware.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := path.Clean(r.URL.Path) // compare ignoring path garbage, trailing slashes, etc.
		if (strings.HasPrefix(cleanPath, rootPath) && len(cleanPath) > len(rootPath)) || cleanPath == logoutPath {
			tokenMiddleware.ServeHTTP(w, r)
			return
		}
		router.ServeHTTP(w, r)
	}), routes
}

// BasicAuthWrapper returns http handlers that wraps the supplied handler with HTTP Basic authentication
func BasicAuthWrapper(router cloudhub.Router, auth *basicAuth.BasicAuth) http.Handler {
	return auth.Wrap(func(response http.ResponseWriter, authRequest *basicAuth.AuthenticatedRequest) {
		router.ServeHTTP(response, &authRequest.Request)
	})
}

// BasicAuthAPI adds the Basic routes if auth is enabled.
// Copy session information to context when oauth is not used
// not using it now.
func BasicAuthAPI(opts MuxOpts, router cloudhub.Router) http.Handler {
	rootPath := path.Join(opts.Basepath, "/cloudhub/v1")

	tokenMiddleware := AuthorizedToken(opts.Auth, opts.Logger, router)
	// Wrap the API with token validation middleware.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := path.Clean(r.URL.Path) // compare ignoring path garbage, trailing slashes, etc.
		if strings.HasPrefix(cleanPath, rootPath) && len(cleanPath) > len(rootPath) {
			tokenMiddleware.ServeHTTP(w, r)
			return
		}
		router.ServeHTTP(w, r)
	})
}

func encodeJSON(w http.ResponseWriter, status int, v interface{}, logger cloudhub.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		unknownErrorWithMessage(w, err, logger)
	}
}

// Error writes an JSON message
func Error(w http.ResponseWriter, code int, msg string, logger cloudhub.Logger) {
	e := ErrorMessage{
		Code:    code,
		Message: msg,
	}
	b, err := json.Marshal(e)
	if err != nil {
		code = http.StatusInternalServerError
		b = []byte(`{"code": 500, "message":"server_error"}`)
	}

	logger.
		WithField("component", "server").
		WithField("http_status ", code).
		Error("Error message ", msg)
	w.Header().Set("Content-Type", JSONType)
	w.WriteHeader(code)
	_, _ = w.Write(b)
}

// ErrorBasic writes an JSON message by basic login
func ErrorBasic(w http.ResponseWriter, code int, msg string, retryCnt int32, lockedTime string, locked bool, logger cloudhub.Logger) {
	e := ErrorMessageBasic{
		Code:       code,
		Message:    msg,
		RetryCount: retryCnt,
		LockedTime: lockedTime,
		Locked:     locked,
	}
	b, err := json.Marshal(e)
	if err != nil {
		code = http.StatusInternalServerError
		b = []byte(`{"code": 500, "message":"server_error"}`)
	}

	logger.
		WithField("component", "server").
		WithField("http_status ", code).
		WithField("retry_count=", retryCnt).
		WithField("locked_time=", lockedTime).
		WithField("locked=", locked).
		Error("Error message ", msg)
	w.Header().Set("Content-Type", JSONType)
	w.WriteHeader(code)
	_, _ = w.Write(b)
}

func invalidData(w http.ResponseWriter, err error, logger cloudhub.Logger) {
	Error(w, http.StatusUnprocessableEntity, fmt.Sprintf("%v", err), logger)
}

func invalidJSON(w http.ResponseWriter, logger cloudhub.Logger) {
	Error(w, http.StatusBadRequest, "Unparsable JSON", logger)
}

func invalidXML(w http.ResponseWriter, logger cloudhub.Logger) {
	Error(w, http.StatusBadRequest, "Unparsable XML", logger)
}

func unknownErrorWithMessage(w http.ResponseWriter, err error, logger cloudhub.Logger) {
	Error(w, http.StatusInternalServerError, fmt.Sprintf("Unknown error: %v", err), logger)
}

func notFound(w http.ResponseWriter, id interface{}, logger cloudhub.Logger) {
	Error(w, http.StatusNotFound, fmt.Sprintf("ID %v not found", id), logger)
}

func paramID(key string, r *http.Request) (int, error) {
	ctx := r.Context()
	param := httprouter.GetParamFromContext(ctx, key)
	id, err := strconv.Atoi(param)
	if err != nil {
		return -1, fmt.Errorf("Error converting ID %s", param)
	}
	return id, nil
}

func paramInt64(key string, r *http.Request) (int64, error) {
	ctx := r.Context()
	param := httprouter.GetParamFromContext(ctx, key)
	v, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("Error converting parameter %s", param)
	}
	return v, nil
}

func paramStr(key string, r *http.Request) (string, error) {
	ctx := r.Context()
	param := httprouter.GetParamFromContext(ctx, key)
	return param, nil
}

// decodeRequest is a generic function to decode JSON request bodies
func decodeRequest[T any](r *http.Request) (T, context.Context, error) {
	var request T
	ctx := r.Context()

	// Decode the request body into the provided struct type
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return request, ctx, errors.New("invalid request data")
	}

	return request, ctx, nil
}

package kapacitor

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	client "github.com/influxdata/kapacitor/client/v1"
	cloudhub "github.com/snetsystems/cloudhub/backend"
)

// MockKapa is a mock implementation of KapaClient interface
type MockKapa struct {
	ResTask     client.Task
	ResTasks    []client.Task
	TaskError   error
	UpdateError error
	CreateError error
	ListError   error
	DeleteError error
	LastStatus  client.TaskStatus

	*client.CreateTaskOptions
	client.Link
	*client.TaskOptions
	*client.ListTasksOptions
	*client.UpdateTaskOptions
}

// CreateTask is a mock implementation of KapaClient.CreateTask
func (m *MockKapa) CreateTask(opt client.CreateTaskOptions) (client.Task, error) {
	m.CreateTaskOptions = &opt
	return m.ResTask, m.CreateError
}

func (m *MockKapa) Task(link client.Link, opt *client.TaskOptions) (client.Task, error) {
	m.Link = link
	m.TaskOptions = opt
	return m.ResTask, m.TaskError
}

func (m *MockKapa) ListTasks(opt *client.ListTasksOptions) ([]client.Task, error) {
	m.ListTasksOptions = opt
	return m.ResTasks, m.ListError
}

func (m *MockKapa) UpdateTask(link client.Link, opt client.UpdateTaskOptions) (client.Task, error) {
	m.Link = link
	m.LastStatus = opt.Status

	if m.UpdateTaskOptions == nil {
		m.UpdateTaskOptions = &opt
	}

	return m.ResTask, m.UpdateError
}

func (m *MockKapa) DeleteTask(link client.Link) error {
	m.Link = link
	return m.DeleteError
}

type MockID struct {
	ID string
}

func (m *MockID) Generate() (string, error) {
	return m.ID, nil
}

func TestClient_All(t *testing.T) {
	type fields struct {
		URL        string
		Username   string
		Password   string
		ID         cloudhub.ID
		Ticker     cloudhub.Ticker
		kapaClient func(url, username, password string, insecureSkipVerify bool) (KapaClient, error)
	}
	type args struct {
		ctx context.Context
	}
	kapa := &MockKapa{}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     map[string]*Task
		wantErr  bool
		resTask  client.Task
		resTasks []client.Task
		resError error

		createTaskOptions client.CreateTaskOptions
		link              client.Link
		taskOptions       *client.TaskOptions
		listTasksOptions  *client.ListTasksOptions
		updateTaskOptions client.UpdateTaskOptions
	}{
		{
			name: "return no tasks",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
			},
			listTasksOptions: &client.ListTasksOptions{},
			want:             map[string]*Task{},
		},
		{
			name: "return a non-reversible named task",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
			},
			listTasksOptions: &client.ListTasksOptions{},
			resTasks: []client.Task{
				{
					ID:     "howdy",
					Status: client.Enabled,
					TICKscript: `var whereFilter = lambda: TRUE

var name = 'rule 1'

var idVar = name + '-{{.Group}}'`,
				},
			},
			want: map[string]*Task{
				"howdy": {
					ID: "howdy",

					HrefOutput: "/kapacitor/v1/tasks/howdy/output",
					Rule: cloudhub.AlertRule{
						ID:   "howdy",
						Name: "rule 1",
						TICKScript: `var whereFilter = lambda: TRUE

var name = 'rule 1'

var idVar = name + '-{{.Group}}'`,
						Type:   "invalid",
						Status: "enabled",
						DBRPs:  []cloudhub.DBRP{},
					},
				},
			},
		},
		{
			name: "return a non-reversible task",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
			},
			listTasksOptions: &client.ListTasksOptions{},
			resTasks: []client.Task{
				{
					ID:     "howdy",
					Status: client.Enabled,
				},
			},
			want: map[string]*Task{
				"howdy": {
					ID: "howdy",

					HrefOutput: "/kapacitor/v1/tasks/howdy/output",
					Rule: cloudhub.AlertRule{
						ID:         "howdy",
						Name:       "howdy",
						TICKScript: "",
						Type:       "invalid",
						Status:     "enabled",
						DBRPs:      []cloudhub.DBRP{},
					},
					TICKScript: "",
				},
			},
		},
		{
			name: "return a reversible task",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
			},
			listTasksOptions: &client.ListTasksOptions{},
			resTasks: []client.Task{
				{
					ID:     "rule 1",
					Status: client.Enabled,
					Type:   client.StreamTask,
					DBRPs: []client.DBRP{
						{
							Database:        "_internal",
							RetentionPolicy: "autogen",
						},
					},
					TICKscript: `var db = '_internal'

var rp = 'monitor'

var measurement = 'cq'

var groupBy = []

var whereFilter = lambda: TRUE

var name = 'rule 1'

var idVar = name + '-{{.Group}}'

var message = ''

var idTag = 'alertID'

var levelTag = 'level'

var messageField = 'message'

var durationField = 'duration'

var outputDB = 'cloudhub'

var outputRP = 'autogen'

var outputMeasurement = 'alerts'

var triggerType = 'threshold'

var crit = 90000

var data = stream
    |from()
        .database(db)
        .retentionPolicy(rp)
        .measurement(measurement)
        .groupBy(groupBy)
        .where(whereFilter)
    |eval(lambda: "queryOk")
        .as('value')

var trigger = data
    |alert()
        .crit(lambda: "value" > crit)
        .stateChangesOnly()
        .message(message)
        .id(idVar)
        .idTag(idTag)
        .levelTag(levelTag)
        .messageField(messageField)
        .durationField(durationField)

trigger
    |eval(lambda: float("value"))
        .as('value')
        .keep()
    |influxDBOut()
        .create()
        .database(outputDB)
        .retentionPolicy(outputRP)
        .measurement(outputMeasurement)
        .tag('alertName', name)
        .tag('triggerType', triggerType)

trigger
    |httpOut('output')
`,
				},
			},
			want: map[string]*Task{
				"rule 1": {
					ID: "rule 1",

					HrefOutput: "/kapacitor/v1/tasks/rule 1/output",
					Rule: cloudhub.AlertRule{
						DBRPs: []cloudhub.DBRP{
							{

								DB: "_internal",
								RP: "autogen",
							},
						},
						Type:   "stream",
						Status: "enabled",
						ID:     "rule 1",
						Name:   "rule 1",
						TICKScript: `var db = '_internal'

var rp = 'monitor'

var measurement = 'cq'

var groupBy = []

var whereFilter = lambda: TRUE

var name = 'rule 1'

var idVar = name + '-{{.Group}}'

var message = ''

var idTag = 'alertID'

var levelTag = 'level'

var messageField = 'message'

var durationField = 'duration'

var outputDB = 'cloudhub'

var outputRP = 'autogen'

var outputMeasurement = 'alerts'

var triggerType = 'threshold'

var crit = 90000

var data = stream
    |from()
        .database(db)
        .retentionPolicy(rp)
        .measurement(measurement)
        .groupBy(groupBy)
        .where(whereFilter)
    |eval(lambda: "queryOk")
        .as('value')

var trigger = data
    |alert()
        .crit(lambda: "value" > crit)
        .stateChangesOnly()
        .message(message)
        .id(idVar)
        .idTag(idTag)
        .levelTag(levelTag)
        .messageField(messageField)
        .durationField(durationField)

trigger
    |eval(lambda: float("value"))
        .as('value')
        .keep()
    |influxDBOut()
        .create()
        .database(outputDB)
        .retentionPolicy(outputRP)
        .measurement(outputMeasurement)
        .tag('alertName', name)
        .tag('triggerType', triggerType)

trigger
    |httpOut('output')
`,
						Trigger: "threshold",
						TriggerValues: cloudhub.TriggerValues{
							Operator: "greater than",
							Value:    "90000",
						},
						AlertNodes: cloudhub.AlertNodes{
							IsStateChangesOnly: true,
						},
						Query: &cloudhub.QueryConfig{
							Database:        "_internal",
							RetentionPolicy: "monitor",
							Measurement:     "cq",
							Fields: []cloudhub.Field{
								{
									Value: "queryOk",
									Type:  "field",
								},
							},
							GroupBy: cloudhub.GroupBy{
								Tags: []string{},
							},
							AreTagsAccepted: false,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kapa.ResTask = tt.resTask
		kapa.ResTasks = tt.resTasks
		kapa.ListError = tt.resError
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				URL:        tt.fields.URL,
				Username:   tt.fields.Username,
				Password:   tt.fields.Password,
				ID:         tt.fields.ID,
				Ticker:     tt.fields.Ticker,
				kapaClient: tt.fields.kapaClient,
			}
			got, err := c.All(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !gocmp.Equal(got, tt.want) {
				t.Errorf("%q. Client.All() = -got/+want %s", tt.name, gocmp.Diff(got, tt.want))
			}
			if !reflect.DeepEqual(kapa.ListTasksOptions, tt.listTasksOptions) {
				t.Errorf("Client.All() = listTasksOptions  %v, want %v", kapa.ListTasksOptions, tt.listTasksOptions)
			}
			if !reflect.DeepEqual(kapa.TaskOptions, tt.taskOptions) {
				t.Errorf("Client.All() = taskOptions  %v, want %v", kapa.TaskOptions, tt.taskOptions)
			}
			if !reflect.DeepEqual(kapa.ListTasksOptions, tt.listTasksOptions) {
				t.Errorf("Client.All() = listTasksOptions  %v, want %v", kapa.ListTasksOptions, tt.listTasksOptions)
			}
			if !reflect.DeepEqual(kapa.Link, tt.link) {
				t.Errorf("Client.All() = Link  %v, want %v", kapa.Link, tt.link)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	type fields struct {
		URL        string
		Username   string
		Password   string
		ID         cloudhub.ID
		Ticker     cloudhub.Ticker
		kapaClient func(url, username, password string, insecureSkipVerify bool) (KapaClient, error)
	}
	type args struct {
		ctx context.Context
		id  string
	}
	kapa := &MockKapa{}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     *Task
		wantErr  bool
		resTask  client.Task
		resTasks []client.Task
		resError error

		createTaskOptions client.CreateTaskOptions
		link              client.Link
		taskOptions       *client.TaskOptions
		listTasksOptions  *client.ListTasksOptions
		updateTaskOptions client.UpdateTaskOptions
	}{
		{
			name: "return no task",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
			},
			args: args{
				id: "myid",
			},
			taskOptions: nil,
			wantErr:     true,
			resError:    fmt.Errorf("No such task"),
			link: client.Link{
				Href: "/kapacitor/v1/tasks/myid",
			},
		},
		{
			name: "return non-reversible task",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
			},
			args: args{
				id: "myid",
			},
			taskOptions: nil,
			resTask: client.Task{
				ID:     "myid",
				Status: client.Enabled,
				Type:   client.StreamTask,
				DBRPs: []client.DBRP{
					{
						Database:        "_internal",
						RetentionPolicy: "autogen",
					},
				},
			},
			want: &Task{
				ID:         "myid",
				HrefOutput: "/kapacitor/v1/tasks/myid/output",
				Rule: cloudhub.AlertRule{
					Type:   "stream",
					Status: "enabled",
					ID:     "myid",
					Name:   "myid",
					DBRPs: []cloudhub.DBRP{
						{
							DB: "_internal",
							RP: "autogen",
						},
					},
				},
			},
			link: client.Link{
				Href: "/kapacitor/v1/tasks/myid",
			},
		},
		{
			name: "return reversible task",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
			},
			args: args{
				id: "rule 1",
			},
			taskOptions: nil,
			resTask: client.Task{
				ID:     "rule 1",
				Status: client.Enabled,
				Type:   client.StreamTask,
				DBRPs: []client.DBRP{
					{
						Database:        "_internal",
						RetentionPolicy: "autogen",
					},
				},
				TICKscript: `var db = '_internal'

var rp = 'monitor'

var measurement = 'cq'

var groupBy = []

var whereFilter = lambda: TRUE

var name = 'rule 1'

var idVar = name + '-{{.Group}}'

var message = ''

var idTag = 'alertID'

var levelTag = 'level'

var messageField = 'message'

var durationField = 'duration'

var outputDB = 'cloudhub'

var outputRP = 'autogen'

var outputMeasurement = 'alerts'

var triggerType = 'threshold'

var crit = 90000

var data = stream
    |from()
        .database(db)
        .retentionPolicy(rp)
        .measurement(measurement)
        .groupBy(groupBy)
        .where(whereFilter)
    |eval(lambda: "queryOk")
        .as('value')

var trigger = data
    |alert()
        .crit(lambda: "value" > crit)
        .stateChangesOnly()
        .message(message)
        .id(idVar)
        .idTag(idTag)
        .levelTag(levelTag)
        .messageField(messageField)
        .durationField(durationField)

trigger
    |eval(lambda: float("value"))
        .as('value')
        .keep()
    |influxDBOut()
        .create()
        .database(outputDB)
        .retentionPolicy(outputRP)
        .measurement(outputMeasurement)
        .tag('alertName', name)
        .tag('triggerType', triggerType)

trigger
    |httpOut('output')
`,
			},
			want: &Task{
				ID:         "rule 1",
				HrefOutput: "/kapacitor/v1/tasks/rule 1/output",
				Rule: cloudhub.AlertRule{
					Type:   "stream",
					Status: "enabled",
					DBRPs: []cloudhub.DBRP{
						{

							DB: "_internal",
							RP: "autogen",
						},
					},
					ID:   "rule 1",
					Name: "rule 1",
					TICKScript: `var db = '_internal'

var rp = 'monitor'

var measurement = 'cq'

var groupBy = []

var whereFilter = lambda: TRUE

var name = 'rule 1'

var idVar = name + '-{{.Group}}'

var message = ''

var idTag = 'alertID'

var levelTag = 'level'

var messageField = 'message'

var durationField = 'duration'

var outputDB = 'cloudhub'

var outputRP = 'autogen'

var outputMeasurement = 'alerts'

var triggerType = 'threshold'

var crit = 90000

var data = stream
    |from()
        .database(db)
        .retentionPolicy(rp)
        .measurement(measurement)
        .groupBy(groupBy)
        .where(whereFilter)
    |eval(lambda: "queryOk")
        .as('value')

var trigger = data
    |alert()
        .crit(lambda: "value" > crit)
        .stateChangesOnly()
        .message(message)
        .id(idVar)
        .idTag(idTag)
        .levelTag(levelTag)
        .messageField(messageField)
        .durationField(durationField)

trigger
    |eval(lambda: float("value"))
        .as('value')
        .keep()
    |influxDBOut()
        .create()
        .database(outputDB)
        .retentionPolicy(outputRP)
        .measurement(outputMeasurement)
        .tag('alertName', name)
        .tag('triggerType', triggerType)

trigger
    |httpOut('output')
`,
					Trigger: "threshold",
					TriggerValues: cloudhub.TriggerValues{
						Operator: "greater than",
						Value:    "90000",
					},
					AlertNodes: cloudhub.AlertNodes{
						IsStateChangesOnly: true,
					},
					Query: &cloudhub.QueryConfig{
						Database:        "_internal",
						RetentionPolicy: "monitor",
						Measurement:     "cq",
						Fields: []cloudhub.Field{
							{
								Value: "queryOk",
								Type:  "field",
							},
						},
						GroupBy: cloudhub.GroupBy{
							Tags: []string{},
						},
						AreTagsAccepted: false,
					},
				},
			},
			link: client.Link{
				Href: "/kapacitor/v1/tasks/rule 1",
			},
		},
	}
	for _, tt := range tests {
		kapa.ResTask = tt.resTask
		kapa.ResTasks = tt.resTasks
		kapa.TaskError = tt.resError
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				URL:        tt.fields.URL,
				Username:   tt.fields.Username,
				Password:   tt.fields.Password,
				ID:         tt.fields.ID,
				Ticker:     tt.fields.Ticker,
				kapaClient: tt.fields.kapaClient,
			}
			got, err := c.Get(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !gocmp.Equal(got, tt.want) {
				t.Errorf("%q. Client.All() = -got/+want %s", tt.name, gocmp.Diff(got, tt.want))
			}
			if !reflect.DeepEqual(kapa.ListTasksOptions, tt.listTasksOptions) {
				t.Errorf("Client.Get() = listTasksOptions  %v, want %v", kapa.ListTasksOptions, tt.listTasksOptions)
			}
			if !reflect.DeepEqual(kapa.TaskOptions, tt.taskOptions) {
				t.Errorf("Client.Get() = taskOptions  %v, want %v", kapa.TaskOptions, tt.taskOptions)
			}
			if !reflect.DeepEqual(kapa.ListTasksOptions, tt.listTasksOptions) {
				t.Errorf("Client.Get() = listTasksOptions  %v, want %v", kapa.ListTasksOptions, tt.listTasksOptions)
			}
			if !reflect.DeepEqual(kapa.Link, tt.link) {
				t.Errorf("Client.Get() = Link  %v, want %v", kapa.Link, tt.link)
			}
		})
	}
}

func TestClient_updateStatus(t *testing.T) {
	type fields struct {
		URL        string
		Username   string
		Password   string
		ID         cloudhub.ID
		Ticker     cloudhub.Ticker
		kapaClient func(url, username, password string, insecureSkipVerify bool) (KapaClient, error)
	}
	type args struct {
		ctx    context.Context
		href   string
		status client.TaskStatus
	}
	kapa := &MockKapa{}
	tests := []struct {
		name              string
		fields            fields
		args              args
		resTask           client.Task
		want              *Task
		resError          error
		wantErr           bool
		updateTaskOptions *client.UpdateTaskOptions
	}{
		{
			name: "disable alert rule",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:    context.Background(),
				href:   "/kapacitor/v1/tasks/howdy",
				status: client.Disabled,
			},
			resTask: client.Task{
				ID:     "howdy",
				Status: client.Disabled,
				Type:   client.StreamTask,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
				Link: client.Link{
					Href: "/kapacitor/v1/tasks/howdy",
				},
			},
			updateTaskOptions: &client.UpdateTaskOptions{
				TICKscript: "",
				Status:     client.Disabled,
			},
			want: &Task{
				ID:         "howdy",
				Href:       "/kapacitor/v1/tasks/howdy",
				HrefOutput: "/kapacitor/v1/tasks/howdy/output",
				Rule: cloudhub.AlertRule{
					ID:   "howdy",
					Name: "howdy",
					Type: "stream",
					DBRPs: []cloudhub.DBRP{
						{

							DB: "db",
							RP: "rp",
						},
					},
					Status: "disabled",
				},
			},
		},
		{
			name: "fail to enable alert rule",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:    context.Background(),
				href:   "/kapacitor/v1/tasks/howdy",
				status: client.Enabled,
			},
			updateTaskOptions: &client.UpdateTaskOptions{
				TICKscript: "",
				Status:     client.Enabled,
			},
			resError: fmt.Errorf("error"),
			wantErr:  true,
		},
		{
			name: "enable alert rule",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:    context.Background(),
				href:   "/kapacitor/v1/tasks/howdy",
				status: client.Enabled,
			},
			resTask: client.Task{
				ID:   "howdy",
				Type: client.StreamTask,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
				Status: client.Enabled,
				Link: client.Link{
					Href: "/kapacitor/v1/tasks/howdy",
				},
			},
			updateTaskOptions: &client.UpdateTaskOptions{
				TICKscript: "",
				Status:     client.Enabled,
			},
			want: &Task{
				ID:         "howdy",
				Href:       "/kapacitor/v1/tasks/howdy",
				HrefOutput: "/kapacitor/v1/tasks/howdy/output",
				Rule: cloudhub.AlertRule{
					ID:   "howdy",
					Name: "howdy",
					Type: "stream",
					DBRPs: []cloudhub.DBRP{
						{

							DB: "db",
							RP: "rp",
						},
					},
					Status: "enabled",
				},
			},
		},
	}
	for _, tt := range tests {
		kapa.ResTask = tt.resTask
		kapa.UpdateError = tt.resError
		kapa.UpdateTaskOptions = nil
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				URL:        tt.fields.URL,
				Username:   tt.fields.Username,
				Password:   tt.fields.Password,
				ID:         tt.fields.ID,
				Ticker:     tt.fields.Ticker,
				kapaClient: tt.fields.kapaClient,
			}
			got, err := c.updateStatus(tt.args.ctx, tt.args.href, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.updateStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !gocmp.Equal(got, tt.want) {
				t.Errorf("%q. Client.updateStatus() = -got/+want %s", tt.name, gocmp.Diff(got, tt.want))
			}
			if !reflect.DeepEqual(kapa.UpdateTaskOptions, tt.updateTaskOptions) {
				t.Errorf("Client.updateStatus() = %v, want %v", kapa.UpdateTaskOptions, tt.updateTaskOptions)
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	type fields struct {
		URL        string
		Username   string
		Password   string
		ID         cloudhub.ID
		Ticker     cloudhub.Ticker
		kapaClient func(url, username, password string, insecureSkipVerify bool) (KapaClient, error)
	}
	type args struct {
		ctx  context.Context
		href string
		rule cloudhub.AlertRule
	}
	kapa := &MockKapa{}
	tests := []struct {
		name              string
		fields            fields
		args              args
		resTask           client.Task
		want              *Task
		resError          error
		wantErr           bool
		updateTaskOptions *client.UpdateTaskOptions
		wantStatus        client.TaskStatus
	}{
		{
			name: "update alert rule error",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:  context.Background(),
				href: "/kapacitor/v1/tasks/howdy",
				rule: cloudhub.AlertRule{
					ID: "howdy",
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
			},
			resError: fmt.Errorf("error"),
			updateTaskOptions: &client.UpdateTaskOptions{
				TICKscript: "",
				Type:       client.StreamTask,
				Status:     client.Disabled,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
			},
			wantErr:    true,
			wantStatus: client.Disabled,
		},
		{
			name: "update alert rule",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:  context.Background(),
				href: "/kapacitor/v1/tasks/howdy",
				rule: cloudhub.AlertRule{
					ID:   "howdy",
					Name: "myname",
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
						Measurement:     "meas",
						Fields: []cloudhub.Field{
							{
								Type:  "field",
								Value: "usage_user",
							},
						},
					},
					Trigger: "threshold",
					TriggerValues: cloudhub.TriggerValues{
						Operator: greaterThan,
					},
				},
			},
			resTask: client.Task{
				ID:   "howdy",
				Type: client.StreamTask,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
				Status: client.Enabled,
				Link: client.Link{
					Href: "/kapacitor/v1/tasks/howdy",
				},
			},
			updateTaskOptions: &client.UpdateTaskOptions{
				TICKscript: "",
				Type:       client.StreamTask,
				Status:     client.Disabled,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
			},
			want: &Task{
				ID:         "howdy",
				Href:       "/kapacitor/v1/tasks/howdy",
				HrefOutput: "/kapacitor/v1/tasks/howdy/output",
				Rule: cloudhub.AlertRule{
					DBRPs: []cloudhub.DBRP{
						{

							DB: "db",
							RP: "rp",
						},
					},
					Status: "enabled",
					Type:   "stream",
					ID:     "howdy",
					Name:   "howdy",
				},
			},
			wantStatus: client.Enabled,
		},
		{
			name: "stays disabled when already disabled",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:  context.Background(),
				href: "/kapacitor/v1/tasks/howdy",
				rule: cloudhub.AlertRule{
					ID:   "howdy",
					Name: "myname",
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
						Measurement:     "meas",
						Fields: []cloudhub.Field{
							{
								Type:  "field",
								Value: "usage_user",
							},
						},
					},
					Trigger: "threshold",
					TriggerValues: cloudhub.TriggerValues{
						Operator: greaterThan,
					},
				},
			},
			resTask: client.Task{
				ID:   "howdy",
				Type: client.StreamTask,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
				Status: client.Disabled,
				Link: client.Link{
					Href: "/kapacitor/v1/tasks/howdy",
				},
			},
			updateTaskOptions: &client.UpdateTaskOptions{
				TICKscript: "",
				Type:       client.StreamTask,
				Status:     client.Disabled,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
			},
			want: &Task{
				ID:         "howdy",
				Href:       "/kapacitor/v1/tasks/howdy",
				HrefOutput: "/kapacitor/v1/tasks/howdy/output",
				Rule: cloudhub.AlertRule{
					ID:   "howdy",
					Name: "howdy",
					DBRPs: []cloudhub.DBRP{
						{

							DB: "db",
							RP: "rp",
						},
					},
					Status: "disabled",
					Type:   "stream",
				},
			},
			wantStatus: client.Disabled,
		},
		{
			name:    "error because relative cannot have inside range",
			wantErr: true,
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:  context.Background(),
				href: "/kapacitor/v1/tasks/error",
				rule: cloudhub.AlertRule{
					ID: "error",
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
						Fields: []cloudhub.Field{
							{
								Value: "usage_user",
								Type:  "field",
							},
						},
					},
					Trigger: Relative,
					TriggerValues: cloudhub.TriggerValues{
						Operator: insideRange,
					},
				},
			},
		},
		{
			name:    "error because rule has an unknown trigger mechanism",
			wantErr: true,
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:  context.Background(),
				href: "/kapacitor/v1/tasks/error",
				rule: cloudhub.AlertRule{
					ID: "error",
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
			},
		},
		{
			name:    "error because query has no fields",
			wantErr: true,
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:  context.Background(),
				href: "/kapacitor/v1/tasks/error",
				rule: cloudhub.AlertRule{
					ID:      "error",
					Trigger: Threshold,
					TriggerValues: cloudhub.TriggerValues{
						Period: "1d",
					},
					Name: "myname",
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
						Measurement:     "meas",
					},
				},
			},
		},
		{
			name:    "error because alert has no name",
			wantErr: true,
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:  context.Background(),
				href: "/kapacitor/v1/tasks/error",
				rule: cloudhub.AlertRule{
					ID:      "error",
					Trigger: Deadman,
					TriggerValues: cloudhub.TriggerValues{
						Period: "1d",
					},
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
						Measurement:     "meas",
					},
				},
			},
		},
		{
			name:    "error because alert period cannot be an empty string in deadman alert",
			wantErr: true,
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
			},
			args: args{
				ctx:  context.Background(),
				href: "/kapacitor/v1/tasks/error",
				rule: cloudhub.AlertRule{
					ID:      "error",
					Name:    "myname",
					Trigger: Deadman,
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
						Measurement:     "meas",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		kapa.ResTask = tt.resTask
		kapa.UpdateError = tt.resError
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				URL:        tt.fields.URL,
				Username:   tt.fields.Username,
				Password:   tt.fields.Password,
				ID:         tt.fields.ID,
				Ticker:     tt.fields.Ticker,
				kapaClient: tt.fields.kapaClient,
			}
			got, err := c.Update(tt.args.ctx, tt.args.href, tt.args.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !gocmp.Equal(got, tt.want) {
				t.Errorf("%q. Client.Update() = -got/+want %s", tt.name, gocmp.Diff(got, tt.want))
			}
			var cloudhubOptions = gocmp.Options{
				cmpopts.IgnoreFields(client.UpdateTaskOptions{}, "TICKscript"),
			}
			if !gocmp.Equal(kapa.UpdateTaskOptions, tt.updateTaskOptions, cloudhubOptions...) {
				t.Errorf("Client.Update() = %s", gocmp.Diff(got, tt.updateTaskOptions, cloudhubOptions...))
			}
			if tt.wantStatus != kapa.LastStatus {
				t.Errorf("Client.Update() = %v, want %v", kapa.LastStatus, tt.wantStatus)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	type fields struct {
		URL        string
		Username   string
		Password   string
		ID         cloudhub.ID
		Ticker     cloudhub.Ticker
		kapaClient func(url, username, password string, insecureSkipVerify bool) (KapaClient, error)
	}
	type args struct {
		ctx  context.Context
		rule cloudhub.AlertRule
	}
	kapa := &MockKapa{}
	tests := []struct {
		name              string
		fields            fields
		args              args
		resTask           client.Task
		want              *Task
		resError          error
		wantErr           bool
		createTaskOptions *client.CreateTaskOptions
	}{
		{
			name: "create alert rule with tags",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
				ID: &MockID{
					ID: "howdy",
				},
			},
			args: args{
				ctx: context.Background(),
				rule: cloudhub.AlertRule{
					ID:   "",
					Name: "cloudhub-v1-howdy",
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
						Measurement:     "meas",
						GroupBy: cloudhub.GroupBy{
							Tags: []string{
								"tag1",
								"tag2",
							},
						},
					},
					Trigger: Deadman,
					TriggerValues: cloudhub.TriggerValues{
						Period: "1d",
					},
				},
			},
			resTask: client.Task{
				ID:     "cloudhub-v1-howdy",
				Status: client.Enabled,
				Type:   client.StreamTask,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
				Link: client.Link{
					Href: "/kapacitor/v1/tasks/cloudhub-v1-howdy",
				},
			},
			createTaskOptions: &client.CreateTaskOptions{
				TICKscript: `var db = 'db'

var rp = 'rp'

var measurement = 'meas'

var groupBy = ['tag1', 'tag2']

var whereFilter = lambda: TRUE

var period = 1d

var name = 'cloudhub-v1-howdy'

var idVar = name + '-{{.Group}}'

var message = ''

var idTag = 'alertID'

var levelTag = 'level'

var messageField = 'message'

var durationField = 'duration'

var outputDB = 'db'

var outputRP = 'autogen'

var outputMeasurement = 'cloudhub_alerts'

var triggerType = 'deadman'

var threshold = 0.0

var data = stream
    |from()
        .database(db)
        .retentionPolicy(rp)
        .measurement(measurement)
        .groupBy(groupBy)
        .where(whereFilter)

var trigger = data
    |deadman(threshold, period)
        .stateChangesOnly()
        .message(message)
        .id(idVar)
        .idTag(idTag)
        .levelTag(levelTag)
        .messageField(messageField)
        .durationField(durationField)

trigger
    |eval(lambda: "emitted")
        .as('value')
        .keep('value', messageField, durationField)
    |eval(lambda: float("value"))
        .as('value')
        .keep()
    |influxDBOut()
        .create()
        .database(outputDB)
        .retentionPolicy(outputRP)
        .measurement(outputMeasurement)
        .tag('alertName', name)
        .tag('triggerType', triggerType)

trigger
    |httpOut('output')
`,
				ID:     "cloudhub-v1-howdy",
				Type:   client.StreamTask,
				Status: client.Enabled,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
			},
			want: &Task{
				ID:         "cloudhub-v1-howdy",
				Href:       "/kapacitor/v1/tasks/cloudhub-v1-howdy",
				HrefOutput: "/kapacitor/v1/tasks/cloudhub-v1-howdy/output",
				Rule: cloudhub.AlertRule{
					Type: "stream",
					DBRPs: []cloudhub.DBRP{
						{
							DB: "db",
							RP: "rp",
						},
					},
					Status: "enabled",
					ID:     "cloudhub-v1-howdy",
					Name:   "cloudhub-v1-howdy",
				},
			},
		},
		{
			name: "create alert rule with no tags",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
				ID: &MockID{
					ID: "howdy",
				},
			},
			args: args{
				ctx: context.Background(),
				rule: cloudhub.AlertRule{
					ID:   "",
					Name: "cloudhub-v1-howdy",
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
						Measurement:     "meas",
					},
					Trigger: Deadman,
					TriggerValues: cloudhub.TriggerValues{
						Period: "1d",
					},
				},
			},
			resTask: client.Task{
				ID:     "cloudhub-v1-howdy",
				Status: client.Enabled,
				Type:   client.StreamTask,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
				Link: client.Link{
					Href: "/kapacitor/v1/tasks/cloudhub-v1-howdy",
				},
			},
			createTaskOptions: &client.CreateTaskOptions{
				TICKscript: `var db = 'db'

var rp = 'rp'

var measurement = 'meas'

var groupBy = []

var whereFilter = lambda: TRUE

var period = 1d

var name = 'cloudhub-v1-howdy'

var idVar = name

var message = ''

var idTag = 'alertID'

var levelTag = 'level'

var messageField = 'message'

var durationField = 'duration'

var outputDB = 'db'

var outputRP = 'autogen'

var outputMeasurement = 'cloudhub_alerts'

var triggerType = 'deadman'

var threshold = 0.0

var data = stream
    |from()
        .database(db)
        .retentionPolicy(rp)
        .measurement(measurement)
        .groupBy(groupBy)
        .where(whereFilter)

var trigger = data
    |deadman(threshold, period)
        .stateChangesOnly()
        .message(message)
        .id(idVar)
        .idTag(idTag)
        .levelTag(levelTag)
        .messageField(messageField)
        .durationField(durationField)

trigger
    |eval(lambda: "emitted")
        .as('value')
        .keep('value', messageField, durationField)
    |eval(lambda: float("value"))
        .as('value')
        .keep()
    |influxDBOut()
        .create()
        .database(outputDB)
        .retentionPolicy(outputRP)
        .measurement(outputMeasurement)
        .tag('alertName', name)
        .tag('triggerType', triggerType)

trigger
    |httpOut('output')
`,
				ID:     "cloudhub-v1-howdy",
				Type:   client.StreamTask,
				Status: client.Enabled,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
			},
			want: &Task{
				ID:         "cloudhub-v1-howdy",
				Href:       "/kapacitor/v1/tasks/cloudhub-v1-howdy",
				HrefOutput: "/kapacitor/v1/tasks/cloudhub-v1-howdy/output",
				Rule: cloudhub.AlertRule{
					Type: "stream",
					DBRPs: []cloudhub.DBRP{
						{
							DB: "db",
							RP: "rp",
						},
					},
					Status: "enabled",
					ID:     "cloudhub-v1-howdy",
					Name:   "cloudhub-v1-howdy",
				},
			},
		},
		{
			name: "create alert rule error",
			fields: fields{
				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
					return kapa, nil
				},
				Ticker: &Alert{},
				ID: &MockID{
					ID: "howdy",
				},
			},
			args: args{
				ctx: context.Background(),
				rule: cloudhub.AlertRule{
					ID: "howdy",
					Query: &cloudhub.QueryConfig{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
			},
			resError: fmt.Errorf("error"),
			createTaskOptions: &client.CreateTaskOptions{
				ID:     "cloudhub-v1-howdy",
				Type:   client.StreamTask,
				Status: client.Enabled,
				DBRPs: []client.DBRP{
					{
						Database:        "db",
						RetentionPolicy: "rp",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		kapa.ResTask = tt.resTask
		kapa.CreateError = tt.resError
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				URL:        tt.fields.URL,
				Username:   tt.fields.Username,
				Password:   tt.fields.Password,
				ID:         tt.fields.ID,
				Ticker:     tt.fields.Ticker,
				kapaClient: tt.fields.kapaClient,
			}
			got, err := c.Create(tt.args.ctx, tt.args.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("%q. Client.Create() = -got/+want %s", tt.name, cmp.Diff(got, tt.want))
			}
			if !reflect.DeepEqual(kapa.CreateTaskOptions, tt.createTaskOptions) {
				t.Errorf("Client.Create() =  %v, want %v", kapa.CreateTaskOptions, tt.createTaskOptions)
			}
		})
	}
}
func (m *MockKapa) EnableTask(id string) error {
	return nil
}

func (m *MockKapa) DisableTask(id string) error {
	return nil
}

func removeWhitespace(s string) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(s), ""), "\n", "")
}

// func TestClient_CustomCreate(t *testing.T) {
// 	type fields struct {
// 		URL        string
// 		Username   string
// 		Password   string
// 		ID         cloudhub.ID
// 		Ticker     cloudhub.Ticker
// 		kapaClient func(url, username, password string, insecureSkipVerify bool) (KapaClient, error)
// 	}
// 	type args struct {
// 		ctx  context.Context
// 		rule cloudhub.AutoGeneratePredictionRule
// 	}
// 	kapa := &MockKapa{}

// 	alertScript := `
// var name = 'Anomaly Prediction'
// var db = 'logstash'

// var predict_mode = 'Ensemble'
// var ensemble_condition = 'and'
// var message = 'Kapacitor-1 Detected: [{{.Level}}] [{{.ID}}] {{ index .Tags "value" }}'
// var triggerType = 'anomaly_predict'

// var tffUsed = stream
//     |from()
//         .database(db)
//         .retentionPolicy('autogen')
//         .measurement('snmp_nx_if')
//         .where(lambda: ("ifDescr" =~ /Ethernet/))
//         .groupBy(['agent_host', 'sys_name'])

// var cpuUsed = stream
//     |from()
//         .database(db)
//         .retentionPolicy('autogen')
//         .measurement('snmp_nx')
//         .groupBy(['agent_host', 'sys_name'])
//     |eval(lambda: "cpu1min")
//         .as('cpu_used')

// var inOctets = tffUsed
//     |sum('ifHCInOctets')
//         .as('total_ifHCInOctets')
//     |derivative('total_ifHCInOctets')
//         .unit(1s)
//         .nonNegative()

// var outOctets = tffUsed
//     |sum('ifHCOutOctets')
//         .as('total_ifHCOutOctets')
//     |derivative('total_ifHCOutOctets')
//         .unit(1s)
//         .nonNegative()

// var joined = inOctets
//     |join(outOctets)
//         .as('in', 'out')
//         .tolerance(1s)

// var tffVolume = joined
//     |eval(lambda: "in.total_ifHCInOctets" + "out.total_ifHCOutOctets")
//         .as('total_traffic')

// var predictData = cpuUsed
//     |join(tffVolume)
//         .as('cpu', 'traffic')
//         .tolerance(1s)
//     |eval(lambda: "cpu.cpu_used", lambda: "traffic.total_traffic")
//         .as('cpu_used', 'tff_volume')
//     |window()
//         .periodCount(5)
//         .everyCount(1)

// var data = predictData
//     @predict_batch()
//         .predict_mode(predict_mode)
//         .ensemble_condition(ensemble_condition)
//     |log()

// var trigger = data
//     |alert()
//         .crit(lambda: "predict_status" == 'false')
//         .message(message)
//         .id(name)
//         .idTag('alertID')
//         .levelTag('level')
//         .messageField('message')
//         .tcp('myaddress:22')

// trigger
//     |eval(lambda: float("value"))
//         .as('value')
//         .keep()
//     |influxDBOut()
//         .create()
//         .database('cloudhub')
//         .retentionPolicy('autogen')
//         .measurement('alerts')
//         .tag('alertName', name)
//         .tag('triggerType', triggerType)
// `

// 	tests := []struct {
// 		name              string
// 		fields            fields
// 		args              args
// 		resTask           client.Task
// 		want              *Task
// 		resError          error
// 		wantErr           bool
// 		createTaskOptions *client.CreateTaskOptions
// 	}{
// 		{
// 			name: "create custom alert rule",
// 			fields: fields{
// 				kapaClient: func(url, username, password string, insecureSkipVerify bool) (KapaClient, error) {
// 					return kapa, nil
// 				},
// 				Ticker: &Alert{},
// 				ID: &MockID{
// 					ID: "howdy",
// 				},
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				rule: cloudhub.AutoGeneratePredictionRule{
// 					AlertRule: cloudhub.AlertRule{
// 						ID:   "",
// 						Name: "Anomaly Prediction",
// 						Query: &cloudhub.QueryConfig{
// 							Database:        "logstash",
// 							RetentionPolicy: "autogen",
// 							Measurement:     "snmp_nx_if",
// 							GroupBy: cloudhub.GroupBy{
// 								Tags: []string{
// 									"agent_host",
// 									"sys_name",
// 								},
// 							},
// 						},
// 						AlertNodes: cloudhub.AlertNodes{
// 							TCPs: []*cloudhub.TCP{
// 								{
// 									Address: "myaddress:22",
// 								},
// 							},
// 						},
// 						Message: "Kapacitor-1 Detected: [{{.Level}}] [{{.ID}}] {{ index .Tags \"value\" }}",
// 					},
// 				},
// 			},
// 			resTask: client.Task{
// 				ID:     "cloudhub-v1-howdy",
// 				Status: client.Enabled,
// 				Type:   client.StreamTask,
// 				DBRPs: []client.DBRP{
// 					{
// 						Database:        "logstash",
// 						RetentionPolicy: "autogen",
// 					},
// 				},
// 				Link: client.Link{
// 					Href: "/kapacitor/v1/tasks/cloudhub-v1-howdy",
// 				},
// 			},
// 			createTaskOptions: &client.CreateTaskOptions{
// 				TICKscript: removeWhitespace(alertScript),
// 				ID:         "cloudhub-v1-howdy",
// 				Type:       client.StreamTask,
// 				Status:     client.Enabled,
// 				DBRPs: []client.DBRP{
// 					{
// 						Database:        "logstash",
// 						RetentionPolicy: "autogen",
// 					},
// 				},
// 			},
// 			want: &Task{
// 				ID:         "cloudhub-v1-howdy",
// 				Href:       "/kapacitor/v1/tasks/cloudhub-v1-howdy",
// 				HrefOutput: "/kapacitor/v1/tasks/cloudhub-v1-howdy/output",
// 				Rule: cloudhub.AlertRule{
// 					Type: "stream",
// 					DBRPs: []cloudhub.DBRP{
// 						{
// 							DB: "logstash",
// 							RP: "autogen",
// 						},
// 					},
// 					Status: "enabled",
// 					ID:     "cloudhub-v1-howdy",
// 					Name:   "cloudhub-v1-howdy",
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		kapa.ResTask = tt.resTask
// 		kapa.CreateError = tt.resError
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := &Client{
// 				URL:        tt.fields.URL,
// 				Username:   tt.fields.Username,
// 				Password:   tt.fields.Password,
// 				ID:         tt.fields.ID,
// 				Ticker:     tt.fields.Ticker,
// 				kapaClient: tt.fields.kapaClient,
// 			}
// 			got, err := c.AutoGenerateCreate(tt.args.ctx, tt.args.rule)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("AutoGenerateCreate Client.Create() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.wantErr {
// 				return
// 			}
// 			if !cmp.Equal(got, tt.want) {
// 				t.Errorf("%q. AutoGenerateCreate Client.Create() = -got/+want %s", tt.name, cmp.Diff(got, tt.want))
// 			}
// 			if removeWhitespace(kapa.CreateTaskOptions.TICKscript) != removeWhitespace(tt.createTaskOptions.TICKscript) {

// 				t.Errorf("TICKscript Diff = %s", cmp.Diff(removeWhitespace(kapa.CreateTaskOptions.TICKscript), removeWhitespace(tt.createTaskOptions.TICKscript)))
// 			}
// 		})
// 	}
// }

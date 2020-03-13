package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.uber.org/cadence"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	initializeWorkflowsActivities()
	cadenceClient := createAndStartWorker()

	r := mux.NewRouter()
	r.HandleFunc("/start-workflow", func(writer http.ResponseWriter, request *http.Request) {

		type payload struct {
			Name string `json:"name"`
		}
		var p payload
		err := json.NewDecoder(request.Body).Decode(&p)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = cadenceClient.StartWorkflow(request.Context(), client.StartWorkflowOptions{
			TaskList: "test-task-list",
			ExecutionStartToCloseTimeout: time.Minute * 10,
			RetryPolicy: &cadence.RetryPolicy{
				InitialInterval: time.Second * 1,
				MaximumAttempts: 10,
			},
		}, TestWorkflow2, p.Name)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	_ = http.ListenAndServe(":1234", r)
}

func createAndStartWorker() client.Client {
	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName("cadence-service"))
	if err != nil {
		panic("Failed to setup tchannel")
	}
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: "cadence-service",
		Outbounds: yarpc.Outbounds{
			"cadence-frontend": {Unary: ch.NewSingleOutbound("127.0.0.1:7933")},
		},
	})
	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher")
	}

	cadenceServiceClient := workflowserviceclient.New(dispatcher.ClientConfig("cadence-frontend"))
	cadenceWorker := worker.New(cadenceServiceClient, "test-domain", "test-task-list", worker.Options{})
	err = cadenceWorker.Start()
	if err != nil {
		fmt.Printf("Error starting worker: %+v\n", err)
	}
	return client.NewClient(cadenceServiceClient, "test-domain", &client.Options{})
}

func initializeWorkflowsActivities() {
	workflow.Register(TestWorkflow2)
	activity.Register(TestActivity3)
	activity.Register(TestActivity4)
}

func TestWorkflow2(ctx workflow.Context, input string) (string, error) {
	workflow.GetLogger(ctx).Info("Started")
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskList: "test-task-list",
		ScheduleToStartTimeout: time.Second * 10,
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval: time.Second * 1,
			MaximumAttempts: 1,
		},
	})

	var result string
	err := workflow.ExecuteActivity(ctx, TestActivity3, fmt.Sprintf("From workflow: %s", input)).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, TestActivity4, fmt.Sprintf("From workflow: %s", result)).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	workflow.GetLogger(ctx).Info("Done", zap.String("result", result))
	return result, nil
}

func TestActivity3(input string) (string, error) {
	time.Sleep(time.Minute * 3)
	return fmt.Sprintf("From activity: %s", input), nil
}
func TestActivity4(input string) (string, error) {
	time.Sleep(time.Minute * 3)
	return fmt.Sprintf("From activity 2: %s", input), nil
}

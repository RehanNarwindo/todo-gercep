package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "monitoring/proto"
)

var (
    // Metrics untuk REST API (dari repo lain)
    restLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "rest_api_latency_seconds",
            Help: "Latency REST API calls ke user-service eksternal",
        },
        []string{"endpoint", "method", "status"},
    )

    restRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rest_api_requests_total",
            Help: "Total REST API requests ke user-service eksternal",
        },
        []string{"endpoint", "method", "status"},
    )

    // Metrics untuk gRPC internal
    grpcLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "grpc_internal_latency_seconds",
            Help: "Latency gRPC calls ke internal server",
        },
        []string{"method", "status"},
    )

    grpcRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_internal_requests_total",
            Help: "Total gRPC calls ke internal server",
        },
        []string{"method", "status"},
    )

    taskCount = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "todo_tasks_total",
            Help: "Total number of tasks in todo service",
        },
    )

    serviceHealth = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "service_health_status",
            Help: "Service health status (1=UP, 0=DOWN)",
        },
        []string{"service", "type"},
    )
)

func main() {
    go monitorExternalRESTAPI()
    go monitorInternalGRPC()

    http.Handle("/metrics", promhttp.Handler())
    log.Println("✅ Monitoring service berjalan di :3002")
    log.Println("📊 Metrics: http://localhost:3002/metrics")
    log.Fatal(http.ListenAndServe(":3002", nil))
}

func monitorExternalRESTAPI() {
    // ✅ Tambahkan semua endpoint user-service
    endpoints := []struct {
        path   string
        method string
    }{
        {"/public", "GET"},
        {"/users/all", "GET"},
        {"/users/profile", "GET"},
        {"/users/profile/1", "GET"},
        {"/users/1", "PUT"},
        {"/users/1", "DELETE"},
    }

    for {
        for _, ep := range endpoints {
            url := "http://localhost:3001" + ep.path
            
            start := time.Now()
            
            // Buat request sesuai method
            var resp *http.Response
            var err error
            
            switch ep.method {
            case "GET":
                resp, err = http.Get(url)
            case "PUT":
                // Buat request PUT (tanpa body untuk monitoring)
                req, _ := http.NewRequest("PUT", url, nil)
                client := &http.Client{}
                resp, err = client.Do(req)
            case "DELETE":
                req, _ := http.NewRequest("DELETE", url, nil)
                client := &http.Client{}
                resp, err = client.Do(req)
            default:
                resp, err = http.Get(url)
            }
            
            latency := time.Since(start).Seconds()

            status := "error"
            if err != nil {
                log.Printf("❌ %s %s error: %v", ep.method, ep.path, err)
                serviceHealth.WithLabelValues("user-service-rest", "rest").Set(0)
            } else {
                defer resp.Body.Close()
                status = resp.Status
                serviceHealth.WithLabelValues("user-service-rest", "rest").Set(1)
                log.Printf("✅ %s %s response: %s", ep.method, ep.path, resp.Status)
            }

            // Record metrics dengan endpoint, method, dan status
            restLatency.WithLabelValues(ep.path, ep.method, status).Observe(latency)
            restRequestsTotal.WithLabelValues(ep.path, ep.method, status).Inc()
        }
        time.Sleep(30 * time.Second)
    }
}

func monitorInternalGRPC() {
    for {
        callGRPCMethod("ListTasks")
        callGRPCMethod("GetTask")

        time.Sleep(30 * time.Second)
    }
}

func callGRPCMethod(method string) {
    start := time.Now()

    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Printf("❌ gRPC connection failed: %v", err)
        recordGRPCMetrics(method, "error", start)
        serviceHealth.WithLabelValues("todo-gercep-grpc", "grpc").Set(0)
        return
    }
    defer conn.Close()

    client := pb.NewTodoServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var resp interface{}

    switch method {
    case "ListTasks":
        resp, err = client.ListTasks(ctx, &pb.EmptyRequest{})
        if err == nil {
            if listResp, ok := resp.(*pb.TaskListResponse); ok {
                taskCount.Set(float64(listResp.Total))
                log.Printf("✅ ListTasks: %d tasks", listResp.Total)
            }
        }
    case "GetTask":
        resp, err = client.GetTask(ctx, &pb.GetTaskRequest{Id: 1})
        if err == nil {
            if taskResp, ok := resp.(*pb.TaskResponse); ok && taskResp.Success {
                log.Printf("✅ GetTask: %s", taskResp.Task.Title)
            }
        }
    default:
        return
    }

    if err != nil {
        log.Printf("❌ gRPC %s failed: %v", method, err)
        recordGRPCMetrics(method, "error", start)
        serviceHealth.WithLabelValues("todo-gercep-grpc", "grpc").Set(0)
        return
    }

    recordGRPCMetrics(method, "success", start)
    serviceHealth.WithLabelValues("todo-gercep-grpc", "grpc").Set(1)
}

func recordGRPCMetrics(method, status string, start time.Time) {
    latency := time.Since(start).Seconds()
    grpcLatency.WithLabelValues(method, status).Observe(latency)
    grpcRequestsTotal.WithLabelValues(method, status).Inc()
}
package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "sync"
    "time"

    pb "todo-grpc/proto"
    "google.golang.org/grpc"
)

type TaskDB struct {
    mu     sync.RWMutex
    tasks  map[int32]*pb.Task
    nextID int32
}

func NewTaskDB() *TaskDB {
    return &TaskDB{
        tasks:  make(map[int32]*pb.Task),
        nextID: 1,
    }
}

func (db *TaskDB) Create(title, desc string) *pb.Task {
    db.mu.Lock()
    defer db.mu.Unlock()

    task := &pb.Task{
        Id:          db.nextID,
        Title:       title,
        Description: desc,
        Completed:   false,
        CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
    }

    db.tasks[db.nextID] = task
    db.nextID++

    return task
}

func (db *TaskDB) GetAll() []*pb.Task {
    db.mu.RLock()
    defer db.mu.RUnlock()

    tasks := make([]*pb.Task, 0, len(db.tasks))
    for _, task := range db.tasks {
        tasks = append(tasks, task)
    }
    return tasks
}

func (db *TaskDB) GetByID(id int32) (*pb.Task, bool) {
    db.mu.RLock()
    defer db.mu.RUnlock()

    task, exists := db.tasks[id]
    return task, exists
}

func (db *TaskDB) Complete(id int32) (*pb.Task, bool) {
    db.mu.Lock()
    defer db.mu.Unlock()

    task, exists := db.tasks[id]
    if !exists {
        return nil, false
    }

    task.Completed = true
    return task, true
}

func (db *TaskDB) Delete(id int32) bool {
    db.mu.Lock()
    defer db.mu.Unlock()

    _, exists := db.tasks[id]
    if exists {
        delete(db.tasks, id)
    }
    return exists
}

type todoServer struct {
    pb.UnimplementedTodoServiceServer
    db *TaskDB
}

func (s *todoServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskResponse, error) {
    log.Printf("[CREATE] Title: %s", req.Title)

    task := s.db.Create(req.Title, req.Description)

    return &pb.TaskResponse{
        Success: true,
        Message: fmt.Sprintf("Task '%s' created successfully with ID %d", req.Title, task.Id),
        Task:    task,
    }, nil
}

func (s *todoServer) ListTasks(ctx context.Context, req *pb.EmptyRequest) (*pb.TaskListResponse, error) {
    tasks := s.db.GetAll()

    log.Printf("[LIST] Returning %d tasks", len(tasks))

    return &pb.TaskListResponse{
        Tasks: tasks,
        Total: int32(len(tasks)),
    }, nil
}

func (s *todoServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.TaskResponse, error) {
    log.Printf("[GET] Looking for task ID: %d", req.Id)

    task, exists := s.db.GetByID(req.Id)
    if !exists {
        return &pb.TaskResponse{
            Success: false,
            Message: fmt.Sprintf("Task with ID %d not found", req.Id),
        }, nil
    }

    return &pb.TaskResponse{
        Success: true,
        Message: "Task found",
        Task:    task,
    }, nil
}

func (s *todoServer) CompleteTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.TaskResponse, error) {
    log.Printf("[COMPLETE] Marking task ID %d as completed", req.Id)

    task, success := s.db.Complete(req.Id)
    if !success {
        return &pb.TaskResponse{
            Success: false,
            Message: fmt.Sprintf("Task with ID %d not found", req.Id),
        }, nil
    }

    return &pb.TaskResponse{
        Success: true,
        Message: fmt.Sprintf("Task '%s' completed!", task.Title),
        Task:    task,
    }, nil
}

func (s *todoServer) DeleteTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.DeleteResponse, error) {
    log.Printf("[DELETE] Removing task ID: %d", req.Id)

    success := s.db.Delete(req.Id)
    if !success {
        return &pb.DeleteResponse{
            Success: false,
            Message: fmt.Sprintf("Task with ID %d not found", req.Id),
        }, nil
    }

    return &pb.DeleteResponse{
        Success: true,
        Message: fmt.Sprintf("Task ID %d successfully deleted", req.Id),
    }, nil
}

func main() {
    db := NewTaskDB()
    db.Create("Learn gRPC", "Learn how gRPC works and build a project")
    db.Create("Build REST API", "Create REST API as frontend later")
    db.Create("Deploy to server", "Deploy service to production")

    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    pb.RegisterTodoServiceServer(grpcServer, &todoServer{db: db})

    log.Println("========================================")
    log.Println("TO-DO LIST SERVICE (gRPC)")
    log.Println("========================================")
    log.Println("Server running on port :50051")
    log.Println("Available methods:")
    log.Println("  • CreateTask   - Add new task")
    log.Println("  • ListTasks    - View all tasks")
    log.Println("  • GetTask      - View task details")
    log.Println("  • CompleteTask - Mark task as done")
    log.Println("  • DeleteTask   - Remove task")
    log.Println("========================================")

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"

    pb "todo-grpc/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type TodoClient struct {
    client pb.TodoServiceClient
}

func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := &TodoClient{
        client: pb.NewTodoServiceClient(conn),
    }

    client.runCLI()
}

func (c *TodoClient) runCLI() {
    reader := bufio.NewReader(os.Stdin)

    for {
        c.showMenu()

        fmt.Print("\nSelect menu (1-6): ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        switch input {
        case "1":
            c.createTask(reader)
        case "2":
            c.listTasks()
        case "3":
            c.getTask(reader)
        case "4":
            c.completeTask(reader)
        case "5":
            c.deleteTask(reader)
        case "6":
            fmt.Println("\n Goodbye!")
            return
        default:
            fmt.Println(" Invalid choice!")
        }
    }
}

func (c *TodoClient) showMenu() {
    fmt.Println("\n========================================")
    fmt.Println(" TO-DO LIST MANAGER")
    fmt.Println("========================================")
    fmt.Println("1.  Add new task")
    fmt.Println("2.  View all tasks")
    fmt.Println("3.  View task details")
    fmt.Println("4.  Mark task as completed")
    fmt.Println("5.  Delete task")
    fmt.Println("6.  Exit")
    fmt.Println("========================================")
}

func (c *TodoClient) createTask(reader *bufio.Reader) {
    fmt.Println("\n ADD NEW TASK")
    
    fmt.Print("Task title: ")
    title, _ := reader.ReadString('\n')
    title = strings.TrimSpace(title)
    
    fmt.Print("Description: ")
    desc, _ := reader.ReadString('\n')
    desc = strings.TrimSpace(desc)
    
    req := &pb.CreateTaskRequest{
        Title:       title,
        Description: desc,
    }
    
    resp, err := c.client.CreateTask(context.Background(), req)
    if err != nil {
        fmt.Printf(" Error: %v\n", err)
        return
    }
    
    fmt.Printf("\n %s\n", resp.Message)
}

func (c *TodoClient) listTasks() {
    fmt.Println("\n ALL TASKS")
    fmt.Println("========================================")
    
    resp, err := c.client.ListTasks(context.Background(), &pb.EmptyRequest{})
    if err != nil {
        fmt.Printf(" Error: %v\n", err)
        return
    }
    
    if resp.Total == 0 {
        fmt.Println(" No tasks yet! Add your first task.")
        return
    }
    
    for _, task := range resp.Tasks {
        status := "Task To Do"
        if task.Completed {
            status = " Done"
        }
        fmt.Printf("%s [%d] %s\n", status, task.Id, task.Title)
        fmt.Printf("   └─ %s\n", task.Description)
        fmt.Printf("   └─ Created: %s\n", task.CreatedAt)
        fmt.Println()
    }
    
    fmt.Printf("Total: %d task(s)\n", resp.Total)
}

func (c *TodoClient) getTask(reader *bufio.Reader) {
    fmt.Print("\n Enter task ID: ")
    idStr, _ := reader.ReadString('\n')
    idStr = strings.TrimSpace(idStr)
    
    id, err := strconv.Atoi(idStr)
    if err != nil {
        fmt.Println(" ID must be a number!")
        return
    }
    
    req := &pb.GetTaskRequest{Id: int32(id)}
    resp, err := c.client.GetTask(context.Background(), req)
    if err != nil {
        fmt.Printf(" Error: %v\n", err)
        return
    }
    
    if !resp.Success {
        fmt.Printf(" %s\n", resp.Message)
        return
    }
    
    task := resp.Task
    status := "Task To Do"
    if task.Completed {
        status = " Completed!"
    }
    
    fmt.Println("\n========================================")
    fmt.Printf(" ID: %d\n", task.Id)
    fmt.Printf(" Title: %s\n", task.Title)
    fmt.Printf(" Description: %s\n", task.Description)
    fmt.Printf(" Status: %s\n", status)
    fmt.Printf(" Created: %s\n", task.CreatedAt)
    fmt.Println("========================================")
}

func (c *TodoClient) completeTask(reader *bufio.Reader) {
    fmt.Print("\n Enter task ID to mark as completed: ")
    idStr, _ := reader.ReadString('\n')
    idStr = strings.TrimSpace(idStr)
    
    id, err := strconv.Atoi(idStr)
    if err != nil {
        fmt.Println(" ID must be a number!")
        return
    }
    
    req := &pb.GetTaskRequest{Id: int32(id)}
    resp, err := c.client.CompleteTask(context.Background(), req)
    if err != nil {
        fmt.Printf(" Error: %v\n", err)
        return
    }
    
    if !resp.Success {
        fmt.Printf(" %s\n", resp.Message)
        return
    }
    
    fmt.Printf("\n %s\n", resp.Message)
}

func (c *TodoClient) deleteTask(reader *bufio.Reader) {
    fmt.Print("\n Enter task ID to delete: ")
    idStr, _ := reader.ReadString('\n')
    idStr = strings.TrimSpace(idStr)
    
    id, err := strconv.Atoi(idStr)
    if err != nil {
        fmt.Println("❌ ID must be a number!")
        return
    }
    
    req := &pb.GetTaskRequest{Id: int32(id)}
    resp, err := c.client.DeleteTask(context.Background(), req)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    if !resp.Success {
        fmt.Printf("%s\n", resp.Message)
        return
    }
    
    fmt.Printf("\n %s\n", resp.Message)
}
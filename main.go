package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type Status int

const (
	StatusWaiting Status = iota
	StatusRunning
	StatusCompleted
)

func (s Status) String() string {
	switch s {
	case StatusWaiting:
		return "Waiting"
	case StatusRunning:
		return "Running"
	case StatusCompleted:
		return "Completed"
	default:
		return "Unknown"
	}
}

type Task struct {
	ID       int
	Memory   int
	Duration time.Duration
	Status   Status
}

func NewTask(id int) *Task {
	return &Task{
		ID:       id,
		Memory:   rand.Intn(9) + 1,
		Duration: time.Duration(rand.Intn(5)+1) * time.Second,
		Status:   StatusWaiting,
	}
}

func (t *Task) Run() {
	t.Status = StatusRunning

	mem := make([]byte, t.Memory*1024*1024)
	_ = mem

	time.Sleep(t.Duration)

	t.Status = StatusCompleted
}

type TaskScheduler struct {
	mu              sync.Mutex
	wg              sync.WaitGroup
	tasks           []*Task
	runningTasks    chan struct{}
	totalMemoryUsed int
}

func NewScheduler(numTask, maxConcurrentTask int) TaskScheduler {
	tasks := make([]*Task, numTask)
	for i := range tasks {
		tasks[i] = NewTask(i)
	}

	return TaskScheduler{
		mu:              sync.Mutex{},
		wg:              sync.WaitGroup{},
		tasks:           tasks,
		runningTasks:    make(chan struct{}, maxConcurrentTask),
		totalMemoryUsed: 0,
	}
}

func (s *TaskScheduler) SumMemory(mem int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalMemoryUsed += mem
}

func (s *TaskScheduler) ExecuteTask(task *Task) {
	s.runningTasks <- struct{}{}
	s.wg.Add(1)

	go func() {
		defer func() {
			<-s.runningTasks
			s.wg.Done()
		}()

		task.Run()
		s.SumMemory(task.Memory)
	}()
}

func (s *TaskScheduler) Start() {
	for _, task := range s.tasks {
		s.ExecuteTask(task)
	}

	s.wg.Wait()
	close(s.runningTasks)

	time.Sleep(200 * time.Millisecond)
	averageMemory := float64(s.totalMemoryUsed) / float64(len(s.tasks))
	fmt.Printf("Average Memory Usage: %.2f MB\n", averageMemory)
}

func (s *TaskScheduler) Monitor() {
	for {
		fmt.Printf("Using %d CPU cores:\n", runtime.NumCPU())
		fmt.Println("-------------------")

		for _, task := range s.tasks {
			if task.Status == StatusRunning {
				fmt.Printf("Task %d: Running (Memory: %d MB)\n", task.ID, task.Memory)
			} else {
				fmt.Printf("Task %d: %s\n", task.ID, task.Status)
			}
		}

		time.Sleep(100 * time.Millisecond)
		clearScreen := "\033[2J\033[H"
		fmt.Print(clearScreen)
	}
}

func main() {
	numCPU := runtime.NumCPU()
	scheduler := NewScheduler(40, numCPU)
	go scheduler.Monitor()
	scheduler.Start()
}

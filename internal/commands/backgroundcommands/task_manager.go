package backgroundcommands

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

type TaskStatus string

// 首先添加一个辅助类型来包装 Reader 和 Writer
type taskIO struct {
	reader io.Reader
	writer io.Writer
}

// 实现 io.Reader 接口
func (t *taskIO) Read(p []byte) (n int, err error) {
	return t.reader.Read(p)
}

// 实现 io.Writer 接口
func (t *taskIO) Write(p []byte) (n int, err error) {
	return t.writer.Write(p)
}

const (
	TaskStatusRunning  TaskStatus = "RUNNING"
	TaskStatusStopped  TaskStatus = "STOPPED"
	TaskStatusFinished TaskStatus = "FINISHED"
)

// Task 结构体增加更多信息
type Task struct {
	ID           int
	Name         string
	Args         []string
	Status       TaskStatus
	StartTime    time.Time
	InputWriter  io.WriteCloser
	OutputReader io.Reader
	outputBuffer *bytes.Buffer
	outputLock   sync.Mutex
	Done         chan struct{}
}

type TaskManager struct {
	tasks       map[int]*Task
	tasksLock   sync.Mutex
	taskID      int
	funcMap     map[string]interface{}
	pool        *ants.Pool
	poolSize    int
	isRebooting bool
}

func NewTaskManager(poolSize int) (*TaskManager, error) {
	pool, err := ants.NewPool(poolSize)
	if err != nil {
		return nil, err
	}
	return &TaskManager{
		tasks:    make(map[int]*Task),
		funcMap:  make(map[string]interface{}),
		pool:     pool,
		poolSize: poolSize,
	}, nil
}

// Reboot 重启任务管理器
func (tm *TaskManager) Reboot(rw io.ReadWriter) error {
	tm.tasksLock.Lock()
	if tm.isRebooting {
		tm.tasksLock.Unlock()
		return fmt.Errorf("reboot already in progress")
	}
	tm.isRebooting = true
	defer func() { tm.isRebooting = false }()

	// 保存当前运行的任务信息
	runningTasks := make([]struct {
		name string
		args []string
	}, 0)

	for _, task := range tm.tasks {
		if task.Status == TaskStatusRunning {
			runningTasks = append(runningTasks, struct {
				name string
				args []string
			}{task.Name, task.Args})
		}
	}

	// 关闭所有任务
	for _, task := range tm.tasks {
		task.Status = TaskStatusStopped
		task.InputWriter.Close()
		close(task.Done)
	}
	tm.tasks = make(map[int]*Task)

	// 关闭并重新创建协程池
	tm.pool.Release()
	pool, err := ants.NewPool(tm.poolSize)
	if err != nil {
		tm.tasksLock.Unlock()
		return fmt.Errorf("failed to create new pool: %v", err)
	}
	tm.pool = pool
	tm.tasksLock.Unlock()

	// 重启之前运行的任务
	fmt.Fprintln(rw, "Rebooting task manager...")
	for _, t := range runningTasks {
		tm.StartTask(rw, t.name, t.args...)
	}
	fmt.Fprintln(rw, "Task manager rebooted successfully")
	return nil
}

func (tm *TaskManager) StartTask(rw io.ReadWriter, name string, args ...string) {
	tm.tasksLock.Lock()
	defer tm.tasksLock.Unlock()

	tm.taskID++
	inputReader, inputWriter := io.Pipe()
	outputReader, outputWriter := io.Pipe()
	outputBuffer := bytes.NewBuffer(nil)

	task := &Task{
		ID:           tm.taskID,
		Name:         name,
		Args:         args,
		Status:       TaskStatusRunning,
		StartTime:    time.Now(),
		InputWriter:  inputWriter,
		OutputReader: outputReader,
		outputBuffer: outputBuffer,
		Done:         make(chan struct{}),
	}

	go func() {
		defer outputReader.Close()
		buf := make([]byte, 1024)
		for {
			n, err := outputReader.Read(buf)
			if err != nil {
				return
			}
			task.outputLock.Lock()
			task.outputBuffer.Write(buf[:n])
			task.outputLock.Unlock()
		}
	}()

	tm.tasks[tm.taskID] = task

	err := tm.pool.Submit(func() {
		tm.runTask(task, inputReader, outputWriter)
	})
	if err != nil {
		task.Status = TaskStatusStopped
		fmt.Fprintf(rw, "Failed to start task %d: %v\n", task.ID, err)
		return
	}
	fmt.Fprintf(rw, "Started task %d: %s %v\n", task.ID, task.Name, task.Args)
}

// ListTasks 修改状态显示
func (tm *TaskManager) ListTasks(rw io.ReadWriter) {
	tm.tasksLock.Lock()
	defer tm.tasksLock.Unlock()

	// 只显示正在运行的任务
	runningTasks := make([]*Task, 0)
	for _, task := range tm.tasks {
		if task.Status == TaskStatusRunning {
			runningTasks = append(runningTasks, task)
		}
	}

	if len(runningTasks) == 0 {
		fmt.Fprintln(rw, "No running background tasks")
		return
	}

	// 简化输出格式
	for _, task := range runningTasks {
		fmt.Fprintf(rw, "%d\t%s\t%s\t%s\n",
			task.ID,
			task.StartTime.Format("2006-01-02 15:04:05"),
			task.Name,
			strings.Join(task.Args, " "))
	}
}

// InteractTask 优化交互体验
func (tm *TaskManager) InteractTask(rw io.ReadWriter, taskIDStr string) error {
	id, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return fmt.Errorf("invalid task ID: %v", err)
	}

	tm.tasksLock.Lock()
	task, exists := tm.tasks[id]
	if !exists || task.Status != TaskStatusRunning {
		tm.tasksLock.Unlock()
		return fmt.Errorf("task %d not found or not running", id)
	}
	tm.tasksLock.Unlock()

	fmt.Fprintf(rw, "Interacting with task %d\n", id)

	task.outputLock.Lock()
	history := task.outputBuffer.String()
	task.outputLock.Unlock()
	fmt.Fprint(rw, history)

	quit := make(chan struct{})
	done := make(chan struct{})
	outputDone := make(chan struct{}) // 新增通道用于同步输出完成

	// 监听任务输出
	go func() {
		defer close(outputDone) // 确保在函数结束时通知输出已完成
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		prevLen := len(history)
		for {
			select {
			case <-ticker.C:
				task.outputLock.Lock()
				current := task.outputBuffer.String()
				task.outputLock.Unlock()
				if len(current) > prevLen {
					fmt.Fprint(rw, current[prevLen:])
					prevLen = len(current)
				}
			case <-done:
				// 在任务完成时，确保输出最后的内容
				task.outputLock.Lock()
				current := task.outputBuffer.String()
				task.outputLock.Unlock()
				if len(current) > prevLen {
					fmt.Fprint(rw, current[prevLen:])
				}
				return
			case <-quit:
				return
			}
		}
	}()

	// 处理用户输入
	scanner := bufio.NewScanner(rw)
	for {
		select {
		case <-task.Done:
			close(done)  // 通知输出 goroutine 任务已完成
			<-outputDone // 等待所有输出完成
			fmt.Fprintln(rw, "\nTask completed")
			return nil
		default:
		}

		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if input == "exit" {
			close(quit)
			<-outputDone // 等待输出完成
			return nil
		}

		if _, err := fmt.Fprintln(task.InputWriter, input); err != nil {
			if err == io.ErrClosedPipe {
				return fmt.Errorf("task has finished")
			}
			return fmt.Errorf("error sending input: %v", err)
		}
	}
	return nil
}

func (tm *TaskManager) KillTask(rw io.ReadWriter, taskIDStr string) error {
	id, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return fmt.Errorf("invalid task ID: %v", err)
	}

	tm.tasksLock.Lock()
	defer tm.tasksLock.Unlock()

	task, exists := tm.tasks[id]
	if !exists {
		return fmt.Errorf("task %d not found", id)
	}

	if task.Status != TaskStatusRunning {
		return fmt.Errorf("task %d is not running", id)
	}

	task.Status = TaskStatusStopped
	task.InputWriter.Close()
	close(task.Done)
	delete(tm.tasks, id)

	fmt.Fprintf(rw, "Killed task %d (%s)\n", id, task.Name)
	return nil
}

func (tm *TaskManager) runTask(task *Task, input io.Reader, output io.Writer) {
	defer func() {
		task.InputWriter.Close()
		if closer, ok := output.(io.Closer); ok {
			closer.Close()
		}
		task.Status = TaskStatusFinished
		close(task.Done)
		tm.removeTask(task.ID)
	}()

	fn, exists := tm.funcMap[task.Name]
	if !exists {
		fmt.Fprintf(output, "Function %s not found\n", task.Name)
		return
	}

	// 创建一个包装了 input 和 output 的 taskIO
	rw := &taskIO{
		reader: input,
		writer: output,
	}

	// 调用注册的函数
	reflect.ValueOf(fn).Call([]reflect.Value{
		reflect.ValueOf(rw),
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(task.Args),
	})
}

func (tm *TaskManager) removeTask(id int) {
	tm.tasksLock.Lock()
	defer tm.tasksLock.Unlock()
	delete(tm.tasks, id)
}
func (tm *TaskManager) RegisterFunction(name string, fn interface{}) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic("not a function")
	}

	t := v.Type()
	// 修改参数检查逻辑
	if t.NumIn() != 3 ||
		// 检查第一个参数是否为 *ReadWriter 或实现了 io.Reader 和 io.Writer 接口的类型
		!(t.In(0).Implements(reflect.TypeOf((*io.Reader)(nil)).Elem()) &&
			t.In(0).Implements(reflect.TypeOf((*io.Writer)(nil)).Elem())) ||
		// 检查第二个参数是否为 context.Context
		!t.In(1).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) ||
		// 检查第三个参数是否为 []string
		t.In(2) != reflect.TypeOf([]string{}) {
		panic("function signature must be func(*ReadWriter, context.Context, []string)")
	}

	tm.funcMap[name] = fn
}

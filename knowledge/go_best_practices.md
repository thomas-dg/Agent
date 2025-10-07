# Go 代码生成最佳实践（使用1.24+）
## 一、核心指导原则（基石）
1. **简洁优先**：优先选择简单清晰的方案（如用`bytes.Buffer`而非复杂拼接逻辑），避免过度设计
   > 例：用`sync.Mutex`零值直接使用，无需显式初始化
2. **可读性至上**：代码为人类编写（如变量名体现用途而非类型），机器执行是次要目的
   > 例：用`userIDs`而非`usersMap`（Map类型由编译器保障，后缀冗余）
3. **生产力导向**：利用Go工具链（`gofmt`/`go mod`/`go tool`）减少摩擦，避免重复工作
   > 例：`go mod`中用`tool`指令管理工具依赖，而非全局安装

## 二、标识符与注释（代码可读性核心）
### 2.1 标识符命名
- **规则1：清晰优先于简短**  
  作用域越广、生命周期越长，名称应越长（循环变量用`i`，包级变量用`userCache`）
  ```go
  // 好：循环变量短（生命周期短）
  for i := 0; i < len(items); i++ { ... }
  // 好：包级变量长（生命周期长）
  var userCache = sync.Map{} // 存储用户ID到用户信息的缓存
  ```
- **规则2：不按类型命名**  
  避免变量名包含类型（如`usersMap`→`users`，`configPtr`→`conf`）
- **规则3：一致性**  
  同类型参数/接收器名统一（如数据库句柄始终用`db *sql.DB`，接收器用固定单字母`r *Reader`）
- **规则4：声明风格**  
  未初始化用`var`（显式零值），初始化用`:=`（显式赋值）
  ```go
  var buf []byte // 零值切片，后续赋值
  user := &User{Name: "Alice"} // 显式初始化
  ```

### 2.2 注释规范
- **规则1：变量/常量注释描述内容**  
  不重复用途（用途由变量名体现），描述值来源或初始化者
  ```go
  const StatusOK = 200 // RFC 7231, 6.3.1（描述值的标准依据）
  var sizeCalculationDisabled bool // 由dowidth函数维护状态（描述初始化者）
  ```
- **规则2：公共符号必文档化**  
  包、函数、方法等公共符号需`godoc`风格注释，接口实现无需重复注释
  ```go
  // ReadAll 从r读取数据直到EOF或错误，成功时err==nil（非EOF）
  func ReadAll(r io.Reader) ([]byte, error) { ... }
  ```
- **规则3：不注释坏代码，重写它**  
  若需注释解释复杂逻辑，优先提取为函数（如长if条件→`anyPositive`函数）

## 三、包与项目设计（可维护性基础）
### 3.1 包设计
- **规则1：包名体现“提供的服务”**  
  避免`base`/`common`/`util`，用`strings`（字符串处理）、`http`（HTTP服务）类名称
  > 例：将`util`包拆分为`fileutil`（文件操作）、`strutil`（字符串工具）
- **规则2：让零值有用**  
  设计结构体时确保零值可直接使用（如`sync.Mutex`、`bytes.Buffer`）
  ```go
  var buf bytes.Buffer
  buf.WriteString("hello") // 零值直接使用，无需初始化
  ```
- **规则3：提前返回，避免嵌套**  
  用守卫语句（guard clause）处理异常路径，成功路径向下延伸
  ```go
  func UnreadRune(b *Buffer) error {
    if b.lastRead <= opInvalid {
      return errors.New("前序操作非ReadRune") // 提前返回异常
    }
    // 成功路径
    if b.off >= int(b.lastRead) {
      b.off -= int(b.lastRead)
    }
    b.lastRead = opInvalid
    return nil
  }
  ```
- **规则4：避免包级状态**  
  将全局变量转为结构体字段，用接口解耦依赖
  ```go
  // 差：包级状态（耦合所有使用处）
  var globalDB *sql.DB
  // 好：结构体封装状态
  type UserService struct {
    db *sql.DB // 依赖注入，无全局状态
  }
  ```

### 3.2 项目结构
- **规则1：少而大的包**  
  避免过度拆分（Java包≈Go文件，Go包≈Maven模块），按“功能”而非“类型”组织
  > 例：`net/http`用`client.go`/`server.go`分文件，而非`client`/`server`子包
- **规则2：main包最小化**  
  `main`仅做参数解析、资源初始化，业务逻辑移至其他包
  ```go
  // main.go（仅初始化）
  func main() {
    cfg := parseConfig()
    db := initDB(cfg.DBAddr)
    svc := user.NewService(db) // 业务逻辑在user包
    log.Fatal(http.ListenAndServe(":8080", svc.Handler()))
  }
  ```
- **规则3：内部测试优先**  
  单元测试用`package xxx`（内部测试，可访问非导出成员），示例用`package xxx_test`（外部测试，符合调用者视角）

## 四、API设计（难误用、高通用）
### 4.1 难误用的API
- **规则1：避免同类型多参数**  
  同类型参数易传错顺序（如`CopyFile(to, from string)`），用类型封装解决
  ```go
  type Source string
  func (src Source) CopyTo(dest string) error {
    return CopyFile(dest, string(src)) // 确保参数顺序正确
  }
  ```
- **规则2：默认用例优先，避免nil参数**  
  不为简化少写参数而支持nil（如`http.ListenAndServe`允许nil Handler易混淆），显式传默认值
  ```go
  // 好：显式传默认ServeMux，无歧义
  http.ListenAndServe(":8080", http.DefaultServeMux)
  ```
- **规则3：用变长参数替代切片参数**  
  需至少一个参数时，组合“固定+变长”参数（避免空调用）
  ```go
  // 确保至少传1个int，其余可变
  func anyPositive(first int, rest ...int) bool {
    if first > 0 { return true }
    for _, v := range rest { if v > 0 { return true } }
    return false
  }
  ```

### 4.2 接口隔离（函数定义所需行为）
- **规则：依赖最小接口，而非具体类型**  
  函数参数用接口（如`io.Writer`）而非具体实现（如`*os.File`），提升通用性
  ```go
  // 好：支持任何io.Writer（文件、网络、缓冲区）
  func SaveDoc(w io.Writer, doc *Document) error {
    _, err := w.Write(doc.Marshal())
    return err
  }
  // 差：仅支持*os.File，无法用于网络写入
  func SaveDoc(f *os.File, doc *Document) error { ... }
  ```

## 五、错误处理（清晰、无冗余）
### 5.1 核心规则
- **规则1：消除错误而非处理**  
  用抽象封装减少错误（如`bufio.Scanner`替代`bufio.Reader`的EOF处理）
  ```go
  // 差：手动处理EOF，逻辑冗余
  func CountLines(r io.Reader) (int, error) {
    br := bufio.NewReader(r)
    lines := 0
    for {
      _, err := br.ReadString('\n')
      lines++
      if err != nil { break }
    }
    if err != io.EOF { return 0, err }
    return lines, nil
  }
  // 好：用Scanner消除EOF处理
  func CountLines(r io.Reader) (int, error) {
    sc := bufio.NewScanner(r)
    lines := 0
    for sc.Scan() { lines++ }
    return lines, sc.Err() // 自动处理EOF
  }
  ```
- **规则2：仅处理一次错误**  
  避免“日志+返回”（重复处理），用错误包装传递上下文
  ```go
  // 差：日志又返回，重复处理
  func WriteConfig(w io.Writer, conf *Config) error {
    buf, err := json.Marshal(conf)
    if err != nil {
      log.Printf("marshal err: %v", err) // 处理1次
      return err // 处理2次
    }
    return WriteAll(w, buf)
  }
  // 好：包装错误，上层统一处理
  func WriteConfig(w io.Writer, conf *Config) error {
    buf, err := json.Marshal(conf)
    if err != nil {
      return fmt.Errorf("marshal config: %w", err) // 仅处理1次
    }
    if err := WriteAll(w, buf); err != nil {
      return fmt.Errorf("write config: %w", err)
    }
    return nil
  }
  ```
- **规则3：用errWriter减少重复判断**  
  批量写操作时，用包装类型缓存错误，避免每次判断
  ```go
  type errWriter struct {
    io.Writer
    err error
  }
  func (e *errWriter) Write(b []byte) (int, error) {
    if e.err != nil { return 0, e.err }
    n, e.err := e.Writer.Write(b)
    return n, nil
  }
  // 使用：无需每次判断err
  func WriteResponse(w io.Writer, status string) error {
    ew := &errWriter{Writer: w}
    fmt.Fprintf(ew, "HTTP/1.1 %s\r\n", status)
    fmt.Fprint(ew, "\r\n")
    return ew.err
  }
  ```

### 5.2 错误判断与包装（Go 1.24+）
- 用`fmt.Errorf("%w", err)`包装错误，`errors.Is`匹配原始错误，`errors.As`断言类型
  ```go
  err := WriteConfig(w, conf)
  if errors.Is(err, os.ErrNotExist) {
    // 处理文件不存在
  } else if errors.As(err, &json.SyntaxError) {
    // 处理JSON语法错误
  }
  ```

## 六、并发编程（安全、可控）
### 6.1 goroutine管理
- **规则1：不滥用goroutine**  
  若goroutine结果需同步等待，优先直接执行（如http服务无需起goroutine）
  ```go
  // 差：多余goroutine，需空select阻塞main
  go http.ListenAndServe(":8080", mux)
  select {}
  // 好：main直接执行，自然阻塞
  log.Fatal(http.ListenAndServe(":8080", mux))
  ```
- **规则2：并发留给调用者**  
  函数不主动起goroutine，由调用者决定是否异步执行
  ```go
  // 好：函数同步执行，并发由调用者控制
  func ServeApp(addr string) error {
    mux := http.NewServeMux()
    return http.ListenAndServe(addr, mux)
  }
  // 调用者决定并发：
  go ServeApp(":8080")
  ```
- **规则3：启动goroutine必知停止时机**  
  用`context`/`channel`控制生命周期，避免泄漏
  ```go
  // HTTP服务优雅关闭（结合context和channel）
  func serve(addr string, stop <-chan struct{}) error {
    s := &http.Server{Addr: addr}
    go func() {
      <-stop // 等待停止信号
      s.Shutdown(context.Background())
    }()
    return s.ListenAndServe()
  }
  // 使用：
  done := make(chan error, 2)
  stop := make(chan struct{})
  go func() { done <- serve(":8080", stop) }()
  go func() { done <- serve(":8001", stop) }()
  // 任一服务出错，关闭所有
  for i := 0; i < cap(done); i++ {
    if err := <-done; err != nil {
      log.Println("err:", err)
      close(stop) // 触发所有服务关闭
    }
  }
  ```

### 6.2 并发控制工具（结合Go 1.24）
- **场景1：任一goroutine出错停止→errgroup.Group**
  ```go
  g, ctx := errgroup.WithContext(context.Background())
  for _, task := range tasks {
    task := task
    g.Go(func() error { return task.Run(ctx) })
  }
  if err := g.Wait(); err != nil { ... }
  ```
- **场景2：无需感知错误→sync.WaitGroup**
  ```go
  var wg sync.WaitGroup
  for _, task := range tasks {
    wg.Add(1)
    go func(t Task) {
      defer wg.Done()
      if err := t.Run(); err != nil { log.Println(err) }
    }(task)
  }
  wg.Wait()
  ```
- **场景3：并发安全Map→读多写少用sync.Map，写多读少用map+sync.RWMutex**
  ```go
  // 读多写少场景
  var userCache sync.Map
  userCache.Store("123", &User{ID: "123"})
  if u, ok := userCache.Load("123"); ok { ... }
  ```

## 七、Go 1.24+高级特性实践
### 7.1 泛型优化
- **泛型类型别名（简化通用结构）**
  ```go
  // 通用结果包装器
  type Result[T any] = struct{ Data T; Err error; Code int }
  // 使用：返回User类型结果
  func GetUser(id string) Result[User] {
    u, err := db.GetUser(id)
    return Result[User]{Data: u, Err: err, Code: 200}
  }
  ```
- **泛型约束（明确类型范围）**
  ```go
  import "golang.org/x/exp/constraints"
  type Number interface { constraints.Integer | constraints.Float }
  func Sum[T Number](a, b T) T { return a + b }
  ```

### 7.2 资源清理与工具链
- **runtime.AddCleanup（灵活清理，替代部分defer）**
  ```go
  res := NewResource()
  runtime.AddCleanup(res, func(r *Resource) { r.Release() }) // 自动清理
  ```
- **工具依赖管理（go mod中tool指令）**
  ```go
  // go.mod中声明工具依赖
  tool (
    golang.org/x/tools/cmd/stringer v0.19.0
    honnef.co/go/tools/cmd/staticcheck latest
  )
  // 执行工具：go tool staticcheck ./...
  ```

## 八、语言风格（避免将 Java、C# 等面向对象语言的设计模式和习惯带入 Go 代码中）
### 1. 禁止使用 Java 风格的命名
- **绝对禁止**创建名为 `XxxManager`, `XxxFactory`, `XxxUtil`, `XxxHelper` 的结构体或包。
- **替代方案**：使用更具体、更简单的名词。例如，负责用户存储的功能，应命名为 `UserStore` 或直接使用函数 `SaveUser`，而非 `UserManager`。

### 2. 推崇简单函数与小接口
- **函数优先**：对于独立操作，优先使用**包级函数**，而非必须先创建结构体实例的方法。
- **接口小而美**：接口应该只包含 1-3 个方法（例如 `io.Reader` 只有 `Read` 方法）。**坚决反对**包含几十个方法的“胖接口”。
- **组合优于继承**：使用结构体嵌入（embedding）来复用功能，而不是构建复杂的继承层次。

### 3. 地道的错误处理
- **禁止异常风格**：绝不使用 `panic` 来处理正常的业务错误，除非是不可恢复的致命错误。
- **显式错误处理**：函数操作可能失败时，必须返回 `(result, error)` 类型，并强制调用者检查 `error`。

### 4. 项目结构与包设计
- **扁平化结构**：包目录结构不宜过深。包名应短小、简洁、见名知义。
- **包职责单一**：一个包只做一件事，并提供清晰的 API（即导出的函数和类型）。

### 5. 并发模型
- **使用 Go 原生并发**：优先使用 **goroutine** 和 **channel** 进行并发通信。
- **明确同步机制**：使用 `sync.Mutex`, `sync.WaitGroup`, `errgroup.Group` 等标准库工具时，需确保逻辑清晰。

### 正面与反面示例

#### 命名示例
- **反例** (`Java` 风格): `type UserManager struct {}`, `func (um *UserManager) ProcessUser()`
- **正例** (`Go` 风格): `type UserStore struct {}`, `func SaveUser(user User) error`

#### 接口示例
- **反例** (`Java` 风格):
  ```go
  // 过于庞大的接口
  type UserService interface {
      CreateUser(...)
      UpdateUser(...)
      DeleteUser(...)
      GetUser(...)
      ListUsers(...)
      // ... 更多方法
  }
  ```
- **正例** (`Go` 风格):
  ```go
  // 拆分为小而专注的接口
  type UserCreator interface {
      CreateUser(...) error
  }
  type UserGetter interface {
      GetUser(...) (*User, error)
  }
  ```


## 九、禁用/慎用规则（生产避坑）
1. 禁用`init`函数→用显式初始化函数（如`InitConfig()`）
2. 慎用`unsafe`包→仅性能瓶颈且无替代方案时用，加安全注释
3. 禁用硬编码敏感信息→用`os.Getenv("DB_PASSWORD")`
4. 慎用`panic`→仅不可恢复错误（如初始化失败）用，业务逻辑用`error`
5. 避免嵌套循环超过3层→拆分函数（如`i/j/k`用尽前拆函数）
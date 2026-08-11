# Lộ trình học Golang 30 ngày để xây dựng hệ thống Honeypot

> Mục tiêu của lộ trình này không phải biến người học thành chuyên gia Go trong 30 ngày, mà giúp đạt mức **project-ready**: đủ khả năng tham gia xây dựng honeypot đa giao thức bằng Go, hiểu kiến trúc, viết được TCP server, xử lý đồng thời, lưu dữ liệu, xây API và kiểm thử đầu vào không tin cậy.

## 1. Mục tiêu đầu ra

Sau đúng 30 ngày, người học phải tự xây được một mini-honeypot gồm:

- TCP/Telnet fake shell, không thực thi lệnh hệ điều hành thật.
- HTTP trap server.
- Kiến trúc hướng đối tượng theo phong cách Go: struct, method, interface, composition và dependency injection.
- Session tracker an toàn khi nhiều goroutine truy cập.
- Event bus và bounded worker pool.
- MITRE rule engine và threat scoring cơ bản.
- Ghi event, session và credential vào PostgreSQL.
- REST API để truy vấn dữ liệu.
- Unit test, integration test, benchmark, fuzz test và race test.
- Graceful shutdown, timeout, connection limit và input-size limit.

Sản phẩm cuối tháng chưa cần frontend. Có thể kiểm tra API bằng curl, Postman hoặc một CLI nhỏ.

---

## 2. Thời lượng và phương pháp học

### Thời lượng đề xuất

- 30 ngày liên tục.
- 4 giờ mỗi ngày.
- Tổng thời gian: khoảng 120 giờ.
- Nếu chỉ có 2 giờ/ngày, giữ nguyên thứ tự nhưng giảm độ sâu bài tập; không bỏ phần test và concurrency.

### Công thức học mỗi ngày

| Hoạt động | Thời lượng |
|---|---:|
| Đọc lý thuyết và tài liệu chuẩn | 45 phút |
| Tự viết code | 120 phút |
| Viết test, debug và chạy công cụ kiểm tra | 45 phút |
| Ghi chú, refactor và commit | 30 phút |

### Quy tắc bắt buộc

1. Chỉ xem lý thuyết tối đa 45 phút rồi phải code.
2. Tất cả bài tập nằm trong cùng một repository để thấy kiến trúc phát triển dần.
3. Không copy đoạn code mà chưa giải thích được từng function.
4. Từ ngày 6, mọi feature mới phải có test.
5. Từ ngày 10, code concurrent phải chạy race detector.
6. Không dùng `panic` để xử lý lỗi nghiệp vụ hoặc lỗi mạng thông thường.
7. Không gọi `os/exec`, shell thật hoặc chạy payload do client gửi.
8. Mỗi ngày tạo ít nhất một commit nhỏ có thông điệp rõ ràng.

---

## 3. Chuẩn bị môi trường

### Công cụ

- Go phiên bản phù hợp với dự án.
- Git.
- IDE có Go language server.
- PostgreSQL chạy bằng Docker hoặc cài cục bộ.
- Một TCP client như `nc`, Telnet client hoặc script Go tự viết.
- curl hoặc Postman để kiểm tra REST API.

### Cấu trúc repository thực hành

```text
honeypot-lab/
├── cmd/
│   ├── sensor/
│   └── api/
├── internal/
│   ├── config/
│   ├── model/
│   ├── event/
│   ├── session/
│   ├── protocol/
│   │   ├── telnet/
│   │   └── httptrap/
│   ├── detection/
│   ├── storage/
│   └── api/
├── migrations/
├── testdata/
├── go.mod
└── README.md
```

Không cần tạo toàn bộ thư mục ngay ngày đầu. Chỉ tạo khi thực sự có code tương ứng.

---

# Tuần 1 — Nền tảng Go và OOP theo phong cách Go

## Ngày 1 — Go toolchain, package và chương trình đầu tiên

### Kiến thức

- Cấu trúc một chương trình Go.
- Package, import và exported/unexported identifier.
- Biến, hằng, zero value.
- Function và multiple return values.
- Go module.
- Các lệnh build, run, test, format và vet.

### Bài tập

1. Khởi tạo module `honeypot-lab`.
2. Viết CLI nhận ba tham số: `source-ip`, `port`, `service`.
3. Tạo một event giả và in ra terminal.
4. Tách phần domain model sang package `internal/model`.

### Sản phẩm cuối ngày

```go
type Event struct {
	ID        string
	SourceIP  string
	Service   string
	Type      string
	Timestamp time.Time
}
```

### Tự kiểm tra

- Giải thích được vì sao tên viết hoa có thể truy cập từ package khác.
- Phân biệt được compile error và runtime error.
- Repository build được từ thư mục gốc.

---

## Ngày 2 — Kiểu dữ liệu, slice, map, string và byte

### Kiến thức

- Kiểu số, boolean và string.
- Array và slice.
- Length, capacity, append và copy.
- Map và kiểm tra key tồn tại.
- String, `[]byte` và `[]rune`.
- Các package `strings`, `bytes`, `strconv`.

### Vì sao quan trọng với honeypot

Protocol server nhận dữ liệu dạng byte. Nếu không hiểu slice, buffer và encoding, parser dễ đọc sai packet, giữ tham chiếu bộ nhớ không cần thiết hoặc panic khi input bất thường.

### Bài tập

1. Parse chuỗi `wget http://example.com/a.sh` thành command và arguments.
2. Trích URL bằng parser đơn giản.
3. Dùng map đếm số event theo IP.
4. Viết function sao chép payload để tránh giữ lại buffer lớn.

### Tiêu chí hoàn thành

- Xử lý được input rỗng.
- Không truy cập slice ngoài phạm vi.
- Giải thích được vì sao `len(string)` không luôn bằng số ký tự hiển thị.

---

## Ngày 3 — OOP trong Go: struct, method và đóng gói

Go không có class và inheritance truyền thống. OOP trong Go được thể hiện bằng:

- Struct giữ trạng thái.
- Method mô tả hành vi.
- Package boundary tạo đóng gói.
- Interface tạo đa hình.
- Composition thay cho inheritance.

### Kiến thức

- Struct và nested struct.
- Method.
- Pointer receiver và value receiver.
- Constructor convention `New...`.
- Validation invariant.
- Exported và unexported field.

### Bài tập

Xây `Session` có trạng thái được đóng gói:

```go
type Session struct {
	id           string
	sourceIP     string
	startedAt    time.Time
	commandCount int
	closed       bool
}

func NewSession(id, sourceIP string, now time.Time) (*Session, error) {
	if id == "" || sourceIP == "" {
		return nil, errors.New("id and source IP are required")
	}
	return &Session{id: id, sourceIP: sourceIP, startedAt: now}, nil
}

func (s *Session) AddCommand() error {
	if s.closed {
		return errors.New("session is closed")
	}
	s.commandCount++
	return nil
}
```

### Tự kiểm tra

- Vì sao field nên để unexported?
- Khi nào dùng pointer receiver?
- Constructor phải bảo vệ invariant nào?
- Có nên viết getter/setter cho mọi field như Java không? Câu trả lời: không.

---

## Ngày 4 — Interface, đa hình và dependency inversion

### Kiến thức

- Implicit interface implementation.
- Interface nhỏ, định nghĩa ở phía sử dụng.
- Polymorphism.
- Dependency inversion.
- Mock/fake implementation trong test.

### Interface cốt lõi

```go
type Service interface {
	Name() string
	Start(ctx context.Context) error
}

type EventStore interface {
	InsertEvent(ctx context.Context, event *model.Event) error
}
```

### Bài tập

1. Tạo `TelnetService` và `HTTPService` cùng triển khai `Service`.
2. Tạo `MemoryEventStore` để test.
3. Viết function nhận `[]Service` và khởi động từng service.
4. Chứng minh code gọi không cần biết concrete type.

### SOLID áp dụng trong Go

- **S:** mỗi type có một trách nhiệm chính.
- **O:** thêm service mới bằng cách triển khai interface, hạn chế sửa orchestration.
- **L:** mọi implementation phải tuân thủ contract của interface.
- **I:** ưu tiên interface 1–3 method.
- **D:** processor phụ thuộc `EventStore`, không phụ thuộc trực tiếp PostgreSQL store.

### Anti-pattern cần tránh

- Interface rất lớn chứa hàng chục method.
- Tạo interface cho mọi struct dù chỉ có một implementation và không cần test double.
- Giả lập inheritance bằng một `BaseService` khổng lồ.
- Lạm dụng `any`/`interface{}` thay cho kiểu cụ thể.

---

## Ngày 5 — Composition, strategy và factory vừa đủ

### Kiến thức

- Struct embedding.
- Composition over inheritance.
- Strategy pattern bằng interface/function.
- Factory function.
- Functional option chỉ dùng khi constructor có nhiều tùy chọn thực sự.

### Bài tập

SSH và Telnet phải dùng chung shell logic:

```go
type CommandHandler interface {
	Execute(ctx context.Context, session *Session, args []string) CommandResult
}

type Shell struct {
	handlers map[string]CommandHandler
}
```

Viết ba handler:

- `WhoamiHandler`
- `ListHandler`
- `CatHandler`

### Tiêu chí hoàn thành

- Thêm command mới không phải sửa một switch dài.
- Shell không phụ thuộc SSH hoặc Telnet transport.
- Command handler không thể chạy lệnh hệ điều hành thật.

---

## Ngày 6 — Error handling và unit test

### Kiến thức

- Kiểu `error`.
- Error wrapping với `%w`.
- `errors.Is` và `errors.As`.
- Sentinel error và custom error.
- `defer`.
- Table-driven test và subtest.
- Test helper.

### Bài tập

1. Viết port parser.
2. Phân biệt port không hợp lệ, cấu hình thiếu và bind thất bại.
3. Test constructor của `Session`.
4. Test command handlers bằng bảng dữ liệu.
5. Đạt tối thiểu 20 test case.

### Không nên làm

```go
if err != nil {
	panic(err)
}
```

Server phải log lỗi, đóng đúng tài nguyên và tiếp tục phục vụ connection khác khi có thể.

---

## Ngày 7 — JSON, file và mini-project OOP

### Kiến thức

- `encoding/json`.
- JSON tag.
- `json.RawMessage`.
- `time.Time` và duration.
- Buffered file I/O.

### Mini-project tuần 1

Xây công cụ:

```text
JSONL event file
    → EventReader
    → Classifier strategies
    → Aggregator
    → Terminal report
```

### Yêu cầu

- `EventReader` chỉ chịu trách nhiệm đọc và decode.
- `Classifier` là interface.
- `Aggregator` quản lý thống kê.
- Không dùng global mutable state.
- Có tối thiểu 20 test case.
- Có README và ví dụ input/output.

### Checkpoint tuần 1

Bạn phải tự giải thích được:

- Go thực hiện OOP mà không có class như thế nào?
- Interface khác inheritance ra sao?
- Vì sao composition phù hợp với SSH/Telnet dùng chung fake shell?
- Vì sao interface nên nhỏ?
- Khi nào không nên tạo abstraction?

---

# Tuần 2 — Concurrency, networking và protocol

## Ngày 8 — Goroutine và lifecycle

### Kiến thức

- Khởi tạo goroutine.
- Closure capture.
- `sync.WaitGroup`.
- Goroutine lifecycle và leak.
- Ownership của tài nguyên.

### Bài tập

1. Xử lý 100 job đồng thời.
2. Chờ tất cả job hoàn tất.
3. Tạo có chủ đích một goroutine leak rồi sửa.
4. Ghi lại số goroutine trước và sau test.

---

## Ngày 9 — Channel và event pipeline

### Kiến thức

- Buffered/unbuffered channel.
- Send, receive và close.
- `select`.
- Fan-in và fan-out.
- Backpressure.

### Bài tập

```text
Event producers → buffered channel → workers → result channel
```

Phải xác định rõ:

- Thành phần nào tạo channel?
- Thành phần nào được quyền đóng channel?
- Buffer đầy thì block, drop hay ghi durable queue?
- Làm sao đo event bị drop?

---

## Ngày 10 — Mutex, session tracker và race detector

### Kiến thức

- `sync.Mutex` và `sync.RWMutex`.
- Map và concurrent access.
- Atomic operation.
- Race condition.

### Bài tập

Viết `SessionTracker`:

- `Start`
- `Get`
- `End`
- `Count`
- `AddCommand`
- `AddTechnique`
- `AddScore`

Viết test cho hàng trăm goroutine truy cập đồng thời và chạy race detector.

### Tiêu chí hoàn thành

- Không có data race.
- Không giữ mutex khi gọi callback bên ngoài.
- Không trả con trỏ cho phép caller sửa state mà không khóa.

---

## Ngày 11 — Context, timeout và graceful shutdown

### Kiến thức

- `context.Context`.
- Cancellation và deadline.
- Context propagation.
- OS signal.
- Graceful shutdown.

### Bài tập

1. Worker dừng khi context bị hủy.
2. Server ngừng accept connection mới.
3. Connection hiện tại có thời gian hoàn thành giới hạn.
4. Shutdown không để goroutine bị treo.

### Quy tắc

- Context là tham số đầu tiên.
- Không lưu context dài hạn trong struct nếu không có lý do rõ ràng.
- Không tự tạo `context.Background()` ở tầng dưới để bỏ qua cancellation.

---

## Ngày 12 — TCP server an toàn

### Kiến thức

- `net.Listen` và `Accept`.
- `net.Conn`.
- Remote address.
- Read/write deadline.
- `bufio.Reader`.
- Connection semaphore.
- Slow client và oversized input.

### Bài tập

Viết TCP echo server có:

- Mỗi connection một goroutine.
- Giới hạn 100 connection đồng thời.
- Idle timeout.
- Dòng đầu vào tối đa 4 KB.
- Structured log gồm source IP và port.
- Graceful shutdown.

---

## Ngày 13 — Text protocol và State pattern

### Kiến thức OOP

- State machine.
- State được biểu diễn bằng enum hoặc các handler riêng.
- Transition validation.
- Tránh boolean state chồng chéo.

### Bài tập

Viết Telnet-like server:

```text
CONNECTED → USERNAME → PASSWORD → SHELL → CLOSED
```

Hỗ trợ:

- Accept mọi credential trong môi trường lab.
- `whoami`, `pwd`, `ls`, `cat /etc/passwd`, `exit`.
- Ghi event connect, auth, command và disconnect.
- Không chạy command thật.

### Test

- PASS trước USER bị từ chối.
- Input quá dài bị đóng connection.
- Client ngắt giữa chừng không làm leak session.
- Hai client có state độc lập.

---

## Ngày 14 — Binary protocol

### Kiến thức

- Little/big endian.
- Packet length và sequence.
- `encoding/binary`.
- `io.ReadFull`.
- Partial read.
- Input-size validation trước khi cấp phát.

### Bài tập

Viết encoder/decoder packet có header:

```text
3 byte payload length
1 byte sequence number
N byte payload
```

Đây là bài tập chuẩn bị cho MySQL wire protocol.

### Test

- Packet hợp lệ.
- Header thiếu.
- Length bằng 0.
- Length lớn hơn giới hạn.
- Payload bị cắt.
- Nhiều packet liên tiếp.

### Checkpoint tuần 2

- TCP fake shell phục vụ đồng thời nhiều client.
- Không có data race.
- Shutdown sạch.
- Parser không panic với packet lỗi cơ bản.
- State của client này không ảnh hưởng client khác.

---

# Tuần 3 — Kiến trúc backend honeypot

## Ngày 15 — Domain-driven model vừa đủ

### Kiến thức

- Entity và value object ở mức thực dụng.
- Invariant.
- Domain behavior thay vì struct chỉ chứa dữ liệu.
- Không đưa SQL/HTTP concern vào domain model.

### Model cần có

- `Event`
- `Session`
- `Credential`
- `CapturedFile`
- `IOC`
- `MITREDetection`
- `Sensor`

### Quy định dữ liệu

- Timestamp sử dụng UTC.
- Có `SchemaVersion`.
- Có `SensorID`, `SessionID`, `EventID`.
- `ServiceData` chứa dữ liệu riêng từng giao thức.
- Password không xuất hiện trong log thông thường.

---

## Ngày 16 — Event bus và Observer pattern

### Kiến thức OOP

- Observer/pub-sub.
- Interface segregation.
- Subscriber ownership.
- Event immutability sau publish.

### Chức năng

- Subscribe theo topic.
- Wildcard subscription.
- Buffered channel.
- Unsubscribe.
- Shutdown.
- Metric đếm event drop.

### Test

- Một event tới nhiều subscriber.
- Filter topic.
- Subscriber chậm.
- Concurrent subscribe/publish.
- Unsubscribe không double-close channel.

---

## Ngày 17 — Worker pool và Pipeline pattern

### Pipeline

```text
Receive → Validate → Enrich → Detect → Persist → Stream
```

### Thiết kế interface

```go
type Enricher interface {
	Enrich(ctx context.Context, event *model.Event) error
}

type Detector interface {
	Detect(ctx context.Context, event *model.Event) ([]model.Detection, error)
}
```

### Yêu cầu

- Worker count cấu hình được.
- Một event lỗi không làm dừng toàn pipeline.
- Có timeout cho dependency ngoài.
- Có metrics success/error/latency.
- Không dùng interface khổng lồ cho mọi stage.

---

## Ngày 18 — PostgreSQL và Repository pattern

### Kiến thức

- Connection pool.
- Parameterized query.
- Transaction.
- Context timeout.
- Scan row.
- Batch insert.
- Unique constraint và upsert.

### Repository interface

```go
type EventRepository interface {
	Insert(ctx context.Context, event *model.Event) error
	Find(ctx context.Context, filter EventFilter) ([]model.Event, error)
}
```

### Bài tập

- Migration tạo bảng events và sessions.
- Insert event.
- Truy vấn theo IP/service/time.
- Pagination.
- Test duplicate event ID.

### OOP cần nhớ

- Domain không import PostgreSQL driver.
- Repository concrete implementation nằm ở storage package.
- Interface nằm gần component sử dụng, không cần một package `interfaces` chung.

---

## Ngày 19 — Persistence workflow và Unit of Work vừa đủ

### Bài tập

- Insert session lúc kết nối.
- Update session khi disconnect.
- Insert credential attempt.
- Upsert attacker.
- Dùng transaction khi nhiều thay đổi phải thành công cùng nhau.
- Dùng batch insert cho nhiều event.

### Tình huống kiểm thử

- Database timeout.
- Duplicate event.
- Transaction rollback.
- Context bị cancel.
- Connection pool cạn.

Không cần tạo một framework Unit of Work tổng quát; chỉ viết transaction orchestration cho use case thực sự cần.

---

## Ngày 20 — REST API và Controller/Service boundary

### Kiến thức

- `net/http` hoặc router nhẹ.
- Handler/controller.
- Application service.
- Validation.
- JSON response.
- Pagination.
- Middleware.
- Panic recovery.

### API tối thiểu

```text
GET /health
GET /events
GET /sessions
GET /sessions/{id}
GET /stats
```

### Boundary

```text
HTTP handler
    → validate/translate request
Application service
    → thực hiện use case
Repository
    → truy cập database
```

Handler không nên chứa SQL hoặc toàn bộ business logic.

---

## Ngày 21 — Logging, metrics và checkpoint kiến trúc

### Structured log cần có

- `event_id`
- `session_id`
- `sensor_id`
- `service`
- `source_ip`
- `error`

### Metrics cần có

- Active connections.
- Events received/processed/failed/dropped.
- Processing latency.
- Database error.
- Session count.

### Checkpoint tuần 3

Luồng sau phải chạy được:

```text
Fake service
  → event bus
  → worker pool
  → in-memory/PostgreSQL repository
  → REST API
```

Bạn phải vẽ và giải thích dependency direction. Nếu domain package import API hoặc PostgreSQL package, cần sửa kiến trúc.

---

# Tuần 4 — Honeypot chuyên sâu, detection và kiểm thử an toàn

## Ngày 22 — Fake filesystem và Command pattern

### Kiến thức OOP

- Command pattern.
- Registry/factory.
- Composition giữa shell, filesystem và recorder.
- Separation of concerns.

### Thành phần

```text
FakeFS
CommandParser
CommandRegistry
Shell
Recorder
```

### Command tối thiểu

- `ls`, `cd`, `pwd`
- `cat`, `head`, `tail`
- `whoami`, `id`, `uname`
- `ps`, `netstat`, `ifconfig`
- `wget`, `curl`
- `crontab`
- `exit`

### An toàn

- Không dùng `os/exec`.
- Không đọc filesystem thật.
- Path traversal không thoát khỏi `FakeFS`.
- Output download chỉ là mô phỏng.

---

## Ngày 23 — Telnet IAC parser

### Kiến thức

- IAC byte.
- `WILL`, `WONT`, `DO`, `DONT`.
- Escaped `0xFF`.
- CRLF.
- Parser state.

### Bài tập

- Viết parser tách IAC sequence khỏi user data.
- Thương lượng ECHO và SUPPRESS-GO-AHEAD cơ bản.
- Tích hợp với fake shell ngày 22.

### Test

- IAC sequence hợp lệ.
- Sequence bị cắt giữa packet.
- Nhiều option liên tiếp.
- Escaped IAC.
- Random byte input không panic.

---

## Ngày 24 — SMTP hoặc FTP state machine

Chọn một protocol để làm sâu; protocol còn lại sẽ thực hiện trong dự án sáu tháng.

### SMTP

```text
CONNECTED → EHLO → MAIL → RCPT → DATA → SAVED
```

### FTP

```text
CONNECTED → USER → PASS → AUTHENTICATED → PASV/STOR → CLOSED
```

### Yêu cầu chung

- Validate transition.
- Giới hạn line length và payload size.
- Timeout mỗi phase.
- Capture event nhưng không thực hiện hành động bên ngoài.
- Unit test cho invalid order.

---

## Ngày 25 — MITRE rule engine và Strategy pattern

### Detection tối thiểu

- Shell command → `T1059.004`.
- System discovery → `T1082`.
- File discovery → `T1083`.
- Tool transfer → `T1105`.
- Cron persistence → `T1053.003`.
- Mining → `T1496`.
- Năm auth attempt trong năm phút → `T1110`.
- Ba service trong 60 giây → `T1046`.

### Thiết kế

```go
type Rule interface {
	ID() string
	Evaluate(event *model.Event, state DetectionState) bool
}
```

Có thể dùng function strategy để giảm boilerplate nếu rule đơn giản.

### Test

- Single-event rule.
- Sliding window boundary.
- Deduplicate technique.
- Event đến ngoài thứ tự.
- Hai IP có state tách biệt.

---

## Ngày 26 — IOC extraction và threat scoring

### IOC

- IPv4/IPv6.
- URL.
- Domain.
- Email.
- SHA-256.
- User-agent.

### Threat scoring

- Auth attempts: có giới hạn đóng góp.
- Discovery.
- Tool transfer.
- Persistence.
- Unique MITRE technique.
- Session duration.
- Mining/destructive behavior.

### Yêu cầu

- Điểm luôn trong `[0,100]`.
- Technique lặp không cộng điểm vô hạn.
- Rule weight cấu hình bằng dữ liệu, không hard-code rải rác.
- Có test boundary 0, 100 và duplicate event.

---

## Ngày 27 — Fuzz testing parser

### Mục tiêu

Fuzz các thành phần nhận dữ liệu không tin cậy:

- Telnet IAC parser.
- Command parser.
- Binary packet decoder.
- SMTP/FTP line parser.
- IOC extractor.

### Điều kiện thành công

- Không panic.
- Không vòng lặp vô hạn.
- Không cấp phát bộ nhớ dựa trên length chưa kiểm tra.
- Không vượt giới hạn thời gian bất thường.
- Corpus chứa các packet hợp lệ và malformed tiêu biểu.

---

## Ngày 28 — Benchmark, load và resilience

### Benchmark

- Event bus publish.
- Worker processing.
- Command dispatch.
- MITRE detector.
- Batch insert.

### Kịch bản resilience

- 1.000 TCP client.
- Database lỗi tạm thời.
- Subscriber chậm.
- Context bị hủy.
- Client gửi rất chậm.
- Client gửi length field cực lớn.
- Connection đóng trong lúc server đang write.

### Kết quả cần ghi

- Throughput.
- p50/p95/p99 latency.
- Event drop count.
- Goroutine trước/sau test.
- Memory allocation cho mỗi event.

---

## Ngày 29 — Capstone mini-honeypot

### Kiến trúc

```text
Telnet + HTTP Trap
        ↓
     Event Bus
        ↓
  4 Processor Workers
        ↓
MITRE + IOC + Threat Score
        ↓
     PostgreSQL
        ↓
      REST API
```

### Yêu cầu chức năng

- Telnet login và fake shell.
- HTTP ghi request và user-agent.
- Session lifecycle.
- Event persistence.
- MITRE detection.
- IOC extraction.
- Threat score.
- Truy vấn event/session qua API.

### Yêu cầu phi chức năng

- Graceful shutdown.
- Timeout.
- Connection limit.
- Input-size limit.
- Structured logging.
- Không data race.
- Không chạy command thật.
- Có unit và integration test.

---

## Ngày 30 — Refactor, đánh giá OOP và bảo vệ sản phẩm

### Refactor checklist

- Interface có thật sự cần thiết không?
- Interface có quá lớn không?
- Có type nào làm nhiều trách nhiệm không?
- Có logic protocol bị lẫn với database không?
- SSH/Telnet shell có thể dùng chung không?
- Handler có phụ thuộc concrete repository không?
- Có callback nào được gọi khi đang giữ mutex không?
- Có code lặp đáng trích xuất nhưng không tạo abstraction quá mức không?

### Mười câu phải trả lời được

1. Một connection đi qua hệ thống như thế nào?
2. Goroutine nào được tạo, sở hữu và kết thúc khi nào?
3. Buffer đầy thì hệ thống xử lý ra sao?
4. Session tracker chống data race bằng cách nào?
5. Struct, interface và composition thể hiện OOP ra sao?
6. Vì sao không dùng inheritance như Java/C++?
7. Parser chống input độc hại thế nào?
8. Database lỗi thì honeypot còn phản hồi client không?
9. Làm sao test 1.000 connection và phát hiện goroutine leak?
10. Vì sao fake shell không thể escape ra hệ điều hành thật?

### Definition of Done ngày 30

- Build thành công.
- Unit test thành công.
- Integration test thành công.
- Race detector không báo lỗi.
- Fuzz smoke test không panic.
- API trả dữ liệu đúng.
- README có sơ đồ kiến trúc và hướng dẫn chạy.
- Có báo cáo benchmark ngắn.
- Có video hoặc kịch bản demo 3–5 phút.

---

# 4. Các design pattern nên biết cho dự án Honeypot

| Pattern | Ứng dụng |
|---|---|
| Strategy | MITRE rule, threat scoring, exporter |
| Observer/Pub-Sub | Event bus và realtime consumer |
| Command | Fake shell command handlers |
| State | FTP, SMTP, Telnet và session lifecycle |
| Repository | Tách domain/application khỏi PostgreSQL |
| Factory | Tạo service theo cấu hình |
| Adapter | SIEM exporter, GeoIP provider, gRPC transport |
| Pipeline | Validate → enrich → detect → persist |
| Decorator/Middleware | Logging, timeout, rate limit |

Không cần ép mọi pattern vào code. Chỉ dùng khi pattern giải quyết một vấn đề đang tồn tại.

---

# 5. OOP trong Go: những điều phải hiểu đúng

## Go có OOP nhưng không có class truyền thống

Go hỗ trợ các đặc điểm cốt lõi:

- Encapsulation qua package và field unexported.
- Abstraction qua interface.
- Polymorphism qua implicit interface implementation.
- Composition qua embedding và chứa dependency trong struct.

Go không cung cấp:

- Class keyword.
- Inheritance hierarchy.
- Method overloading.
- Abstract class.

## Composition thay inheritance

Không thiết kế:

```text
BaseService
  ├── SSHService
  ├── TelnetService
  └── FTPService
```

Nên thiết kế:

```text
SSH transport ──┐
                ├── Shared Shell ── FakeFS ── Command Registry
Telnet transport┘
```

## Interface ở phía sử dụng

Processor chỉ cần:

```go
type Store interface {
	InsertEvent(context.Context, *model.Event) error
}
```

Nó không cần biết store còn có 30 method khác. Interface nhỏ giúp mock dễ và giảm coupling.

## Dependency injection không cần framework

```go
processor := event.NewProcessor(store, detector, geoResolver, logger)
```

Truyền dependency qua constructor là đủ. Không cần DI container.

---

# 6. Kiến thức tạm thời chưa cần học

- Generics nâng cao.
- Reflection nâng cao.
- `unsafe`.
- CGO.
- Compiler internals.
- Kubernetes.
- Distributed consensus.
- Tự triển khai mật mã SSH.
- Full SMB2 stack.
- Frontend chuyên sâu.
- Design pattern kiểu Java áp dụng máy móc.

Ưu tiên:

```text
Go fundamentals
→ OOP theo phong cách Go
→ concurrency
→ networking
→ state machine/parser
→ database/API
→ testing/security
```

---

# 7. Checklist tốt nghiệp lộ trình

## Go cơ bản

- [ ] Hiểu package, module và visibility.
- [ ] Dùng thành thạo slice, map, string và byte.
- [ ] Xử lý error có context.
- [ ] Không lạm dụng panic/global state.

## OOP và kiến trúc

- [ ] Dùng struct và method để đóng gói behavior.
- [ ] Dùng interface nhỏ để tạo đa hình.
- [ ] Dùng composition thay inheritance.
- [ ] Áp dụng dependency inversion và constructor injection.
- [ ] Biết khi nào không nên tạo interface/pattern.
- [ ] Tách protocol, domain, storage và transport.

## Concurrency

- [ ] Tự viết worker pool.
- [ ] Hiểu ownership và close channel.
- [ ] Dùng context để shutdown.
- [ ] Session tracker không có data race.
- [ ] Không để goroutine leak.

## Networking và security

- [ ] Tự viết TCP server.
- [ ] Có read/write deadline.
- [ ] Có connection và input-size limit.
- [ ] Parser không panic khi fuzz.
- [ ] Fake shell không chạy command thật.

## Backend

- [ ] Ghi và đọc PostgreSQL an toàn.
- [ ] Dùng transaction/upsert đúng chỗ.
- [ ] Viết REST API có validation và pagination.
- [ ] Event pipeline hoạt động đồng thời.
- [ ] Có structured logs và metrics.

## Honeypot

- [ ] Hiểu state machine của ít nhất hai protocol.
- [ ] Viết được MITRE single-event và sliding-window rule.
- [ ] Trích được IOC.
- [ ] Tính threat score 0–100.
- [ ] Hoàn thành capstone ngày 29.

---

# 8. Kết quả kỳ vọng thực tế

Sau 30 ngày, bạn chưa phải Go master theo nghĩa chuyên gia. Bạn đạt mức:

> **Go backend/networking developer ở cấp project-ready, đủ khả năng tham gia xây dựng honeypot đa giao thức dưới quy trình code review.**

Giai đoạn sáu tháng làm đồ án chính sẽ giúp nâng tiếp các năng lực:

- Protocol emulation sâu hơn.
- Multi-sensor gRPC.
- Zero-loss buffering và deduplication.
- SIEM integration.
- Machine learning anomaly detection.
- Performance tuning tới tải lớn.
- Production hardening.

Điều quan trọng nhất của tháng đầu không phải số dòng code, mà là khả năng tự giải thích kiến trúc, kiểm soát concurrency, xử lý input độc hại và viết test chứng minh hệ thống hoạt động đúng.

# Day 4 — Interface, đa hình và dependency inversion trong 2 giờ

Day 4 nối trực tiếp từ Day 3. Ở Day 3, `Session` tự bảo vệ state bằng struct, method và package boundary. Hôm nay ta thêm hai vấn đề mới:

1. Hệ thống có nhiều loại service nhưng orchestration không nên biết chi tiết từng loại.
2. Processor cần lưu event nhưng không nên bị khóa cứng vào PostgreSQL hay một database cụ thể.

Ta vẫn dùng nhịp học của Day 3:

```text
Vấn đề → contract nhỏ → code/test → chạy → giải thích → checkpoint
```

Không đọc và chép một lời giải hoàn chỉnh từ đầu. Hãy hoàn thành từng checkpoint rồi mới chuyển sang chặng kế tiếp.

> Phạm vi 2 giờ chỉ mô phỏng việc khởi động service và ghi một startup event. Ta không mở cổng Telnet/HTTP thật. Một network server thật thường block trong `Start`; khi đó cần thiết kế concurrency và shutdown riêng, không thuộc Day 4.

---

## 1. Đích đến sau 120 phút

Bạn hoàn thành Day 4 khi chứng minh được các hành vi sau:

1. `TelnetService` và `HTTPService` cùng dùng được ở vị trí yêu cầu `Service` dù không có từ khóa `implements`.
2. `StartServices` nhận `[]Service`, khởi động lần lượt và không dùng type switch.
3. `Processor` chỉ phụ thuộc `EventStore`, không import một store cụ thể.
4. `MemoryEventStore` lưu event trong RAM và dùng được như fake trong test.
5. Test chứng minh orchestration không cần biết concrete type.
6. Thêm một service mới không buộc sửa `StartServices`.

Luồng cuối buổi:

```text
main
  ├─ tạo MemoryEventStore
  ├─ inject store vào Processor
  ├─ tạo TelnetService và HTTPService
  └─ đưa cả hai vào []Service
          ↓
     StartServices
          ↓
   interface dispatch
      ├─ TelnetService.Start
      └─ HTTPService.Start
          ↓
       Processor
          ↓
      EventStore interface
          ↓
   MemoryEventStore
```

### Lịch 120 phút

| Thời gian | Chặng | Kết quả bắt buộc |
|---:|---|---|
| 0–10 | Khởi động project | Module và package rỗng build được |
| 10–30 | `Service` và implicit implementation | Hai concrete service thỏa interface |
| 30–50 | Đa hình trong orchestration | `StartServices([]Service)` có test |
| 50–55 | Nghỉ và tự giải thích | Nói được interface thuộc phía nào |
| 55–75 | Dependency inversion | `Processor` phụ thuộc `EventStore` |
| 75–95 | Fake trong test | Memory store và processor test xanh |
| 95–110 | Ghép end-to-end | CLI chạy hai service qua interface |
| 110–116 | Chứng minh OCP | Không sửa runner khi thêm implementation |
| 116–120 | Quality gate | `fmt`, `test`, `vet`, `build` đều xanh |

Nếu một checkpoint chưa xanh, không chuyển chặng. Cắt phần mở rộng trước, không cắt test cốt lõi.

---

## 2. Chặng 0 — Khởi động project (0–10 phút)

### 2.1. Tạo module

Trong PowerShell tại `C:\Go_project`, tự tạo cấu trúc sau:

```text
go_day4/
├── cmd/
│   └── servicelab/
├── internal/
│   ├── model/
│   ├── orchestration/
│   ├── processor/
│   ├── service/
│   └── store/
└── go.mod
```

Tạo các thư mục và khởi tạo module:

```powershell
New-Item -ItemType Directory -Force -Path C:\Go_project\go_day4\cmd\servicelab
New-Item -ItemType Directory -Force -Path C:\Go_project\go_day4\internal\model
New-Item -ItemType Directory -Force -Path C:\Go_project\go_day4\internal\orchestration
New-Item -ItemType Directory -Force -Path C:\Go_project\go_day4\internal\processor
New-Item -ItemType Directory -Force -Path C:\Go_project\go_day4\internal\service
New-Item -ItemType Directory -Force -Path C:\Go_project\go_day4\internal\store
Set-Location C:\Go_project\go_day4
go mod init honeypot-day4
go list -m
```

Nếu `go.mod` đã tồn tại, không chạy `go mod init` lần nữa. `go list -m` phải in:

```text
honeypot-day4
```

Tạo `cmd/servicelab/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("servicelab: ready")
}
```

Chạy:

```powershell
go run ./cmd/servicelab
go test ./...
go build ./...
```

### 2.2. Quy ước domain nhỏ

Tạo `internal/model/event.go`:

```go
package model

type Event struct {
	Service string
	Message string
}
```

Day 4 không cần ID, timestamp, JSON hay database schema. Chúng không giúp hiểu interface và sẽ làm loãng mục tiêu 2 giờ.

### Checkpoint 0

- Module là `honeypot-day4`.
- CLI in `servicelab: ready`.
- `go test ./...` và `go build ./...` thành công.

---

## 3. Chặng 1 — Một contract, hai implementation (10–30 phút)

### 3.1. Bắt đầu từ phía sử dụng

Ai cần điều phối service? Package `orchestration`. Vì vậy hãy đặt interface trong `internal/orchestration`, không đặt nó vào package `service` chỉ vì các concrete type nằm ở đó.

Tạo `internal/orchestration/orchestration.go`:

```go
package orchestration

import "context"

type Service interface {
	Name() string
	Start(ctx context.Context) error
}
```

Đây là interface tốt cho bài học vì:

- Nó có đúng hai method mà orchestration cần.
- Nó không chứa cấu hình, logging, health check hay persistence.
- Nó mô tả khả năng, không mô tả cây kế thừa.

### 3.2. Tạo hai concrete type tối thiểu

Tạo `internal/service/telnet.go`:

```go
package service

import (
	"context"

	"honeypot-day4/internal/orchestration"
)

type TelnetService struct{}

func (s *TelnetService) Name() string {
	return "telnet"
}

func (s *TelnetService) Start(ctx context.Context) error {
	return nil
}

var _ orchestration.Service = (*TelnetService)(nil)
```

Tạo `internal/service/http.go`:

```go
package service

import (
	"context"

	"honeypot-day4/internal/orchestration"
)

type HTTPService struct{}

func (s *HTTPService) Name() string {
	return "http"
}

func (s *HTTPService) Start(ctx context.Context) error {
	return nil
}

var _ orchestration.Service = (*HTTPService)(nil)
```

Không viết `implements Service`; Go không có cú pháp đó. Method set quyết định type có thỏa interface hay không.

### 3.3. Nhờ compiler chứng minh contract

Hai file vừa tạo đều có compile-time assertion dạng:

```go
var _ orchestration.Service = (*TelnetService)(nil)
```

Viết assertion tương tự cho HTTP. Dấu `_` bỏ giá trị đi; mục đích của dòng này chỉ là yêu cầu compiler kiểm tra phép gán.

Thí nghiệm trong 2 phút:

1. Tạm đổi tên `HTTPService.Start` thành `Run`.
2. Chạy `go test ./...`.
3. Đọc lỗi: `*HTTPService does not implement orchestration.Service`.
4. Đổi lại thành `Start` và chạy test lần nữa.

Compiler vừa chứng minh implicit interface implementation. Không có đăng ký, inheritance hay annotation ẩn phía sau.

### 3.4. Pointer receiver hay value receiver?

Hãy dùng pointer receiver cho hai service. Hai dòng dưới đây chỉ minh họa chữ ký method đã có trong `telnet.go`; **không tạo file test và không chép riêng hai dòng này vào source code**:

```go
func (s *TelnetService) Name() string
func (s *TelnetService) Start(ctx context.Context) error
```

Hiện tại struct rỗng nên value receiver cũng chạy. Ta chọn pointer receiver vì chặng sau service sẽ giữ dependency là processor. Quan trọng hơn, cả hai method nên có receiver nhất quán để method set dễ dự đoán.

Quy tắc method set cần nhớ:

- Method có receiver `T` thuộc method set của cả `T` và `*T`.
- Method có receiver `*T` chỉ thuộc method set của `*T`.
- Vì hai service dùng pointer receiver, `*TelnetService` và `*HTTPService` thỏa `Service`; value không phải pointer của chúng thì không.

### Checkpoint 1

Chạy:

```powershell
go fmt ./internal/service ./internal/orchestration
go test ./...
```

Bạn phải trả lời được:

- Type thỏa interface từ lúc nào? Khi method set khớp contract, không phải khi assertion xuất hiện.
- Assertion có bắt buộc không? Không; nó chỉ làm tài liệu có kiểm tra bởi compiler.
- Interface đặt ở đâu? Cạnh consumer là orchestration.

---

## 4. Chặng 2 — Đa hình trong orchestration (30–50 phút)

### 4.1. Viết function chỉ biết `Service`

Thay toàn bộ `internal/orchestration/orchestration.go` bằng:

```go
package orchestration

import (
	"context"
	"fmt"
)

type Service interface {
	Name() string
	Start(ctx context.Context) error
}

func StartServices(ctx context.Context, services []Service) error {
	for _, service := range services {
		if err := service.Start(ctx); err != nil {
			return fmt.Errorf("start %s: %w", service.Name(), err)
		}
	}
	return nil
}
```

Contract của bài:

1. Slice rỗng trả `nil`.
2. Service được gọi đúng thứ tự trong slice.
3. Gặp lỗi đầu tiên thì dừng.
4. Lỗi trả về có dạng `start <name>: <cause>`.
5. Cause vẫn kiểm tra được bằng `errors.Is`, nên dùng `%w`, không dùng `%v` khi wrap.

Không được dùng:

```go
switch s := service.(type) { ... }
```

Nếu cần type switch để gọi `Start`, abstraction đã thất bại.

### 4.2. Tạo test double nhỏ ngay trong test

Tạo `internal/orchestration/orchestration_test.go` với toàn bộ nội dung:

```go
package orchestration_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"honeypot-day4/internal/orchestration"
)

type fakeService struct {
	name    string
	started *[]string
	err     error
}

func (s fakeService) Name() string {
	return s.name
}

func (s fakeService) Start(context.Context) error {
	*s.started = append(*s.started, s.name)
	return s.err
}

func TestStartServicesUsesOnlyServiceContract(t *testing.T) {
	var started []string
	services := []orchestration.Service{
		fakeService{name: "telnet", started: &started},
		fakeService{name: "http", started: &started},
	}

	if err := orchestration.StartServices(context.Background(), services); err != nil {
		t.Fatalf("StartServices() error = %v", err)
	}

	want := []string{"telnet", "http"}
	if !reflect.DeepEqual(started, want) {
		t.Fatalf("started = %v, want %v", started, want)
	}
}

func TestStartServicesStopsAndWrapsError(t *testing.T) {
	var started []string
	startErr := errors.New("port unavailable")
	services := []orchestration.Service{
		fakeService{name: "telnet", started: &started, err: startErr},
		fakeService{name: "http", started: &started},
	}

	err := orchestration.StartServices(context.Background(), services)
	if !errors.Is(err, startErr) {
		t.Fatalf("StartServices() error = %v, want wrapped %v", err, startErr)
	}
	if got, want := err.Error(), "start telnet: port unavailable"; got != want {
		t.Fatalf("StartServices() error = %q, want %q", got, want)
	}

	want := []string{"telnet"}
	if !reflect.DeepEqual(started, want) {
		t.Fatalf("started = %v, want %v", started, want)
	}
}
```

Test đầu chứng minh runner chỉ dùng contract. Test sau chứng minh runner dừng ở lỗi đầu tiên và giữ nguyên cause qua `%w`.

Chạy riêng package:

```powershell
go test -count=1 ./internal/orchestration -v
```

### 4.3. Đây là đa hình ở điểm nào?

Biến trong vòng lặp có static type là `Service`. Giá trị thật bên trong có thể là `*TelnetService`, `*HTTPService` hoặc `fakeService`. Khi gọi `service.Start(ctx)`, Go dispatch tới method của concrete value đang nằm trong interface.

Caller chỉ biết hai điều:

```text
Name() string
Start(context.Context) error
```

Caller không biết field, constructor, protocol hay cách concrete service tạo event.

### Checkpoint 2

- Hai test orchestration xanh.
- `StartServices` không import package concrete `service`.
- `StartServices` không có type switch, reflection hoặc `any`.
- Bạn giải thích được static interface type và dynamic concrete value.

---

## 5. Nghỉ 5 phút và tự kiểm tra (50–55 phút)

Không code trong 5 phút này. Tự trả lời bằng lời:

1. Vì sao `Service` nằm trong `orchestration`?
2. Nếu thêm `SSHService`, file runner có phải sửa không?
3. Fake service giúp test orchestration mà không khởi động mạng thật như thế nào?

Nếu câu 2 là “phải thêm case vào switch”, hãy quay lại Chặng 2 và bỏ type switch.

---

## 6. Chặng 3 — Đảo chiều dependency của processor (55–75 phút)

### 6.1. Nhận diện dependency sai

Thiết kế cần tránh:

```text
Processor → PostgresEventStore cụ thể
```

Hậu quả:

- Unit test processor cần database hoặc phải biết chi tiết PostgreSQL.
- Đổi nơi lưu event buộc sửa processor.
- Business flow và hạ tầng bị ghép vào nhau.

Ta muốn:

```text
Processor → EventStore contract ← Memory/PostgreSQL implementation
```

Mũi tên source-code dependency của processor hướng vào abstraction mà chính processor sở hữu.

### 6.2. Đặt interface cạnh consumer

Tạo `internal/processor/processor.go`. Trước hết chỉ viết interface:

```go
package processor

import (
	"context"

	"honeypot-day4/internal/model"
)

type EventStore interface {
	InsertEvent(ctx context.Context, event *model.Event) error
}
```

Chỉ một method vì processor chỉ cần một operation. Đừng thêm `Update`, `Delete`, `List`, `Connect`, `Close` để “phòng tương lai”.

### 6.3. Inject dependency

Hoàn thiện file `internal/processor/processor.go` thành:

```go
package processor

import (
	"context"
	"errors"

	"honeypot-day4/internal/model"
)

type EventStore interface {
	InsertEvent(ctx context.Context, event *model.Event) error
}

type Processor struct {
	store EventStore
}

func New(store EventStore) *Processor {
	return &Processor{store: store}
}

func (p *Processor) Process(ctx context.Context, event *model.Event) error {
	if event == nil {
		return errors.New("event is required")
	}

	return p.store.InsertEvent(ctx, event)
}
```

Code vừa viết giữ đúng contract:

1. `New` lưu dependency caller truyền vào.
2. `Process` từ chối `nil` bằng lỗi `event is required`.
3. Event hợp lệ được chuyển nguyên vẹn cho `InsertEvent`.
4. Lỗi từ store được trả về cho caller, không bị nuốt.

### 6.4. Kiểm tra import direction

Mở `processor.go` và kiểm tra import. File này chỉ import:

- `context`.
- package `model`.
- `errors`.

Nó không được import package `store`. Đây là dấu hiệu code cụ thể cho dependency inversion trong project này.

### Checkpoint 3

```powershell
go test ./...
go build ./...
```

Ở thời điểm này chưa có implementation của `EventStore`, nhưng package processor vẫn build được. Interface được thỏa ở chặng kế tiếp.

---

## 7. Chặng 4 — MemoryEventStore làm fake (75–95 phút)

### 7.1. Mock, stub, spy và fake khác nhau thế nào?

Trong phạm vi buổi học:

- Stub trả về dữ liệu/lỗi đã định trước.
- Spy ghi lại lời gọi để test kiểm tra.
- Fake có implementation hoạt động đơn giản hơn production, ví dụ lưu vào RAM.
- Mock thường đặt expectation chi tiết về lời gọi, có thể do framework sinh ra.

`MemoryEventStore` là fake. `fakeService` ở test orchestration vừa là stub cho error, vừa là spy ghi thứ tự gọi. Không cần mocking framework cho interface một method.

### 7.2. Thiết kế MemoryEventStore

Tạo `internal/store/memory.go` với toàn bộ nội dung:

```go
package store

import (
	"context"
	"sync"

	"honeypot-day4/internal/model"
)

type MemoryEventStore struct {
	mu     sync.RWMutex
	events []model.Event
}

func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{}
}

func (s *MemoryEventStore) InsertEvent(ctx context.Context, event *model.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, *event)
	return nil
}

func (s *MemoryEventStore) Events() []model.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]model.Event(nil), s.events...)
}
```

Yêu cầu hành vi:

1. Nếu `ctx.Err()` khác `nil`, không insert và trả đúng lỗi context.
2. Khi insert, lưu `*event` dưới dạng value để caller sửa pointer sau đó không đổi dữ liệu đã lưu.
3. `Events()` trả một slice copy để caller không thay đổi backing array nội bộ.
4. Dùng mutex vì store có thể được gọi từ nhiều service trong phiên bản tương lai.

Phần copy slice nằm ở dòng:

```go
return append([]model.Event(nil), s.events...)
```

Không expose trực tiếp `s.events`.

Đây là shallow copy, nhưng đủ an toàn vì `Event` hiện chỉ chứa `string`. Nếu sau này `Event` có slice, map hoặc pointer tới mutable data, fake cần copy sâu các field đó một cách có chủ đích.

### 7.3. Test fake trước khi dùng nó

Tạo `internal/store/memory_test.go`:

```go
package store_test

import (
	"context"
	"testing"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/store"
)

func TestMemoryEventStoreKeepsIndependentCopies(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	event := &model.Event{Service: "http", Message: "request"}

	if err := eventStore.InsertEvent(context.Background(), event); err != nil {
		t.Fatalf("InsertEvent() error = %v", err)
	}
	event.Message = "changed after insert"

	firstRead := eventStore.Events()
	firstRead[0].Message = "changed returned copy"

	secondRead := eventStore.Events()
	if got, want := secondRead[0].Message, "request"; got != want {
		t.Fatalf("stored Message = %q, want %q", got, want)
	}
}

func TestMemoryEventStoreHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	eventStore := store.NewMemoryEventStore()
	err := eventStore.InsertEvent(ctx, &model.Event{})
	if err != context.Canceled {
		t.Fatalf("InsertEvent() error = %v, want %v", err, context.Canceled)
	}
	if got := len(eventStore.Events()); got != 0 {
		t.Fatalf("len(Events()) = %d, want 0", got)
	}
}
```

Chạy:

```powershell
go test -count=1 ./internal/store -v
```

### 7.4. Dùng fake để test processor

Tạo `internal/processor/processor_test.go`:

```go
package processor_test

import (
	"context"
	"testing"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/processor"
	"honeypot-day4/internal/store"
)

func TestProcessStoresEvent(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)
	event := &model.Event{Service: "telnet", Message: "login attempt"}

	if err := eventProcessor.Process(context.Background(), event); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	events := eventStore.Events()
	if len(events) != 1 {
		t.Fatalf("len(Events()) = %d, want 1", len(events))
	}
	if events[0] != *event {
		t.Fatalf("stored event = %#v, want %#v", events[0], *event)
	}
}

func TestProcessRejectsNilEvent(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)

	err := eventProcessor.Process(context.Background(), nil)
	if err == nil {
		t.Fatal("Process() error = nil, want an error")
	}
	if err.Error() != "event is required" {
		t.Fatalf("Process() error = %q, want %q", err, "event is required")
	}
	if got := len(eventStore.Events()); got != 0 {
		t.Fatalf("len(Events()) = %d, want 0", got)
	}
}
```

Điểm quan trọng: processor test không import PostgreSQL, không mở connection và không cần cleanup database.

### Checkpoint 4

```powershell
go test -count=1 ./internal/store ./internal/processor -v
```

- Memory store thỏa `EventStore` một cách implicit.
- Processor test nhanh và deterministic.
- Processor không biết fake là memory store.

---

## 8. Chặng 5 — Ghép service với processor (95–110 phút)

### 8.1. Nâng cấp hai service

Hai service hiện có `Start` chỉ trả `nil`. Bây giờ refactor chúng để giữ processor được inject từ ngoài.

Thay toàn bộ `internal/service/telnet.go` bằng:

```go
package service

import (
	"context"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
)

type TelnetService struct {
	processor *processor.Processor
}

func NewTelnet(p *processor.Processor) *TelnetService {
	return &TelnetService{processor: p}
}

func (s *TelnetService) Name() string {
	return "telnet"
}

func (s *TelnetService) Start(ctx context.Context) error {
	return s.processor.Process(ctx, &model.Event{
		Service: s.Name(),
		Message: "listener started",
	})
}

var _ orchestration.Service = (*TelnetService)(nil)
```

Thay toàn bộ `internal/service/http.go` bằng:

```go
package service

import (
	"context"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
)

type HTTPService struct {
	processor *processor.Processor
}

func NewHTTP(p *processor.Processor) *HTTPService {
	return &HTTPService{processor: p}
}

func (s *HTTPService) Name() string {
	return "http"
}

func (s *HTTPService) Start(ctx context.Context) error {
	return s.processor.Process(ctx, &model.Event{
		Service: s.Name(),
		Message: "listener started",
	})
}

var _ orchestration.Service = (*HTTPService)(nil)
```

Hai method `Start` vừa viết có hành vi:

- Telnet tạo event `{Service: "telnet", Message: "listener started"}`.
- HTTP tạo event `{Service: "http", Message: "listener started"}`.
- Cả hai gọi `Process(ctx, event)` và trả lại error.

Không copy logic lưu trữ vào service. Service tạo event; processor xử lý; store lưu trữ. Mỗi type giữ một trách nhiệm chính.

Sau refactor, compile-time assertions từ Chặng 1 vẫn phải xanh. Nếu method signature bị lệch, compiler báo ngay tại concrete service.

### 8.2. Ghép composition root trong main

`main` là nơi được phép biết concrete type để lắp hệ thống:

```text
MemoryEventStore → Processor → TelnetService/HTTPService → []Service
```

Thay `cmd/servicelab/main.go` bằng:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
	"honeypot-day4/internal/service"
	"honeypot-day4/internal/store"
)

func main() {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)

	services := []orchestration.Service{
		service.NewTelnet(eventProcessor),
		service.NewHTTP(eventProcessor),
	}

	if err := orchestration.StartServices(context.Background(), services); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, runningService := range services {
		fmt.Printf("started: %s\n", runningService.Name())
	}

	for _, event := range eventStore.Events() {
		fmt.Printf("stored event: service=%s message=%q\n", event.Service, event.Message)
	}
}
```

Tạo `internal/service/services_test.go` để chứng minh hai concrete type thật cùng chạy qua `[]Service`:

```go
package service_test

import (
	"context"
	"testing"

	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
	"honeypot-day4/internal/service"
	"honeypot-day4/internal/store"
)

func TestServicesStartThroughInterface(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)
	services := []orchestration.Service{
		service.NewTelnet(eventProcessor),
		service.NewHTTP(eventProcessor),
	}

	if err := orchestration.StartServices(context.Background(), services); err != nil {
		t.Fatalf("StartServices() error = %v", err)
	}

	events := eventStore.Events()
	if len(events) != 2 {
		t.Fatalf("len(Events()) = %d, want 2", len(events))
	}
	if events[0].Service != "telnet" || events[1].Service != "http" {
		t.Fatalf(
			"event services = [%q %q], want [telnet http]",
			events[0].Service,
			events[1].Service,
		)
	}
}
```

Output mục tiêu:

```text
started: telnet
started: http
stored event: service=telnet message="listener started"
stored event: service=http message="listener started"
```

Chạy:

```powershell
go test -count=1 ./internal/service -v
go run ./cmd/servicelab
```

### 8.3. Chứng minh caller không biết concrete type

Quan sát `StartServices`:

- Input là `[]Service`.
- Biến vòng lặp là `Service`.
- Function chỉ gọi `Name` và `Start`.
- Package orchestration không import concrete package.

`main` biết concrete type lúc khởi tạo vì phải lắp dependency graph. Sau khi đưa chúng vào `[]Service`, orchestration không còn cần biết concrete type. Dependency inversion không có nghĩa toàn chương trình tuyệt đối không được nhắc đến concrete type; concrete type phải được chọn ở composition root.

### Checkpoint 5

- CLI in đúng bốn dòng theo thứ tự.
- Store chứa hai event.
- Telnet và HTTP đi qua cùng một runner.
- Không service nào tự tạo memory store bên trong constructor hoặc `Start`.

---

## 9. Chặng 6 — Chứng minh Open/Closed Principle (110–116 phút)

Không cần xây SSH server thật. Tạo tạm `internal/service/ssh.go`:

```go
package service

import (
	"context"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
)

type SSHService struct {
	processor *processor.Processor
}

func NewSSH(p *processor.Processor) *SSHService {
	return &SSHService{processor: p}
}

func (s *SSHService) Name() string {
	return "ssh"
}

func (s *SSHService) Start(ctx context.Context) error {
	return s.processor.Process(ctx, &model.Event{
		Service: s.Name(),
		Message: "listener started",
	})
}

var _ orchestration.Service = (*SSHService)(nil)
```

Thêm đúng một dòng vào slice trong `main`:

```go
services := []orchestration.Service{
	service.NewTelnet(eventProcessor),
	service.NewHTTP(eventProcessor),
	service.NewSSH(eventProcessor),
}
```

Làm thí nghiệm:

1. Tạo `SSHService` có cùng hai method.
2. Cho `Start` ghi một event `ssh` qua processor.
3. Thêm instance vào slice trong `main`.
4. Không sửa một ký tự nào trong `StartServices`.
5. Chạy CLI và thấy event thứ ba.

Sau khi quan sát event thứ ba, hãy xóa `ssh.go` và dòng `service.NewSSH(...)` để quay lại đáp án chính chỉ có Telnet và HTTP. Điều cần chứng minh là orchestration đóng với việc sửa hành vi cũ nhưng mở với implementation mới.

Đây là OCP theo cách thực tế trong Go: thêm type thỏa một interface nhỏ, không tạo hierarchy hay `BaseService`.

---

## 10. Quality gate cuối buổi (116–120 phút)

Chạy đúng thứ tự:

```powershell
go fmt ./...
go test -count=1 ./...
go vet ./...
go build ./...
go run ./cmd/servicelab
```

Không coi project hoàn thành nếu chỉ CLI chạy nhưng test fail.

### Cấu trúc dự kiến

```text
go_day4/
├── cmd/
│   └── servicelab/
│       └── main.go
├── internal/
│   ├── model/
│   │   └── event.go
│   ├── orchestration/
│   │   ├── orchestration.go
│   │   └── orchestration_test.go
│   ├── processor/
│   │   ├── processor.go
│   │   └── processor_test.go
│   ├── service/
│   │   ├── telnet.go
│   │   ├── http.go
│   │   └── services_test.go
│   └── store/
│       ├── memory.go
│       └── memory_test.go
├── day4_tutorial.md
└── go.mod
```

---

## 11. Bản đồ SOLID trong chính project

| Nguyên tắc | Bằng chứng trong code |
|---|---|
| **S — Single Responsibility** | Service tạo event, processor điều phối use case, store lưu dữ liệu, runner khởi động service |
| **O — Open/Closed** | Thêm SSH service mà không sửa `StartServices` |
| **L — Liskov Substitution** | Telnet, HTTP và fake thay thế nhau tại vị trí `Service` mà vẫn giữ contract |
| **I — Interface Segregation** | `Service` có 2 method; `EventStore` có 1 method |
| **D — Dependency Inversion** | `Processor` giữ `EventStore`, không giữ concrete PostgreSQL/memory store |

Lưu ý về LSP: “compile được” mới chỉ chứng minh method set khớp. Implementation còn phải giữ semantic contract. Ví dụ một service trả `nil` từ `Start` dù việc khởi động thật sự thất bại vẫn compile, nhưng vi phạm contract thành công mà caller dựa vào.

---

## 12. Anti-pattern checklist

Trước khi kết thúc, tìm trong code và loại bỏ các dấu hiệu sau:

### Interface quá lớn

Không biến `Service` thành:

```text
Start + Stop + Restart + Health + Metrics + Config + Store + Logger + ...
```

Consumer hiện chỉ cần `Name` và `Start`.

### Interface đặt theo implementation

Không tạo `TelnetServiceInterface`. Interface nên nói khả năng consumer cần, không lặp lại tên concrete type.

### Interface cho mọi struct

`Event` chỉ là data; nó không cần `EventInterface`. `Processor` cũng không tự động cần interface nếu chưa có consumer cần thay thế nó.

### BaseService khổng lồ

Không nhúng một `BaseService` chứa mọi dependency và method chỉ để mô phỏng inheritance. Chia sẻ dependency bằng composition có chủ đích.

### `any` thay cho contract

Không dùng `[]any`, reflection hoặc type switch để gọi service. `[]Service` vừa rõ hơn vừa được compiler kiểm tra.

### Concrete dependency ẩn trong constructor

Không để `NewProcessor()` tự tạo PostgreSQL/memory store. Caller phải inject `EventStore`; nếu không, test không thể thay dependency sạch sẽ.

### Fake làm rò state

Không trả trực tiếp backing slice và không giữ pointer event của caller. Nếu fake cư xử quá khác production về ownership, test có thể cho tín hiệu sai.

---

## 13. Tự chấm cuối ngày

Không nhìn code, tự trả lời:

1. Tại sao Go gọi là implicit interface implementation?
2. Method set của `T` và `*T` ảnh hưởng phép gán vào interface thế nào?
3. Tại sao interface thường nên nằm ở phía consumer?
4. Đa hình xảy ra ở dòng nào trong `StartServices`?
5. Tại sao `MemoryEventStore` là fake, không chỉ là một slice?
6. Dependency nào đã bị đảo trong `Processor`?
7. `main` biết concrete type có vi phạm DIP không? Vì sao?
8. Tại sao `StartServices` tuần tự không phù hợp nguyên trạng cho hai server mạng block lâu dài?

Bạn thực sự hoàn thành Day 4 khi vừa trả lời được tám câu này, vừa có toàn bộ quality gate màu xanh.

---

## 14. Khi bị kẹt, debug theo thứ tự này

1. **Lỗi import cycle:** kiểm tra orchestration có import concrete `service` hay processor có import concrete `store` không.
2. **Does not implement:** so sánh chính xác tên method, receiver, tham số và kiểu trả về.
3. **Test order sai:** kiểm tra runner có duyệt đúng slice và fake có dùng chung slice ghi nhận không.
4. **`errors.Is` false:** dùng `%w` khi wrap cause.
5. **Event bị đổi ngoài ý muốn:** kiểm tra store có lưu pointer/backing slice của caller không.
6. **HTTP không chạy sau Telnet:** kiểm tra Telnet có trả lỗi và runner có contract dừng ở lỗi đầu tiên không.

Không sửa nhiều điểm cùng lúc. Chạy test package nhỏ nhất sau mỗi thay đổi, rồi mới chạy `go test ./...`.

---

## 15. Đáp án hoàn chỉnh để đối chiếu

Chỉ dùng phần này sau khi bạn đã tự đi hết các checkpoint. Đáp án cuối không giữ `SSHService` vì đó chỉ là thí nghiệm OCP; yêu cầu chính gồm Telnet và HTTP.

### 15.1. `go.mod`

```go
module honeypot-day4

go 1.26.5
```

Nếu máy bạn tạo một phiên bản Go khác ở dòng `go`, giữ phiên bản do `go mod init` sinh ra; không cần sửa chỉ để giống tutorial.

### 15.2. `internal/model/event.go`

```go
package model

type Event struct {
	Service string
	Message string
}
```

### 15.3. `internal/orchestration/orchestration.go`

```go
package orchestration

import (
	"context"
	"fmt"
)

type Service interface {
	Name() string
	Start(ctx context.Context) error
}

func StartServices(ctx context.Context, services []Service) error {
	for _, service := range services {
		if err := service.Start(ctx); err != nil {
			return fmt.Errorf("start %s: %w", service.Name(), err)
		}
	}
	return nil
}
```

### 15.4. `internal/orchestration/orchestration_test.go`

```go
package orchestration_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"honeypot-day4/internal/orchestration"
)

type fakeService struct {
	name    string
	started *[]string
	err     error
}

func (s fakeService) Name() string {
	return s.name
}

func (s fakeService) Start(context.Context) error {
	*s.started = append(*s.started, s.name)
	return s.err
}

func TestStartServicesUsesOnlyServiceContract(t *testing.T) {
	var started []string
	services := []orchestration.Service{
		fakeService{name: "telnet", started: &started},
		fakeService{name: "http", started: &started},
	}

	if err := orchestration.StartServices(context.Background(), services); err != nil {
		t.Fatalf("StartServices() error = %v", err)
	}

	want := []string{"telnet", "http"}
	if !reflect.DeepEqual(started, want) {
		t.Fatalf("started = %v, want %v", started, want)
	}
}

func TestStartServicesStopsAndWrapsError(t *testing.T) {
	var started []string
	startErr := errors.New("port unavailable")
	services := []orchestration.Service{
		fakeService{name: "telnet", started: &started, err: startErr},
		fakeService{name: "http", started: &started},
	}

	err := orchestration.StartServices(context.Background(), services)
	if !errors.Is(err, startErr) {
		t.Fatalf("StartServices() error = %v, want wrapped %v", err, startErr)
	}
	if got, want := err.Error(), "start telnet: port unavailable"; got != want {
		t.Fatalf("StartServices() error = %q, want %q", got, want)
	}

	want := []string{"telnet"}
	if !reflect.DeepEqual(started, want) {
		t.Fatalf("started = %v, want %v", started, want)
	}
}
```

### 15.5. `internal/processor/processor.go`

```go
package processor

import (
	"context"
	"errors"

	"honeypot-day4/internal/model"
)

type EventStore interface {
	InsertEvent(ctx context.Context, event *model.Event) error
}

type Processor struct {
	store EventStore
}

func New(store EventStore) *Processor {
	return &Processor{store: store}
}

func (p *Processor) Process(ctx context.Context, event *model.Event) error {
	if event == nil {
		return errors.New("event is required")
	}

	return p.store.InsertEvent(ctx, event)
}
```

### 15.6. `internal/processor/processor_test.go`

```go
package processor_test

import (
	"context"
	"testing"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/processor"
	"honeypot-day4/internal/store"
)

func TestProcessStoresEvent(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)
	event := &model.Event{Service: "telnet", Message: "login attempt"}

	if err := eventProcessor.Process(context.Background(), event); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	events := eventStore.Events()
	if len(events) != 1 {
		t.Fatalf("len(Events()) = %d, want 1", len(events))
	}
	if events[0] != *event {
		t.Fatalf("stored event = %#v, want %#v", events[0], *event)
	}
}

func TestProcessRejectsNilEvent(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)

	err := eventProcessor.Process(context.Background(), nil)
	if err == nil {
		t.Fatal("Process() error = nil, want an error")
	}
	if err.Error() != "event is required" {
		t.Fatalf("Process() error = %q, want %q", err, "event is required")
	}
	if got := len(eventStore.Events()); got != 0 {
		t.Fatalf("len(Events()) = %d, want 0", got)
	}
}
```

### 15.7. `internal/store/memory.go`

```go
package store

import (
	"context"
	"sync"

	"honeypot-day4/internal/model"
)

type MemoryEventStore struct {
	mu     sync.RWMutex
	events []model.Event
}

func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{}
}

func (s *MemoryEventStore) InsertEvent(ctx context.Context, event *model.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, *event)
	return nil
}

func (s *MemoryEventStore) Events() []model.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]model.Event(nil), s.events...)
}
```

### 15.8. `internal/store/memory_test.go`

```go
package store_test

import (
	"context"
	"testing"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/store"
)

func TestMemoryEventStoreKeepsIndependentCopies(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	event := &model.Event{Service: "http", Message: "request"}

	if err := eventStore.InsertEvent(context.Background(), event); err != nil {
		t.Fatalf("InsertEvent() error = %v", err)
	}
	event.Message = "changed after insert"

	firstRead := eventStore.Events()
	firstRead[0].Message = "changed returned copy"

	secondRead := eventStore.Events()
	if got, want := secondRead[0].Message, "request"; got != want {
		t.Fatalf("stored Message = %q, want %q", got, want)
	}
}

func TestMemoryEventStoreHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	eventStore := store.NewMemoryEventStore()
	err := eventStore.InsertEvent(ctx, &model.Event{})
	if err != context.Canceled {
		t.Fatalf("InsertEvent() error = %v, want %v", err, context.Canceled)
	}
	if got := len(eventStore.Events()); got != 0 {
		t.Fatalf("len(Events()) = %d, want 0", got)
	}
}
```

### 15.9. `internal/service/telnet.go`

```go
package service

import (
	"context"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
)

type TelnetService struct {
	processor *processor.Processor
}

func NewTelnet(p *processor.Processor) *TelnetService {
	return &TelnetService{processor: p}
}

func (s *TelnetService) Name() string {
	return "telnet"
}

func (s *TelnetService) Start(ctx context.Context) error {
	return s.processor.Process(ctx, &model.Event{
		Service: s.Name(),
		Message: "listener started",
	})
}

var _ orchestration.Service = (*TelnetService)(nil)
```

### 15.10. `internal/service/http.go`

```go
package service

import (
	"context"

	"honeypot-day4/internal/model"
	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
)

type HTTPService struct {
	processor *processor.Processor
}

func NewHTTP(p *processor.Processor) *HTTPService {
	return &HTTPService{processor: p}
}

func (s *HTTPService) Name() string {
	return "http"
}

func (s *HTTPService) Start(ctx context.Context) error {
	return s.processor.Process(ctx, &model.Event{
		Service: s.Name(),
		Message: "listener started",
	})
}

var _ orchestration.Service = (*HTTPService)(nil)
```

### 15.11. `internal/service/services_test.go`

```go
package service_test

import (
	"context"
	"testing"

	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
	"honeypot-day4/internal/service"
	"honeypot-day4/internal/store"
)

func TestServicesStartThroughInterface(t *testing.T) {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)
	services := []orchestration.Service{
		service.NewTelnet(eventProcessor),
		service.NewHTTP(eventProcessor),
	}

	if err := orchestration.StartServices(context.Background(), services); err != nil {
		t.Fatalf("StartServices() error = %v", err)
	}

	events := eventStore.Events()
	if len(events) != 2 {
		t.Fatalf("len(Events()) = %d, want 2", len(events))
	}
	if events[0].Service != "telnet" || events[1].Service != "http" {
		t.Fatalf(
			"event services = [%q %q], want [telnet http]",
			events[0].Service,
			events[1].Service,
		)
	}
}
```

### 15.12. `cmd/servicelab/main.go`

```go
package main

import (
	"context"
	"fmt"
	"os"

	"honeypot-day4/internal/orchestration"
	"honeypot-day4/internal/processor"
	"honeypot-day4/internal/service"
	"honeypot-day4/internal/store"
)

func main() {
	eventStore := store.NewMemoryEventStore()
	eventProcessor := processor.New(eventStore)

	services := []orchestration.Service{
		service.NewTelnet(eventProcessor),
		service.NewHTTP(eventProcessor),
	}

	if err := orchestration.StartServices(context.Background(), services); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, runningService := range services {
		fmt.Printf("started: %s\n", runningService.Name())
	}

	for _, event := range eventStore.Events() {
		fmt.Printf("stored event: service=%s message=%q\n", event.Service, event.Message)
	}
}
```

### 15.13. Chạy đáp án từ đầu đến cuối

```powershell
go fmt ./...
go test -count=1 ./...
go vet ./...
go build ./...
go run ./cmd/servicelab
```

Kết quả test có dạng:

```text
?    honeypot-day4/cmd/servicelab             [no test files]
?    honeypot-day4/internal/model              [no test files]
ok   honeypot-day4/internal/orchestration
ok   honeypot-day4/internal/processor
ok   honeypot-day4/internal/service
ok   honeypot-day4/internal/store
```

Kết quả chương trình:

```text
started: telnet
started: http
stored event: service=telnet message="listener started"
stored event: service=http message="listener started"
```

Nếu output đúng và toàn bộ quality gate thành công, bạn đã hoàn thành Day 4: implicit interface implementation, polymorphism, interface phía consumer, dependency inversion và test double bằng fake implementation.

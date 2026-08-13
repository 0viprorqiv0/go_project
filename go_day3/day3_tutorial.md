# Day 3 — Vừa học OOP trong Go, vừa xây Session Lifecycle

Hôm nay ta không học hết lý thuyết rồi mới làm bài. Ta sẽ xây một project nhỏ theo từng chặng. Mỗi khi project gặp một vấn đề, ta học đúng khái niệm Go cần để giải quyết vấn đề đó.

Nhịp làm việc xuyên suốt:

```text
Vấn đề → Kiến thức vừa đủ → Viết code/test → Chạy → Giải thích kết quả
```

Trong một chặng, compiler hoặc test có thể tạm RED vì ta cố ý phơi bày phần còn thiếu. Đến checkpoint cuối chặng, project phải quay lại trạng thái build và test được trước khi đi tiếp.

> Các file trong `C:\Go_project\go_day3` là lời giải cuối cùng. Để thấy đúng quá trình RED → GREEN, hãy thực hành trong một thư mục trống như `C:\Go_project\go_day3_practice`. Nếu làm trực tiếp trên lời giải, các method và test của những chặng tương lai đã tồn tại nên kết quả quan sát sẽ không còn đúng.

---

## 1. Nhiệm vụ của project

Ta sẽ xây `sessionlab`, một chương trình mô phỏng lifecycle của một kết nối honeypot:

```text
Tạo session
    ↓
Ghi nhận hai command
    ↓
Đóng session
    ↓
Từ chối command đến sau khi đóng
    ↓
In snapshot cuối cùng
```

Đây chưa phải server mạng. Các command chỉ là dữ liệu giả lập; chương trình tuyệt đối không thực thi chúng bằng shell.

### Yêu cầu hành vi

Project hoàn thành khi thỏa các điều sau:

1. ID và source IP là bắt buộc.
2. Session mới bắt đầu ở trạng thái mở với `0` command.
3. Mỗi lần `AddCommand` thành công làm count tăng đúng `1`.
4. Session đã đóng không nhận thêm command.
5. Code ngoài package không sửa trực tiếp state thật.
6. Caller có thể lấy một snapshot để hiển thị mà không làm lộ mutable state.

Output cuối ngày:

```text
recorded command: "whoami"
recorded command: "wget http://example.com/a.sh"
rejected command after Close: session is closed

=== Session Snapshot ===
ID:            sess-001
Source IP:     203.0.113.10
Started at:    2026-08-13T09:00:00Z
Command count: 2
Closed:        true
```

Trong quá trình xây project, ta sẽ lần lượt áp dụng:

- Struct và nested struct.
- Method.
- Pointer receiver và value receiver.
- Constructor convention `New...`.
- Validation invariant.
- Exported và unexported field.
- Package boundary để đóng gói.
- Composition thay cho inheritance.

Thời gian đề xuất: **4 giờ** để đi hết project từ thư mục trống đến chương trình hoàn chỉnh.

---

## 2. Chặng 0 — Tạo vòng phản hồi đầu tiên

Mục tiêu đầu tiên rất nhỏ: tạo module và chạy được một chương trình. Ta cần vòng phản hồi nhanh trước khi viết domain model.

### 2.1. Khởi tạo project

Mở PowerShell tại `C:\Go_project`. Tạo một thư mục thực hành mới; nếu `go_day3_practice` đã có dữ liệu, hãy chọn tên khác thay vì ghi đè:

```powershell
New-Item -ItemType Directory -Path go_day3_practice
New-Item -ItemType Directory -Force -Path go_day3_practice\cmd\sessionlab
New-Item -ItemType Directory -Force -Path go_day3_practice\internal\session
Set-Location go_day3_practice
go mod init honeypot-day3
```

Tên thư mục thực hành không quyết định import path; dòng `module honeypot-day3` trong `go.mod` mới quyết định. Nếu bạn chủ động dùng một thư mục trống đã có `go.mod`, không chạy lại `go mod init`. Chỉ kiểm tra:

```powershell
go env GOMOD
go list -m
```

Kết quả của `go list -m` phải là:

```text
honeypot-day3
```

### 2.2. Tạo chương trình đầu tiên

Tạo `cmd/sessionlab/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("sessionlab: ready")
}
```

Chạy:

```powershell
go run ./cmd/sessionlab
```

Kết quả:

```text
sessionlab: ready
```

### Kết quả chặng 0

Chạy gate đầu tiên:

```powershell
go test ./...
go build ./...
```

- `go list -m` in `honeypot-day3`.
- `go run ./cmd/sessionlab` in `sessionlab: ready`.
- Executable nằm dưới `cmd/sessionlab`.

Nếu chưa đạt đủ ba điều này, dừng lại sửa trước khi đi tiếp.

---

## 3. Chặng 1 — Dùng struct để biểu diễn một session

Project cần giữ năm mẩu trạng thái:

- ID của session.
- IP nguồn.
- Thời điểm bắt đầu.
- Số command đã thấy.
- Session đã đóng hay chưa.

Struct là công cụ của Go để gom các dữ liệu có liên quan thành một kiểu.

### 3.1. Làm prototype nhanh

Tạo `internal/session/session.go`:

```go
package session

import "time"

type Session struct {
	ID           string
	SourceIP     string
	StartedAt    time.Time
	CommandCount int
	Closed       bool
}
```

Tạm thay nội dung `cmd/sessionlab/main.go` bằng:

```go
package main

import (
	"fmt"
	"time"

	"honeypot-day3/internal/session"
)

func main() {
	sess := session.Session{
		ID:        "sess-001",
		SourceIP:  "203.0.113.10",
		StartedAt: time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC),
	}

	fmt.Printf("initial: count=%d closed=%t\n", sess.CommandCount, sess.Closed)

	// Caller có thể tạo state vô nghĩa.
	sess.CommandCount = -100
	sess.Closed = true
	sess.Closed = false

	fmt.Printf("changed: count=%d closed=%t\n", sess.CommandCount, sess.Closed)
}
```

Chạy:

```powershell
go run ./cmd/sessionlab
```

Bạn sẽ thấy:

```text
initial: count=0 closed=false
changed: count=-100 closed=false
```

### 3.2. Ta vừa học gì?

`Session` là một kiểu struct. Mỗi giá trị bên trong là một field.

Ta không gán `CommandCount` và `Closed` trong struct literal, nhưng Go vẫn khởi tạo chúng bằng zero value:

| Kiểu | Zero value |
|---|---|
| `int` | `0` |
| `bool` | `false` |
| `string` | `""` |
| pointer | `nil` |

Tên `Session` và các field đều bắt đầu bằng chữ hoa, vì vậy chúng được **export**. Package `main` có thể đọc và ghi trực tiếp.

Prototype chạy được, nhưng thiết kế có vấn đề:

- Count âm vẫn được chấp nhận.
- Caller có thể tự mở lại session.
- Không có nơi tập trung để kiểm tra dữ liệu đầu vào.

Struct đã giữ được trạng thái, nhưng chưa bảo vệ được trạng thái.

### Đáp án của thí nghiệm

- `CommandCount` tự có giá trị `0` vì zero value của `int` là `0`.
- Package `main` sửa được mọi field vì tên field bắt đầu bằng chữ hoa nên được export.
- Nếu caller tự gán field, validation sẽ bị phân tán hoặc bị bỏ qua. Chặng tiếp theo đưa việc tạo state hợp lệ vào `NewSession` và giữ mutable state ở field unexported.

### Kết quả chặng 1

Ta đã có một prototype chạy được và quan sát rõ lý do cần encapsulation. Phiên bản đầu cho phép state sai; chặng tiếp theo sẽ thay thế thiết kế này.

Trước khi refactor, xác nhận toàn project vẫn xanh:

```powershell
go test ./...
go build ./...
```

---

## 4. Chặng 2 — Đóng gói state và bảo vệ lúc khởi tạo

Ta muốn giữ quyền thay đổi state bên trong package `session`. Trong Go, package boundary và cách viết hoa tên tạo nên cơ chế đóng gói.

### 4.1. Đổi field thành unexported

Thay `internal/session/session.go` bằng:

```go
package session

import "time"

type Session struct {
	id           string
	sourceIP     string
	startedAt    time.Time
	commandCount int
	closed       bool
}
```

Các tên bắt đầu bằng chữ thường là unexported. Thử build CLI cũ:

```powershell
go build ./cmd/sessionlab
```

Build phải thất bại với lỗi có dạng:

```text
unknown field ID in struct literal of type session.Session
```

Đây là lỗi có chủ đích. Compiler vừa chứng minh package `main` không còn được tự ý dựng và sửa state bên trong.

### 4.2. Project cần một đường tạo object hợp lệ

Khi field đã ẩn, caller cần một function public để tạo session. Trước khi viết function đó, ta viết test cho quy tắc đầu tiên:

> Không được tạo session nếu thiếu ID hoặc source IP.

Tạo mới `internal/session/session_test.go` với toàn bộ nội dung sau:

```go
package session_test

import (
	"testing"
	"time"

	"honeypot-day3/internal/session"
)

func TestNewSessionRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		sourceIP string
	}{
		{name: "missing ID", sourceIP: "203.0.113.10"},
		{name: "missing source IP", id: "sess-001"},
		{name: "missing both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := session.NewSession(tt.id, tt.sourceIP, time.Now())
			if err == nil {
				t.Fatal("NewSession() error = nil, want an error")
			}
			if err.Error() != "id and source IP are required" {
				t.Fatalf("NewSession() error = %q, want %q", err, "id and source IP are required")
			}
			if got != nil {
				t.Fatalf("NewSession() session = %#v, want nil", got)
			}
		})
	}
}
```

Đọc cấu trúc test trước khi chạy:

- `package session_test` là external test package. Test chỉ dùng được public API, giống caller thật.
- Slice `tests` chứa nhiều input; vòng lặp chạy cùng một quy tắc cho từng input. Đây là table-driven test đã gặp ở Day 2.
- `t.Run` tạo một subtest có tên riêng, nên khi fail ta biết case nào sai.
- Bài kiểm tra cả error text vì đặc tả Day 3 đã cho thông báo cụ thể. Trong API lớn, nếu caller cần phân nhánh ổn định, sentinel error và `errors.Is` thường phù hợp hơn.

Chạy đúng test này:

```powershell
go test ./internal/session -run '^TestNewSessionRejectsMissingRequiredFields$' -v
```

`-run` nhận regular expression và chỉ **chạy** test có tên khớp; Go vẫn compile toàn bộ package cùng mọi test file trong đó. Đây là một lý do ta dùng thư mục thực hành trống thay vì đặt tutorial từng bước lên trên lời giải hoàn chỉnh.

Test chưa thể compile vì `NewSession` chưa tồn tại. Đây là trạng thái **RED**: test mô tả API và hành vi project đang thiếu. Khi implementation làm test pass, ta có trạng thái **GREEN**.

### 4.3. Viết constructor nhỏ nhất để test pass

Sửa phần import trong `session.go`:

```go
import (
	"errors"
	"time"
)
```

Thêm function:

```go
func NewSession(id, sourceIP string, now time.Time) (*Session, error) {
	if id == "" || sourceIP == "" {
		return nil, errors.New("id and source IP are required")
	}

	return &Session{
		id:        id,
		sourceIP:  sourceIP,
		startedAt: now,
	}, nil
}
```

Chạy lại:

```powershell
go test ./internal/session -run '^TestNewSessionRejectsMissingRequiredFields$' -v
```

Lần này test phải **PASS**.

### 4.4. Kiến thức vừa đủ: constructor và invariant

Go không có constructor keyword. `NewSession` là convention do cộng đồng sử dụng, không phải function được runtime tự gọi.

Function trả `(*Session, error)` vì:

- Session là object có state thay đổi, nên caller giữ pointer đến cùng một object.
- Input có thể sai, nên lỗi được trả về như một giá trị.

Invariant là quy tắc phải được duy trì trong suốt trạng thái hợp lệ của object. Public construction path hiện bảo vệ hai quy tắc nền tảng: `id` và `sourceIP` không rỗng.

Constructor còn tạo ra các **postcondition ban đầu**:

- `startedAt` nhận đúng đối số `now`.
- Session mới có count `0`.
- Session mới có `closed == false`.

Count và `closed` sẽ thay đổi hợp lệ trong lifecycle, nên hai giá trị ban đầu đó không phải invariant vĩnh viễn. Ở các chặng sau, `AddCommand` và `Close` sẽ bảo vệ thêm quy tắc: count không âm và không tăng sau khi session đã đóng.

Truyền `now` từ caller thay vì gọi `time.Now()` trong constructor giúp test dùng thời gian cố định.

Contract của project chỉ kiểm tra chuỗi rỗng. Nó không từ chối chuỗi toàn khoảng trắng, IP sai hoặc `time.Time{}`; các trường hợp đó nằm ngoài phạm vi Day 3.

### 4.5. Sửa CLI để dùng constructor

CLI cũ vẫn dùng struct literal nên chưa build được. Tạm thay `main` bằng:

```go
func main() {
	startedAt := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)
	sess, err := session.NewSession("sess-001", "203.0.113.10", startedAt)
	if err != nil {
		fmt.Println("create session:", err)
		return
	}

	fmt.Printf("session created: %T\n", sess)
}
```

Giữ các import `fmt`, `time` và `honeypot-day3/internal/session`, rồi chạy:

```powershell
go run ./cmd/sessionlab
```

Kết quả:

```text
session created: *session.Session
```

### Package `internal` bảo vệ điều gì?

Với `<thư_mục_thực_hành>/internal/session`, package import phải nằm trong cây bắt đầu tại thư mục cha của `internal`. CLI trong cùng project import được package này; code nằm ngoài cây đó thì không.

`internal` giới hạn nơi được import package. Field unexported giới hạn identifier nào caller được truy cập. Hai cơ chế có vai trò khác nhau.

### Kết quả chặng 2

Chạy gate của cả chặng, không chỉ test vừa viết:

```powershell
go test ./...
go build ./...
```

- Field thật của `Session` đều unexported.
- Constructor từ chối dữ liệu bắt buộc bị thiếu.
- Test constructor pass.
- CLI tạo session bằng `NewSession`, không dùng struct literal.

---

## 5. Chặng 3 — Cho caller đọc state mà không mở quyền sửa

State đã được đóng gói, nhưng CLI sắp cần hiển thị ID, IP, thời gian, count và trạng thái đóng. Nếu export lại field, ta quay về vấn đề cũ.

Giải pháp của project là trả một **snapshot**: bản sao của state tại một thời điểm.

### 5.1. Viết test cho initial snapshot

Thêm test sau vào `session_test.go`:

```go
func TestNewSessionInitializesValidState(t *testing.T) {
	now := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)

	sess, err := session.NewSession("sess-001", "203.0.113.10", now)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	got := sess.Snapshot()
	if got.Identity.ID != "sess-001" {
		t.Fatalf("ID = %q, want %q", got.Identity.ID, "sess-001")
	}
	if got.Identity.SourceIP != "203.0.113.10" {
		t.Fatalf("SourceIP = %q, want %q", got.Identity.SourceIP, "203.0.113.10")
	}
	if !got.StartedAt.Equal(now) {
		t.Fatalf("StartedAt = %v, want %v", got.StartedAt, now)
	}
	if got.CommandCount != 0 {
		t.Fatalf("CommandCount = %d, want 0", got.CommandCount)
	}
	if got.Closed {
		t.Fatal("new session must be open")
	}
}
```

Chạy:

```powershell
go test ./internal/session -run '^TestNewSessionInitializesValidState$' -v
```

Test RED vì `Snapshot` chưa tồn tại.

### 5.2. Tạo nested struct và snapshot

Thêm hai type vào `session.go`:

```go
type Identity struct {
	ID       string
	SourceIP string
}

type Snapshot struct {
	Identity     Identity
	StartedAt    time.Time
	CommandCount int
	Closed       bool
}
```

`Snapshot` chứa một `Identity`. Đây là nested struct và cũng là composition:

```text
Snapshot có một Identity
```

Không có quan hệ kế thừa `Snapshot is-a Identity`.

Thêm method:

```go
func (s Session) Snapshot() Snapshot {
	return Snapshot{
		Identity: Identity{
			ID:       s.id,
			SourceIP: s.sourceIP,
		},
		StartedAt:    s.startedAt,
		CommandCount: s.commandCount,
		Closed:       s.closed,
	}
}
```

Chạy lại test:

```powershell
go test ./internal/session -run '^TestNewSessionInitializesValidState$' -v
```

Test phải PASS.

### 5.3. Kiến thức vừa đủ: method và value receiver

Method là function có receiver:

```go
func (s Session) Snapshot() Snapshot
```

- `s` là tên receiver.
- `Session` là kiểu nhận method.
- Caller dùng `sess.Snapshot()`.

Receiver ở đây là **value receiver**. Method nhận một bản sao của struct `Session`, chỉ đọc nó rồi tạo một giá trị mới. Struct hiện không giữ mutable reference cần deep-copy hay mutex; `time.Time` được thiết kế để copy an toàn. Vì vậy lựa chọn này dễ quan sát và phù hợp với mục tiêu bài học.

Snapshot không “bất biến” theo nghĩa tuyệt đối. Caller có thể sửa snapshot, nhưng thay đổi đó không đi ngược vào state thật của session. Ta sẽ chứng minh bằng test ở chặng sau.

> Khi type lớn hơn, chứa `sync.Mutex`, hoặc team muốn receiver nhất quán cho một mutable aggregate, `Snapshot` có thể dùng pointer receiver. Không có quy tắc “method chỉ đọc thì luôn dùng value receiver”.

### 5.4. Thử phá encapsulation

Tạm thêm vào CLI:

```go
sess.closed = false
```

Chạy:

```powershell
go build ./cmd/sessionlab
```

Compiler phải báo lỗi có dạng `sess.closed undefined (cannot refer to unexported field closed)`. Xóa dòng thử nghiệm để project build lại.

Unexported không phải encryption hay authorization. Code trong cùng package `session` vẫn truy cập được field; code package khác thì không.

### Kết quả chặng 3

Sau khi đã xóa dòng compile-fail, chạy full gate:

```powershell
go test ./...
go build ./...
```

- `Snapshot` chứa nested `Identity`.
- Initial-state test pass.
- Caller đọc state qua snapshot.
- Caller không truy cập trực tiếp `closed` hoặc `commandCount`.

---

## 6. Chặng 4 — Ghi nhận command bằng hành vi

Project cần tăng count mỗi khi thấy một command. Ta không tạo `SetCommandCount`, vì caller có thể truyền `-10` hoặc `999999` mà không diễn đạt hành vi nghiệp vụ nào.

API nên nói điều caller muốn làm:

```go
sess.AddCommand()
```

### 6.1. Viết test trước

Thêm test:

```go
func TestAddCommandIncrementsCount(t *testing.T) {
	sess := newTestSession(t)

	for want := 1; want <= 2; want++ {
		if err := sess.AddCommand(); err != nil {
			t.Fatalf("AddCommand() error = %v", err)
		}
		if got := sess.Snapshot().CommandCount; got != want {
			t.Fatalf("CommandCount = %d, want %d", got, want)
		}
	}
}
```

Thêm helper ở cuối file:

```go
func newTestSession(t *testing.T) *session.Session {
	t.Helper()

	sess, err := session.NewSession(
		"sess-test",
		"192.0.2.10",
		time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}
	return sess
}
```

`t.Helper()` báo cho Go rằng đây là test helper. Nếu constructor lỗi, vị trí failure sẽ trỏ về test gọi helper thay vì chỉ trỏ vào bên trong `newTestSession`.

Chạy:

```powershell
go test ./internal/session -run '^TestAddCommandIncrementsCount$' -v
```

Test chưa compile vì method chưa tồn tại.

### 6.2. Thử value receiver và quan sát lỗi logic

Đầu tiên, cố tình viết sai:

```go
func (s Session) AddCommand() error {
	s.commandCount++
	return nil
}
```

Chạy lại test. Code biên dịch, nhưng test phải fail với kết quả có dạng:

```text
CommandCount = 0, want 1
```

Vì sao? Value receiver nhận một bản sao. Method tăng count trên bản sao rồi bỏ bản sao đó khi return.

### 6.3. Đổi sang pointer receiver

Sửa receiver thành `*Session`:

```go
func (s *Session) AddCommand() error {
	s.commandCount++
	return nil
}
```

Chạy lại:

```powershell
go test ./internal/session -run '^TestAddCommandIncrementsCount$' -v
```

Test phải PASS.

Focused test cho phản hồi nhanh; sau đó chạy toàn bộ suite để bắt regression:

```powershell
go test ./...
```

### 6.4. Khi nào dùng pointer receiver?

Dùng pointer receiver khi method:

- Cần thay đổi object gốc.
- Không nên copy một struct lớn.
- Làm việc với type chứa mutex hoặc state không được copy.
- Cần duy trì identity/state dùng chung qua nhiều caller.

`AddCommand` thay đổi `commandCount`, nên pointer receiver là bắt buộc cho hành vi mong muốn.

### Đáp án của thí nghiệm receiver

`Snapshot` tạo dữ liệu mới từ một bản sao; `AddCommand` phải thay đổi session gốc. Vì vậy hai method có receiver khác nhau trong project nhỏ này.

### Kết quả chặng 4

- Test đã chỉ ra value receiver không thay đổi count thật.
- `AddCommand` dùng pointer receiver.
- Gọi hai lần làm count lần lượt là `1` rồi `2`.
- Project không có `SetCommandCount`.

---

## 7. Chặng 5 — Xây lifecycle đóng session

Hiện tại session nhận command mãi mãi. Project cần transition từ `open` sang `closed` và không được tăng count sau transition đó.

### 7.1. Viết test cho transition trước

Thêm test:

```go
func TestClosedSessionRejectsNewCommand(t *testing.T) {
	sess := newTestSession(t)
	if err := sess.AddCommand(); err != nil {
		t.Fatalf("first AddCommand() error = %v", err)
	}

	sess.Close()
	err := sess.AddCommand()
	if err == nil {
		t.Fatal("AddCommand() error = nil after Close, want an error")
	}
	if err.Error() != "session is closed" {
		t.Fatalf("AddCommand() error = %q, want %q", err, "session is closed")
	}

	got := sess.Snapshot()
	if !got.Closed {
		t.Fatal("Closed = false after Close, want true")
	}
	if got.CommandCount != 1 {
		t.Fatalf("CommandCount = %d after rejected command, want 1", got.CommandCount)
	}
}
```

Chạy:

```powershell
go test ./internal/session -run '^TestClosedSessionRejectsNewCommand$' -v
```

Test RED ngay ở bước compile vì `Close` chưa tồn tại. Test đang buộc ta thiết kế hai hành vi cùng phối hợp: đóng session và từ chối command tiếp theo.

### 7.2. Thêm hành vi `Close`

Thêm vào `session.go`:

```go
func (s *Session) Close() {
	s.closed = true
}
```

Gọi lại focused test:

```powershell
go test ./internal/session -run '^TestClosedSessionRejectsNewCommand$' -v
```

Lần này code compile, nhưng test vẫn RED vì `AddCommand` tiếp tục tăng count và trả `nil`. Một test có thể dẫn ta qua nhiều bước nhỏ trước khi GREEN.

Project chọn `Close` là idempotent: gọi nhiều lần vẫn cho cùng trạng thái đóng, để cleanup code không cần kiểm tra trước. Ta sẽ khóa lựa chọn contract này bằng test riêng.

### 7.3. Validate trước, mutate sau

Thay `AddCommand` bằng:

```go
func (s *Session) AddCommand() error {
	if s.closed {
		return errors.New("session is closed")
	}

	s.commandCount++
	return nil
}
```

Chạy lại focused test:

```powershell
go test ./internal/session -run '^TestClosedSessionRejectsNewCommand$' -v
```

Lần này nó phải PASS.

Thứ tự rất quan trọng. Implementation sau là sai:

```go
s.commandCount++
if s.closed {
	return errors.New("session is closed")
}
```

Function vẫn trả lỗi, nhưng state đã bị thay đổi. Vì vậy test nhánh lỗi phải kiểm tra cả error **và** count không đổi.

### 7.4. Khóa contract idempotent bằng test

Thêm:

```go
func TestCloseIsIdempotent(t *testing.T) {
	sess := newTestSession(t)

	sess.Close()
	sess.Close()

	if !sess.Snapshot().Closed {
		t.Fatal("Closed = false after two calls to Close, want true")
	}
}
```

Chạy hai test lifecycle:

```powershell
go test ./internal/session -run '^(TestClosedSessionRejectsNewCommand|TestCloseIsIdempotent)$' -v
```

Cả hai phải PASS.

Cuối chặng, chạy lại toàn bộ suite:

```powershell
go test ./...
```

### Kết quả chặng 5

- `Close` dùng pointer receiver.
- Session chuyển từ open sang closed.
- Command sau `Close` trả lỗi.
- Command bị từ chối không làm count tăng.
- Gọi `Close` nhiều lần vẫn an toàn theo contract đã chọn.

---

## 8. Chặng 6 — Chứng minh snapshot không làm lộ state thật

Caller có thể sửa field của `Snapshot` vì các field đó exported. Điều cần bảo đảm là việc sửa bản sao không sửa session đã tạo ra nó.

### 8.1. Viết test độc lập

Thêm:

```go
func TestSnapshotIsIndependentCopy(t *testing.T) {
	sess := newTestSession(t)

	copyOfState := sess.Snapshot()
	copyOfState.Identity.ID = "changed"
	copyOfState.Identity.SourceIP = "198.51.100.5"
	copyOfState.CommandCount = 99
	copyOfState.Closed = true

	got := sess.Snapshot()
	if got.Identity.ID != "sess-test" {
		t.Fatalf("session ID changed through Snapshot: %q", got.Identity.ID)
	}
	if got.Identity.SourceIP != "192.0.2.10" {
		t.Fatalf("session source IP changed through Snapshot: %q", got.Identity.SourceIP)
	}
	if got.CommandCount != 0 {
		t.Fatalf("session command count changed through Snapshot: %d", got.CommandCount)
	}
	if got.Closed {
		t.Fatal("session was closed by changing Snapshot")
	}
}
```

Chạy:

```powershell
go test ./internal/session -run '^TestSnapshotIsIndependentCopy$' -v
```

Test phải PASS mà không cần sửa implementation.

Chạy toàn bộ suite trước khi tiếp tục:

```powershell
go test ./...
```

### 8.2. Vì sao không viết getter/setter cho mọi field?

API kiểu này không giúp model bảo vệ chính nó:

```text
GetID / SetID
GetSourceIP / SetSourceIP
GetCommandCount / SetCommandCount
GetClosed / SetClosed
```

Setter tổng quát cho caller quyền phá invariant. API hiện tại diễn đạt hành vi nghiệp vụ:

```go
sess.AddCommand()
sess.Close()
snapshot := sess.Snapshot()
```

Getter có chủ đích vẫn có thể hợp lý nếu một giá trị thật sự thuộc public contract. Câu trả lời không phải “cấm mọi getter”, mà là không sinh getter/setter máy móc như JavaBean.

### 8.3. Lưu ý về shallow copy

Snapshot hiện chứa string, int, bool và `time.Time`, nên việc copy rất dễ hiểu.

Nếu sau này struct chứa slice, map hoặc pointer, value copy chỉ copy header/địa chỉ tham chiếu. Hai giá trị có thể vẫn dùng chung dữ liệu bên dưới. Khi đó cần clone dữ liệu mutable một cách có chủ đích, giống bài `ClonePayload` ở Day 2.

### External test package

Test dùng:

```go
package session_test
```

Nó chỉ thấy public API giống một caller thật. Test không thể đi đường tắt bằng cách đọc `sess.commandCount` hay tự gán `sess.closed`.

### Kết quả chặng 6

- Sửa snapshot không đổi session thật.
- Test chỉ quan sát qua public API.
- Không có setter tổng quát cho state.
- Snapshot là bản sao có thể sửa, không phải object immutable.

---

## 9. Chặng 7 — Ghép domain model vào CLI hoàn chỉnh

Domain behavior đã được test. Bây giờ ta ghép thành một lát cắt chạy từ đầu đến cuối:

```text
main → NewSession → AddCommand → Close → Snapshot → terminal
```

Thay `cmd/sessionlab/main.go` bằng:

```go
package main

import (
	"fmt"
	"os"
	"time"

	"honeypot-day3/internal/session"
)

func main() {
	startedAt := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)
	sess, err := session.NewSession("sess-001", "203.0.113.10", startedAt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create session:", err)
		os.Exit(1)
	}

	commands := []string{
		"whoami",
		"wget http://example.com/a.sh",
	}

	for _, command := range commands {
		if err := sess.AddCommand(); err != nil {
			fmt.Fprintln(os.Stderr, "record command:", err)
			os.Exit(1)
		}
		fmt.Printf("recorded command: %q\n", command)
	}

	sess.Close()
	if err := sess.AddCommand(); err != nil {
		fmt.Println("rejected command after Close:", err)
	} else {
		fmt.Fprintln(os.Stderr, "expected command after Close to be rejected")
		os.Exit(1)
	}

	fmt.Println()
	printSnapshot(sess.Snapshot())
}

func printSnapshot(snapshot session.Snapshot) {
	fmt.Println("=== Session Snapshot ===")
	fmt.Printf("ID:            %s\n", snapshot.Identity.ID)
	fmt.Printf("Source IP:     %s\n", snapshot.Identity.SourceIP)
	fmt.Printf("Started at:    %s\n", snapshot.StartedAt.Format(time.RFC3339))
	fmt.Printf("Command count: %d\n", snapshot.CommandCount)
	fmt.Printf("Closed:        %t\n", snapshot.Closed)
}
```

Chạy:

```powershell
go run ./cmd/sessionlab
```

Đối chiếu output với yêu cầu ở đầu bài. Count phải là `2`, không phải `3`, vì command sau `Close` bị từ chối.

### Phân tích ranh giới trách nhiệm

Package `session` chịu trách nhiệm:

- Giữ state.
- Bảo vệ invariant.
- Thực hiện transition của lifecycle.
- Tạo snapshot.

Package `main` chịu trách nhiệm:

- Chọn input demo.
- Xử lý error.
- Điều phối các hành vi.
- In kết quả ra terminal.

CLI không đọc field private và domain package không biết gì về terminal. Hai phần có lý do thay đổi khác nhau.

> Hai chuỗi `whoami` và `wget ...` không được chạy. Project chỉ đếm việc nhận command giả lập.

### Kết quả chặng 7

Sau khi output đúng, kiểm tra cả module:

```powershell
go test ./...
go build ./...
```

- CLI tạo session hợp lệ.
- Hai command đầu được ghi nhận.
- Command thứ ba bị từ chối.
- Snapshot cuối có count `2` và `Closed: true`.
- Output khớp acceptance criteria.

---

## 10. Chặng 8 — Quality gate cuối project

Cấu trúc cuối cùng:

```text
go_day3_practice/          # Thư mục được xây tuần tự trong tutorial.
├── cmd/
│   └── sessionlab/
│       └── main.go
├── internal/
│   └── session/
│       ├── session.go
│       └── session_test.go
└── go.mod
```

Mỗi executable nằm trong một thư mục riêng dưới `cmd`. Không đặt nhiều file có `func main()` trong cùng một package, vì `go test ./...` và `go build ./...` sẽ báo `main redeclared`.

### 10.1. Chạy toàn bộ test

```powershell
go test -count=1 ./...
```

`-count=1` buộc Go chạy lại thay vì dùng kết quả cache.

### 10.2. Xem coverage

```powershell
go test -count=1 -cover ./...
```

Package `internal/session` của lời giải đạt 100% statement coverage. CLI chưa có test nên có thể hiện `0.0%`; không được diễn giải thành toàn module đạt 100%.

Coverage cao không chứng minh thiết kế hoàn hảo. Nó chỉ nói statement nào đã được chạy khi test.

### 10.3. Format, vet và build

```powershell
go fmt ./...
go vet ./...
go build ./...
```

- `go fmt` chuẩn hóa format.
- `go vet` tìm một số mẫu code đáng ngờ.
- `go build` xác nhận mọi package đều biên dịch.

### 10.4. Chạy sản phẩm lần cuối

```powershell
go run ./cmd/sessionlab
```

Quality gate chỉ hoàn thành khi tất cả lệnh trên thành công và output đúng.

---

## 11. Bản đồ kiến thức vừa dùng trong project

| Vấn đề project gặp | Công cụ Go đã dùng |
|---|---|
| Cần gom trạng thái session | `struct` |
| Caller phá state bằng cách gán field | Field unexported + package boundary |
| Cần tạo session hợp lệ | `NewSession(...)(*Session, error)` |
| Cần bảo vệ quy tắc luôn đúng | Invariant + validation |
| Cần tăng count trên object thật | Method với pointer receiver |
| Cần lấy một snapshot từ type nhỏ, copy-safe | `Snapshot()` dùng value receiver trong thiết kế này |
| Cần nhóm ID và source IP | Nested struct/composition |
| Cần diễn đạt hành vi | `AddCommand`, `Close` thay setter |
| Cần kiểm tra qua góc nhìn caller | External test package `session_test` |

OOP trong Go ở project này không cần class:

- Struct giữ trạng thái.
- Method mô tả hành vi.
- Package boundary tạo đóng gói.
- Composition nhóm các phần nhỏ.

Interface và đa hình sẽ xuất hiện ở Day 4, khi project có một consumer thật sự cần nhiều implementation.

Đến đây project đã hoàn tất. Phần tiếp theo cung cấp nguyên văn đáp án cuối để bạn đối chiếu file của mình.

---

## 12. Đáp án hoàn chỉnh

Cấu trúc đáp án:

```text
go_day3_practice/
├── cmd/
│   └── sessionlab/
│       └── main.go
├── internal/
│   └── session/
│       ├── session.go
│       └── session_test.go
└── go.mod
```

### 12.1. `go.mod`

```go
module honeypot-day3

go 1.26.5
```

### 12.2. `internal/session/session.go`

```go
// Package session models the lifecycle of a honeypot session.
package session

import (
	"errors"
	"time"
)

// Session keeps the mutable state of one honeypot connection.
// Its fields are unexported so state changes must go through its methods.
type Session struct {
	id           string
	sourceIP     string
	startedAt    time.Time
	commandCount int
	closed       bool
}

// Identity contains the stable identity of a session.
type Identity struct {
	ID       string
	SourceIP string
}

// Snapshot is a read-only-by-convention copy of a session's current state.
// Changing a Snapshot does not change the Session that produced it.
type Snapshot struct {
	Identity     Identity
	StartedAt    time.Time
	CommandCount int
	Closed       bool
}

// NewSession creates an open session with no recorded commands.
func NewSession(id, sourceIP string, now time.Time) (*Session, error) {
	if id == "" || sourceIP == "" {
		return nil, errors.New("id and source IP are required")
	}

	return &Session{
		id:        id,
		sourceIP:  sourceIP,
		startedAt: now,
	}, nil
}

// AddCommand records one command unless the session has already been closed.
func (s *Session) AddCommand() error {
	if s.closed {
		return errors.New("session is closed")
	}

	s.commandCount++
	return nil
}

// Close marks the session as closed. Calling Close more than once is safe.
func (s *Session) Close() {
	s.closed = true
}

// Snapshot returns a copy of the current session state.
func (s Session) Snapshot() Snapshot {
	return Snapshot{
		Identity: Identity{
			ID:       s.id,
			SourceIP: s.sourceIP,
		},
		StartedAt:    s.startedAt,
		CommandCount: s.commandCount,
		Closed:       s.closed,
	}
}
```

### 12.3. `internal/session/session_test.go`

```go
package session_test

import (
	"testing"
	"time"

	"honeypot-day3/internal/session"
)

func TestNewSessionInitializesValidState(t *testing.T) {
	now := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)

	sess, err := session.NewSession("sess-001", "203.0.113.10", now)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	got := sess.Snapshot()
	if got.Identity.ID != "sess-001" {
		t.Fatalf("ID = %q, want %q", got.Identity.ID, "sess-001")
	}
	if got.Identity.SourceIP != "203.0.113.10" {
		t.Fatalf("SourceIP = %q, want %q", got.Identity.SourceIP, "203.0.113.10")
	}
	if !got.StartedAt.Equal(now) {
		t.Fatalf("StartedAt = %v, want %v", got.StartedAt, now)
	}
	if got.CommandCount != 0 {
		t.Fatalf("CommandCount = %d, want 0", got.CommandCount)
	}
	if got.Closed {
		t.Fatal("new session must be open")
	}
}

func TestNewSessionRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		sourceIP string
	}{
		{name: "missing ID", sourceIP: "203.0.113.10"},
		{name: "missing source IP", id: "sess-001"},
		{name: "missing both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := session.NewSession(tt.id, tt.sourceIP, time.Now())
			if err == nil {
				t.Fatal("NewSession() error = nil, want an error")
			}
			if err.Error() != "id and source IP are required" {
				t.Fatalf("NewSession() error = %q, want %q", err, "id and source IP are required")
			}
			if got != nil {
				t.Fatalf("NewSession() session = %#v, want nil", got)
			}
		})
	}
}

func TestAddCommandIncrementsCount(t *testing.T) {
	sess := newTestSession(t)

	for want := 1; want <= 2; want++ {
		if err := sess.AddCommand(); err != nil {
			t.Fatalf("AddCommand() error = %v", err)
		}
		if got := sess.Snapshot().CommandCount; got != want {
			t.Fatalf("CommandCount = %d, want %d", got, want)
		}
	}
}

func TestClosedSessionRejectsNewCommand(t *testing.T) {
	sess := newTestSession(t)
	if err := sess.AddCommand(); err != nil {
		t.Fatalf("first AddCommand() error = %v", err)
	}

	sess.Close()
	err := sess.AddCommand()
	if err == nil {
		t.Fatal("AddCommand() error = nil after Close, want an error")
	}
	if err.Error() != "session is closed" {
		t.Fatalf("AddCommand() error = %q, want %q", err, "session is closed")
	}

	got := sess.Snapshot()
	if !got.Closed {
		t.Fatal("Closed = false after Close, want true")
	}
	if got.CommandCount != 1 {
		t.Fatalf("CommandCount = %d after rejected command, want 1", got.CommandCount)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	sess := newTestSession(t)

	sess.Close()
	sess.Close()

	if !sess.Snapshot().Closed {
		t.Fatal("Closed = false after two calls to Close, want true")
	}
}

func TestSnapshotIsIndependentCopy(t *testing.T) {
	sess := newTestSession(t)

	copyOfState := sess.Snapshot()
	copyOfState.Identity.ID = "changed"
	copyOfState.Identity.SourceIP = "198.51.100.5"
	copyOfState.CommandCount = 99
	copyOfState.Closed = true

	got := sess.Snapshot()
	if got.Identity.ID != "sess-test" {
		t.Fatalf("session ID changed through Snapshot: %q", got.Identity.ID)
	}
	if got.Identity.SourceIP != "192.0.2.10" {
		t.Fatalf("session source IP changed through Snapshot: %q", got.Identity.SourceIP)
	}
	if got.CommandCount != 0 {
		t.Fatalf("session command count changed through Snapshot: %d", got.CommandCount)
	}
	if got.Closed {
		t.Fatal("session was closed by changing Snapshot")
	}
}

func newTestSession(t *testing.T) *session.Session {
	t.Helper()

	sess, err := session.NewSession(
		"sess-test",
		"192.0.2.10",
		time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}
	return sess
}
```

### 12.4. `cmd/sessionlab/main.go`

```go
package main

import (
	"fmt"
	"os"
	"time"

	"honeypot-day3/internal/session"
)

func main() {
	startedAt := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)
	sess, err := session.NewSession("sess-001", "203.0.113.10", startedAt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create session:", err)
		os.Exit(1)
	}

	commands := []string{
		"whoami",
		"wget http://example.com/a.sh",
	}

	for _, command := range commands {
		if err := sess.AddCommand(); err != nil {
			fmt.Fprintln(os.Stderr, "record command:", err)
			os.Exit(1)
		}
		fmt.Printf("recorded command: %q\n", command)
	}

	sess.Close()
	if err := sess.AddCommand(); err != nil {
		fmt.Println("rejected command after Close:", err)
	} else {
		fmt.Fprintln(os.Stderr, "expected command after Close to be rejected")
		os.Exit(1)
	}

	fmt.Println()
	printSnapshot(sess.Snapshot())
}

func printSnapshot(snapshot session.Snapshot) {
	fmt.Println("=== Session Snapshot ===")
	fmt.Printf("ID:            %s\n", snapshot.Identity.ID)
	fmt.Printf("Source IP:     %s\n", snapshot.Identity.SourceIP)
	fmt.Printf("Started at:    %s\n", snapshot.StartedAt.Format(time.RFC3339))
	fmt.Printf("Command count: %d\n", snapshot.CommandCount)
	fmt.Printf("Closed:        %t\n", snapshot.Closed)
}
```

---

## 13. Chạy đáp án từ đầu đến cuối

Trong thư mục thực hành, chạy đúng thứ tự:

```powershell
go fmt ./...
go test -count=1 ./...
go vet ./...
go build ./...
go run ./cmd/sessionlab
```

Kết quả test:

```text
?    honeypot-day3/cmd/sessionlab       [no test files]
ok   honeypot-day3/internal/session
```

Kết quả chương trình:

```text
recorded command: "whoami"
recorded command: "wget http://example.com/a.sh"
rejected command after Close: session is closed

=== Session Snapshot ===
ID:            sess-001
Source IP:     203.0.113.10
Started at:    2026-08-13T09:00:00Z
Command count: 2
Closed:        true
```

Nếu output này xuất hiện và bốn lệnh kiểm tra không báo lỗi, bạn đã hoàn thành toàn bộ Day 3 từ lúc tạo module đến khi có domain model, unit test và CLI chạy được.

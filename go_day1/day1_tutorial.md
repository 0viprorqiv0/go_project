# Day 1 — Go toolchain, package và chương trình Honeypot đầu tiên

## 1. Mục tiêu của ngày học

Sau khi hoàn thành tutorial này, bạn sẽ:

- Hiểu cấu trúc cơ bản của một chương trình Go.
- Biết cách khởi tạo và sử dụng Go module.
- Hiểu `package`, `import`, tên exported và unexported.
- Biết khai báo biến, hằng số và nhận biết zero value.
- Biết viết function trả về nhiều giá trị.
- Viết được CLI nhận `source-ip`, `port`, `service`.
- Tạo một honeypot event giả và in ra terminal.
- Tách domain model sang package `internal/model`.
- Biết dùng `go run`, `go build`, `go test`, `go fmt` và `go vet`.

Thời gian đề xuất: **4 giờ**.

> Hôm nay chưa viết honeypot server thật. Mục tiêu là làm chủ quy trình tối thiểu từ source code đến một chương trình chạy được và có cấu trúc tốt.

---

## 2. Sản phẩm cuối ngày

Ta sẽ xây một CLI có cách dùng như sau:

```powershell
go run ./cmd/eventcli --source-ip 203.0.113.10 --port 23 --service telnet
```

Kết quả dự kiến:

```text
=== Fake Honeypot Event ===
ID:         evt-...
Source IP:  203.0.113.10
Port:       23
Service:    telnet
Type:       connection_attempt
Timestamp:  2026-08-06T09:00:00Z
```

Cấu trúc thư mục cuối bài:

```text
go_day1/
├── cmd/
│   └── eventcli/
│       └── main.go
├── internal/
│   └── model/
│       └── event.go
├── day1_tutorial.md
└── go.mod
```

Ta chỉ tạo những thư mục có code sử dụng ngay trong ngày đầu.

---

## 3. Kiểm tra môi trường

Mở PowerShell tại thư mục `C:\Go_project\go_day1`:

```powershell
cd C:\Go_project\go_day1
go version
```

Kết quả có dạng:

```text
go version go1.x.x windows/amd64
```

Kiểm tra thêm các lệnh:

```powershell
go env GOROOT
go env GOPATH
go help
```

- `GOROOT`: nơi Go toolchain được cài đặt.
- `GOPATH`: workspace/cache mặc định của Go. Khi dùng Go module, source code của dự án không bắt buộc phải nằm trong `GOPATH`.
- `go help`: liệt kê các lệnh mà Go toolchain cung cấp.

Nếu PowerShell báo không tìm thấy lệnh `go`, hãy cài Go và mở lại terminal để biến môi trường `PATH` được cập nhật.

---

## 4. Chương trình Go tối thiểu hoạt động như thế nào?

Ví dụ nhỏ nhất:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, honeypot!")
}
```

Giải thích từng phần:

### `package main`

Mọi file `.go` thuộc một package. Package giúp gom các type, function và biến có liên quan.

`package main` có ý nghĩa đặc biệt: nó tạo ra một chương trình thực thi được. Một package dùng để tái sử dụng thường có tên khác, ví dụ `model`, và không tự chạy trực tiếp.

### `import "fmt"`

`import` đưa package khác vào file hiện tại. Package chuẩn `fmt` cung cấp các function định dạng và in dữ liệu.

Go không chấp nhận import không được sử dụng. Quy tắc này giúp source code sạch và làm lỗi phụ thuộc được phát hiện ngay khi biên dịch.

### `func main()`

`main` là điểm bắt đầu của chương trình thuộc `package main`. Khi chạy executable, Go gọi function này.

### `fmt.Println(...)`

`Println` được export bởi package `fmt`, nên tên bắt đầu bằng chữ hoa. Function này in các giá trị, thêm khoảng trắng khi cần và kết thúc bằng ký tự xuống dòng.

---

## 5. Go module là gì?

Module là đơn vị quản lý source code và dependency của một dự án Go. File `go.mod` ghi tên module và phiên bản Go mà dự án sử dụng.

Tại `C:\Go_project\go_day1`, chạy:

```powershell
go mod init honeypot-lab
```

Go tạo file `go.mod` tương tự:

```go
module honeypot-lab

go 1.x.x
```

Trong tutorial này, `honeypot-lab` cũng là tiền tố import cho package nội bộ:

```go
import "honeypot-lab/internal/model"
```

Không chạy lại `go mod init` nếu `go.mod` đã tồn tại. Có thể kiểm tra module hiện tại bằng:

```powershell
go env GOMOD
go list -m
```

> Tên module dùng trong dự án học có thể ngắn như `honeypot-lab`. Với repository được chia sẻ công khai, tên module thường là đường dẫn repository, ví dụ `github.com/username/honeypot-lab`.

---

## 6. Package và quy tắc visibility

Go dùng chữ cái đầu tiên để quyết định identifier có được truy cập từ package khác hay không.

```go
type Event struct{} // Exported: package khác có thể dùng Event.
type event struct{} // Unexported: chỉ package hiện tại dùng được event.

func NewEvent() {} // Exported.
func newEvent() {} // Unexported.
```

Quy tắc này áp dụng cho type, function, method, field, biến và hằng số cấp package.

Ví dụ:

```go
type Event struct {
	ID       string // Package khác đọc/ghi được.
	internal string // Chỉ package khai báo Event truy cập được.
}
```

Trong bài hôm nay, CLI thuộc package `main`, còn `Event` thuộc package `model`. Do đó tên type `Event` và các field CLI cần đọc phải viết hoa.

### Package `internal` có gì đặc biệt?

Thư mục mang tên `internal` được Go bảo vệ ở mức toolchain. Chỉ code nằm trong cây thư mục cha của `internal` mới được import package bên trong nó.

Với cấu trúc hiện tại:

```text
go_day1/internal/model
```

code bên trong module `go_day1` có thể import `honeypot-lab/internal/model`, nhưng một dự án bên ngoài không thể tùy ý import package đó. Đây là cách phù hợp để giữ domain model nội bộ của ứng dụng.

---

## 7. Biến, hằng số và zero value

### Khai báo biến bằng `var`

```go
var sourceIP string
var port int
```

Khi chưa gán giá trị, biến nhận zero value của kiểu:

| Kiểu | Zero value |
|---|---|
| `string` | `""` |
| `int`, `int64`, `float64` | `0` |
| `bool` | `false` |
| pointer, slice, map, channel, interface | `nil` |
| `time.Time` | một giá trị thời gian zero; kiểm tra bằng `IsZero()` |

Go không để biến cục bộ ở trạng thái “rác bộ nhớ”. Tuy nhiên, zero value không phải lúc nào cũng hợp lệ về nghiệp vụ. Ví dụ port `0` không hợp lệ cho CLI của bài này, vì vậy ta vẫn phải validate.

### Khai báo ngắn bằng `:=`

Trong function, Go có thể suy luận kiểu:

```go
service := "telnet"
port := 23
```

`:=` chỉ dùng trong function. Khi không có biến mới nào ở bên trái, dùng `=` thay vì `:=`.

### Hằng số

```go
const eventTypeConnectionAttempt = "connection_attempt"
```

Hằng số dùng cho giá trị không thay đổi trong quá trình chạy. Trong bài, loại event là một hằng số unexported vì chỉ package `main` cần nó.

---

## 8. Function và multiple return values

Go cho phép một function trả về nhiều giá trị. Mẫu phổ biến nhất là `value, error`:

```go
func parsePort(raw string) (int, error) {
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid port %q: %w", raw, err)
	}

	return port, nil
}
```

Caller nhận cả hai giá trị:

```go
port, err := parsePort("23")
if err != nil {
	// Xử lý lỗi.
}
```

Không có exception cho lỗi nghiệp vụ thông thường. Lỗi là một giá trị và cần được kiểm tra rõ ràng.

Trong CLI, package `flag` đã parse `port` thành `int`, nhưng function đọc cấu hình vẫn trả `(config, error)`. Đây là cơ hội thực hành multiple return và tách validation khỏi `main`.

---

## 9. Bước 1 — Tạo domain model `Event`

Tạo thư mục:

```powershell
New-Item -ItemType Directory -Force internal\model
```

Tạo file `internal/model/event.go`:

```go
package model

import "time"

// Event mô tả một hoạt động được honeypot quan sát.
type Event struct {
	ID        string
	SourceIP  string
	Service   string
	Type      string
	Timestamp time.Time
}
```

### Vì sao model không chứa logic in terminal?

`Event` là dữ liệu domain. Việc đọc cờ dòng lệnh và in ra terminal thuộc tầng CLI. Nếu đưa `flag.Parse()` hoặc toàn bộ định dạng terminal vào package `model`, model sẽ bị gắn chặt với một giao diện đầu vào cụ thể.

### Vì sao `time.Time` tốt hơn `string`?

`time.Time` biểu diễn thời gian có cấu trúc, hỗ trợ so sánh, cộng/trừ và chuyển múi giờ. Chỉ chuyển thành chuỗi khi cần hiển thị hoặc serialize.

### Port nằm ở đâu?

Roadmap yêu cầu sản phẩm ngày đầu có `Event` đúng với năm field trên, trong khi CLI vẫn phải nhận `port`. Vì vậy tutorial coi port là metadata đầu vào của CLI và in nó cùng event. Khi model được mở rộng ở những ngày sau, ta có thể thêm `DestinationPort` hoặc đưa dữ liệu riêng từng giao thức vào `ServiceData`.

---

## 10. Bước 2 — Viết CLI

Tạo thư mục:

```powershell
New-Item -ItemType Directory -Force cmd\eventcli
```

Tạo file `cmd/eventcli/main.go`:

```go
package main // Package main tạo chương trình có thể chạy được.

import (
	"errors"  // Tạo lỗi đơn giản.
	"flag"    // Đọc tham số dòng lệnh.
	"fmt"     // In và định dạng chuỗi.
	"net"     // Kiểm tra địa chỉ IP.
	"os"      // Truy cập stderr và thoát chương trình.
	"strings" // Xử lý chuỗi.
	"time"    // Lấy và định dạng thời gian.

	// Import package model do chúng ta tự viết.
	"honeypot-lab/internal/model"
)

// Hằng số không thay đổi trong lúc chương trình chạy.
const eventTypeConnectionAttempt = "connection_attempt"

// config gom các tham số người dùng truyền vào CLI.
type config struct {
	sourceIP string
	port     int
	service  string
}

func main() {
	// parseConfig trả về hai giá trị: cấu hình và lỗi.
	cfg, err := parseConfig()

	// err khác nil nghĩa là đã xảy ra lỗi.
	if err != nil {
		// In lỗi ra luồng dành cho thông báo lỗi.
		fmt.Fprintln(os.Stderr, "error:", err)

		// In hướng dẫn sử dụng CLI.
		flag.Usage()

		// Kết thúc chương trình với mã thất bại.
		os.Exit(1)
	}

	// Tạo một event giả từ cấu hình.
	event := newFakeEvent(cfg)

	// In event ra terminal.
	printEvent(event, cfg.port)
}

func parseConfig() (config, error) {
	// Tạo config với zero value:
	// sourceIP="", port=0, service="".
	var cfg config

	// & lấy địa chỉ của field để package flag có thể ghi giá trị vào đó.
	flag.StringVar(
		&cfg.sourceIP,
		"source-ip",
		"",
		"source IP observed by the honeypot",
	)

	flag.IntVar(
		&cfg.port,
		"port",
		0,
		"destination port from 1 to 65535",
	)

	flag.StringVar(
		&cfg.service,
		"service",
		"",
		"emulated service, for example telnet",
	)

	// Đọc các tham số người dùng truyền từ terminal.
	flag.Parse()

	// Xóa khoảng trắng thừa ở đầu và cuối IP.
	cfg.sourceIP = strings.TrimSpace(cfg.sourceIP)

	// Xóa khoảng trắng và chuyển service thành chữ thường.
	cfg.service = strings.ToLower(strings.TrimSpace(cfg.service))

	// Kiểm tra người dùng có truyền source-ip không.
	if cfg.sourceIP == "" {
		// config{} là config rỗng.
		return config{}, errors.New("source-ip is required")
	}

	// ParseIP trả về nil nếu địa chỉ IP không hợp lệ.
	if net.ParseIP(cfg.sourceIP) == nil {
		return config{}, fmt.Errorf(
			"source-ip %q is not a valid IP address",
			cfg.sourceIP,
		)
	}

	// Port mạng phải nằm trong khoảng 1–65535.
	if cfg.port < 1 || cfg.port > 65535 {
		return config{}, fmt.Errorf(
			"port must be between 1 and 65535, got %d",
			cfg.port,
		)
	}

	// Kiểm tra service có bị bỏ trống không.
	if cfg.service == "" {
		return config{}, errors.New("service is required")
	}

	// nil nghĩa là không có lỗi.
	return cfg, nil
}

func newFakeEvent(cfg config) model.Event {
	// Lấy thời gian hiện tại và chuyển sang UTC.
	now := time.Now().UTC()

	// Khởi tạo và trả về một Event.
	return model.Event{
		// Dùng thời gian nanosecond để tạo ID tạm thời.
		ID: fmt.Sprintf("evt-%d", now.UnixNano()),

		SourceIP: cfg.sourceIP,
		Service:  cfg.service,
		Type:     eventTypeConnectionAttempt,

		// Dùng lại biến now để ID và Timestamp cùng thời điểm.
		Timestamp: now,
	}
}

func printEvent(event model.Event, port int) {
	fmt.Println("=== Fake Honeypot Event ===")

	// %s dùng để in string.
	fmt.Printf("ID:         %s\n", event.ID)
	fmt.Printf("Source IP:  %s\n", event.SourceIP)

	// %d dùng để in số nguyên.
	fmt.Printf("Port:       %d\n", port)

	fmt.Printf("Service:    %s\n", event.Service)
	fmt.Printf("Type:       %s\n", event.Type)

	// Chuyển time.Time thành chuỗi theo chuẩn RFC3339.
	fmt.Printf(
		"Timestamp:  %s\n",
		event.Timestamp.Format(time.RFC3339Nano),
	)
}
```

### Phân tích thiết kế

- `config` viết thường nên chỉ package `main` sử dụng được. Đây là chi tiết riêng của CLI.
- `parseConfig` trả về `(config, error)`, thể hiện cách xử lý lỗi quen thuộc trong Go.
- `flag.StringVar` và `flag.IntVar` ghi giá trị trực tiếp vào các field của `cfg` thông qua pointer.
- `strings.TrimSpace` loại khoảng trắng thừa; `strings.ToLower` chuẩn hóa tên service.
- `net.ParseIP` kiểm tra input có phải địa chỉ IPv4 hoặc IPv6 hợp lệ hay không.
- Port chỉ hợp lệ trong khoảng `1..65535`.
- `newFakeEvent` chỉ chịu trách nhiệm tạo event.
- `time.Now().UTC()` chuẩn hóa timestamp sang UTC, thuận tiện khi nhiều sensor chạy ở nhiều múi giờ.
- `printEvent` chỉ chịu trách nhiệm trình bày dữ liệu.
- `os.Stderr` dành cho thông báo lỗi; output thành công đi ra standard output.
- `os.Exit(1)` báo cho shell biết chương trình thất bại. Không dùng `panic` cho input CLI không hợp lệ.

### Lưu ý về ID

`evt-<UnixNano>` đủ cho bài học ngày đầu nhưng chưa phải chiến lược ID production-grade: nhiều tiến trình vẫn có khả năng tạo trùng ID. Các ngày sau có thể thay bằng UUID/ULID hoặc ID do hệ thống lưu trữ quản lý.

---

## 11. Bước 3 — Format code

Chạy từ thư mục chứa `go.mod`:

```powershell
go fmt ./...
```

`go fmt` chuẩn hóa indentation, dấu ngoặc và cách trình bày code. Go có một style formatter tiêu chuẩn nên team không cần tranh luận nhiều về format.

Kiểm tra các package được Go nhận diện:

```powershell
go list ./...
```

Kết quả dự kiến:

```text
honeypot-lab/cmd/eventcli
honeypot-lab/internal/model
```

---

## 12. Bước 4 — Chạy chương trình

### Xem hướng dẫn

```powershell
go run ./cmd/eventcli -h
```

### Chạy với input hợp lệ

```powershell
go run ./cmd/eventcli --source-ip 203.0.113.10 --port 23 --service telnet
```

Bạn cũng có thể thử IPv6:

```powershell
go run ./cmd/eventcli --source-ip 2001:db8::10 --port 8080 --service http
```

> Các dải `203.0.113.0/24` và `2001:db8::/32` được dùng làm địa chỉ tài liệu, phù hợp cho ví dụ thay vì sử dụng địa chỉ thật của một bên thứ ba.

### Thử input lỗi

Thiếu IP:

```powershell
go run ./cmd/eventcli --port 23 --service telnet
```

IP sai:

```powershell
go run ./cmd/eventcli --source-ip not-an-ip --port 23 --service telnet
```

Port ngoài phạm vi:

```powershell
go run ./cmd/eventcli --source-ip 203.0.113.10 --port 70000 --service telnet
```

Mỗi lệnh phải in thông báo dễ hiểu và kết thúc với exit code khác `0`.

Trong PowerShell, xem exit code gần nhất bằng:

```powershell
$LASTEXITCODE
```

---

## 13. Bước 5 — Build executable

`go run` biên dịch vào vị trí tạm rồi chạy ngay. `go build` tạo executable để chạy nhiều lần mà không cần biên dịch lại.

Tạo thư mục output và build:

```powershell
New-Item -ItemType Directory -Force bin
go build -o .\bin\eventcli.exe .\cmd\eventcli
```

Chạy file đã build:

```powershell
.\bin\eventcli.exe --source-ip 203.0.113.10 --port 23 --service telnet
```

Build toàn bộ package mà không cần giữ executable:

```powershell
go build ./...
```

Nếu lệnh này thành công và không in gì thì đó là bình thường.

Không commit thư mục `bin` nếu repository chỉ cần lưu source code. Có thể thêm `bin/` vào `.gitignore` khi bắt đầu dùng Git cho module này.

---

## 14. Bước 6 — Test và kiểm tra tĩnh

Chạy:

```powershell
go test ./...
go vet ./...
```

Ở ngày đầu chưa bắt buộc có test file, nên `go test` có thể báo `[no test files]`. Lệnh vẫn hữu ích vì nó biên dịch các package trong chế độ test và phát hiện nhiều lỗi cấu trúc.

`go vet` kiểm tra các mẫu code đáng ngờ mà compiler vẫn có thể chấp nhận, ví dụ một số lỗi format string. `vet` không thay thế test và cũng không chứng minh chương trình hoàn toàn đúng.

Quy trình kiểm tra nhanh nên chạy sau mỗi thay đổi:

```powershell
go fmt ./...
go test ./...
go vet ./...
go build ./...
```

---

## 15. Compile error và runtime error

### Compile error

Compile error xảy ra trước khi chương trình chạy. Ví dụ đổi:

```go
fmt.Println("hello")
```

thành:

```go
fmt.Prinln("hello")
```

`Prinln` không tồn tại, nên compiler dừng và chỉ ra file cùng dòng lỗi.

Các compile error thường gặp trong Go:

- Import package nhưng không sử dụng.
- Khai báo biến cục bộ nhưng không sử dụng.
- Gọi function không tồn tại.
- Gán sai kiểu.
- Truy cập tên unexported từ package khác.
- Package trong cùng một thư mục có tên không nhất quán.

### Runtime error

Runtime error xảy ra sau khi chương trình đã biên dịch và đang chạy. Ví dụ:

```go
numbers := []int{1, 2}
fmt.Println(numbers[10])
```

Code biên dịch được nhưng panic khi truy cập ngoài phạm vi slice.

Input không hợp lệ trong CLI của bài này **không phải** compile error. Nó là lỗi được chương trình phát hiện lúc chạy và xử lý bằng `error`, thông báo rõ ràng, rồi trả exit code `1`. Đây tốt hơn việc để chương trình panic.

---

## 16. Bài tập thực hành mở rộng

Hãy tự làm trước khi xem gợi ý.

### Bài 1 — Thêm cờ `event-type`

Thêm cờ tùy chọn:

```text
--event-type connection_attempt
```

Nếu người dùng không truyền, dùng `connection_attempt` làm mặc định. Chuẩn hóa input bằng `TrimSpace` và `ToLower`.

### Bài 2 — Giới hạn service

Chỉ cho phép ba service của lab:

- `telnet`
- `http`
- `ssh`

Viết function:

```go
func validateService(service string) error
```

Không đặt toàn bộ validation trong `main`.

### Bài 3 — Trả thêm giá trị từ function

Viết function suy ra tên service mặc định theo port:

```go
func serviceFromPort(port int) (string, bool)
```

Quy ước:

- `22` → `ssh`
- `23` → `telnet`
- `80` hoặc `8080` → `http`
- Port khác → `"", false`

Qua bài này, hãy giải thích vì sao `bool` thứ hai cần thiết. Nếu chỉ trả chuỗi rỗng, caller không biết đó là giá trị hợp lệ hay trạng thái “không tìm thấy”.

### Bài 4 — Quan sát zero value

Tạo một function thử nghiệm:

```go
func printZeroValues() {
	var name string
	var count int
	var enabled bool
	var createdAt time.Time

	fmt.Printf("name=%q count=%d enabled=%t createdAtIsZero=%t\n",
		name, count, enabled, createdAt.IsZero())
}
```

Gọi tạm function, quan sát kết quả, rồi xóa lời gọi để output chính của CLI không bị nhiễu.

### Bài 5 — Thử phá visibility

Tạm đổi `SourceIP` trong `model.Event` thành `sourceIP`, sau đó chạy:

```powershell
go build ./...
```

Đọc compiler error, giải thích nguyên nhân, rồi đổi lại `SourceIP`.

---

## 17. Các lỗi thường gặp

### `go: go.mod file not found`

Bạn đang đứng sai thư mục hoặc chưa chạy `go mod init`. Chuyển về thư mục module:

```powershell
cd C:\Go_project\go_day1
```

### `package honeypot-lab/internal/model is not in std`

Kiểm tra:

- `go.mod` có dòng `module honeypot-lab`.
- File model nằm đúng tại `internal/model/event.go`.
- Lệnh được chạy trong cây thư mục của module.

### `found packages ... in the same directory`

Các file `.go` trong cùng một thư mục phải thuộc cùng một package, ngoại trừ file test ngoài package có hậu tố `_test`. Không đặt `package main` và `package model` cạnh nhau trong cùng thư mục.

### `imported and not used`

Xóa import chưa dùng hoặc sử dụng đúng package. Đừng dùng blank import chỉ để che lỗi trong bài cơ bản.

### `declared and not used`

Go không cho phép biến cục bộ bị khai báo rồi bỏ quên. Hãy dùng biến đúng mục đích hoặc xóa nó.

### Cờ CLI không được nhận

Với package `flag`, cờ phải đứng trước positional argument đầu tiên. Trong bài này không cần positional argument, vì vậy hãy truyền toàn bộ input dưới dạng `--tên-cờ giá-trị`.

---

## 18. Checklist hoàn thành Day 1

- [ ] `go version` chạy thành công.
- [ ] `go.mod` khai báo module `honeypot-lab`.
- [ ] `Event` nằm trong `internal/model`.
- [ ] CLI nằm trong `cmd/eventcli`.
- [ ] CLI nhận đủ `source-ip`, `port`, `service`.
- [ ] IP, port và service rỗng được validate.
- [ ] Event sử dụng timestamp UTC.
- [ ] Input hợp lệ in ra event giả.
- [ ] Input lỗi không làm chương trình panic.
- [ ] `go fmt ./...` chạy thành công.
- [ ] `go test ./...` chạy thành công.
- [ ] `go vet ./...` chạy thành công.
- [ ] `go build ./...` chạy thành công từ thư mục gốc module.

---

## 19. Câu hỏi tự kiểm tra

1. Vì sao `main` phải nằm trong `package main` để tạo executable?
2. `go run` khác `go build` ở điểm nào?
3. Vì sao `Event` viết hoa nhưng `config` viết thường?
4. Package bên ngoài có truy cập được field `sourceIP` viết thường không?
5. Zero value của `string`, `int` và `bool` là gì?
6. Vì sao port có zero value là `0` nhưng ta vẫn báo lỗi?
7. `(config, error)` thể hiện multiple return values như thế nào?
8. Tại sao lỗi input không nên xử lý bằng `panic`?
9. Vì sao domain model không nên gọi `flag.Parse()`?
10. Vì sao cần chạy lệnh build từ thư mục gốc module?

### Đáp án ngắn

1. Go toolchain coi `package main` có `func main()` là entry point của executable.
2. `go run` biên dịch tạm rồi chạy; `go build` biên dịch package và có thể tạo executable để dùng lại.
3. `Event` cần dùng từ package `main`; `config` chỉ là chi tiết nội bộ của CLI.
4. Không. Identifier bắt đầu bằng chữ thường là unexported.
5. Lần lượt là `""`, `0`, `false`.
6. Zero value hợp lệ ở cấp ngôn ngữ nhưng không thỏa quy tắc nghiệp vụ của CLI.
7. Function trả đồng thời cấu hình thành công và một giá trị lỗi để caller kiểm tra.
8. Input sai là tình huống dự kiến; chương trình nên báo lỗi có kiểm soát và trả exit code phù hợp.
9. Làm vậy khiến domain phụ thuộc vào giao diện dòng lệnh và khó tái sử dụng.
10. Để Go tìm đúng `go.mod`, resolve import theo module và kiểm tra toàn bộ package bằng `./...`.

---

## 20. Definition of Done

Day 1 chỉ hoàn thành khi bạn tự thực hiện được chuỗi lệnh sau từ `C:\Go_project\go_day1`:

```powershell
go fmt ./...
go test ./...
go vet ./...
go build ./...
go run ./cmd/eventcli --source-ip 203.0.113.10 --port 23 --service telnet
```

Ngoài việc chương trình chạy, bạn phải tự giải thích được:

- Chữ hoa quyết định khả năng truy cập từ package khác.
- Compile error xảy ra trước khi chạy; runtime error xảy ra trong lúc chạy.
- `go.mod` xác định ranh giới và import path của module.
- Tách `model` khỏi `main` giúp domain không bị gắn chặt với CLI.

Khi đã làm được những điều trên, bạn có nền móng để sang Day 2: xử lý kiểu dữ liệu, slice, map, string và byte — các thành phần xuất hiện liên tục khi honeypot nhận và phân tích payload mạng.

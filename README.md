# 🔐 CipherScope

CipherScope is a **distributed and concurrent hash cracking system** written in Go.  
It demonstrates practical concepts of **parallel computing and distributed systems**, combining:

- Multi-core processing (goroutines)
- Distributed execution across multiple machines (e.g., TV Boxes)
- Task distribution via TCP sockets
- Efficient worker coordination and early termination

## 📌 Overview

CipherScope supports two main attack strategies:

### 1. Dictionary Attack
- Reads passwords from a wordlist file
- Hashes each entry and compares with the target

### 2. Brute Force Attack
- Generates all possible combinations from a given charset and length
- Uses recursion + streaming (no full memory allocation)

## 🧠 Architecture

### 🔹 Hybrid Parallelism

CipherScope combines:

- **Intra-node parallelism**
  - Goroutines + channels
  - Utilizes all CPU cores

- **Inter-node parallelism**
  - Multiple workers across devices
  - Master distributes workload

### 🔹 System Components

CipherScope/
├── master/        # Task coordinator (distributes work)
├── worker/        # Executes tasks (runs on remote machines)
├── core/          # Engine, jobs, and task definitions
├── attacks/       # Dictionary and brute-force logic
├── crypto/        # Hashing algorithms
├── utils/         # Logging and helpers
├── main.go        # Local (non-distributed) execution

### 🔹 Execution Flow
       ┌────────────┐
       │   Master   │
       └─────┬──────┘
             │
    ┌────────┼────────┐
    │        │        │
┌────▼───┐ ┌──▼────┐ ┌─▼─────┐
│ Worker │ │ Worker│ │ Worker│
└────┬───┘ └──┬────┘ └──┬────┘
│         │         │
┌────▼─────────▼─────────▼────┐
│   Parallel Hash Processing  │
└─────────────────────────────┘

## ⚙️ Requirements

- Go 1.20+
- Machines in the same network (for distributed mode)

## 🔧 Installation

Clone the repository:

```bash
git clone https://github.com/your-username/cipherscope.git
cd cipherscope
````

## 🚀 Usage

## 🖥️ 1. Run Master

```bash
go run master/main.go
```

You will be prompted:

Enter hash:
Enter hash type (md5, sha1, sha256):

## 📺 2. Configure Workers

Edit:

📁 `worker/main.go`

Replace:

```go
"net.Dial("tcp", "MASTER_IP:9000")
```

With your master IP:

```go
"net.Dial("tcp", "192.168.0.10:9000")
```

## ▶️ 3. Run Workers

### Option A: Local testing

```bash
go run worker/main.go
```

(Open multiple terminals)

### Option B: TV Boxes / Remote Devices

#### Build for ARM:

```bash
GOOS=linux GOARCH=arm64 go build -o worker ./worker
```

#### Transfer:

```bash
scp worker user@DEVICE_IP:/home/user/
```

#### Run:

```bash
chmod +x worker
./worker
```

## 🧪 Example

Generate a test hash:

```bash
echo -n "abcde" | md5sum
```

Run master and input:

```
Hash: <generated hash>
Type: md5
```

Expected output:

```
Password found: abcde
All workers stopped.
```

---

## 📊 Performance

### Time Complexity

* Dictionary Attack:

  ```
  O(n)
  ```

* Brute Force:

  ```
  O(|charset|^length)
  ```

---

### Distributed Speedup

```
T ≈ N / (workers × CPU cores)
```

Where:

* N = total combinations
* workers = number of machines

---

## ⚠️ Limitations

* No support for salted hashes (e.g., bcrypt, argon2)
* Brute force grows exponentially
* No dynamic load balancing (static distribution)
* No fault tolerance (worker failure not handled)

---

## 🧠 Future Improvements

* 🔄 Dynamic task queue (work stealing)
* 📊 Real-time metrics (hashes/sec)
* 🔐 Support for bcrypt / argon2
* ⚡ GPU acceleration
* 🌍 gRPC instead of raw TCP
* 🖥️ CLI interface (cobra)
* 📈 Benchmarking tools

---

## 🧪 Testing Tips

Use small values for quick tests:

```go
charset := "abc"
length := 3
```

---

## 📚 Concepts Demonstrated

* Goroutines and channels
* Producer–consumer pattern
* Context cancellation
* TCP networking in Go
* Distributed task scheduling
* Parallel brute force search

---

## 📄 License

MIT License

---

## 👨‍💻 Author

Developed for academic purposes in **Parallel Computing**.

---

## ⭐ Final Notes

CipherScope is a **practical demonstration of distributed brute-force computation**, showing how:

* Workloads can be split across machines
* CPU cores can be fully utilized
* Systems can scale horizontally

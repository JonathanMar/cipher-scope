<div align="center">

<h1>🔐 Cipher Scope</h1>

<p>Hash cracker distribuído com dashboard web em tempo real.<br>
Escrito em Go puro — roda em qualquer máquina, incluindo <strong>TV Box (ARM)</strong>.</p>

<img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white"/>
<img src="https://img.shields.io/badge/Platform-Linux%20%7C%20ARM%20%7C%20Android-informational?style=flat-square"/>
<img src="https://img.shields.io/badge/License-MIT-green?style=flat-square"/>

</div>

---

## ✨ Funcionalidades

- **Dashboard Web** — interface dark mode com progresso em tempo real via SSE
- **Gerador de Hash** — gera MD5, SHA1 e SHA256 diretamente no browser
- **Dictionary Attack** — lê wordlist em streaming (baixo uso de RAM)
- **Brute Force Attack** — tenta combinações de 1 até N caracteres com charset configurável
- **Modo Auto** — dictionary primeiro, brute force como fallback
- **Modo Distribuído** — master/worker via TCP para usar múltiplas máquinas em paralelo
- **Cross-platform** — compila para ARM32, ARM64, x86, Windows, macOS

---

## 🚀 Como Usar

### Pré-requisitos

```bash
# Go 1.22+
go version
```

### Compilar e rodar

```bash
git clone https://github.com/JonathanMar/cipher-scope.git
cd cipher-scope

# Rodar localmente (abre o browser automaticamente)
go run .

# Compilar binário
go build -o cipher-scope .
./cipher-scope
```

O dashboard abre em `http://localhost:8080`.

### Flags disponíveis

```
-addr      string   Endereço de escuta (default ":8080")
-no-browser         Não abrir o browser automaticamente
```

### Acesso pela rede (TV Box / outro dispositivo)

```bash
# Na TV Box ou servidor:
./cipher-scope -addr 0.0.0.0:8080 -no-browser

# No celular ou PC na mesma rede, abra:
# http://<IP-DA-TVBOX>:8080
```

---

## 📦 Cross-compile para TV Box (ARM)

```bash
# ARM64 (TV Boxes modernas: Amlogic S905X3+, Rockchip RK3318...)
GOOS=linux GOARCH=arm64 go build -o cipher-scope-arm64 .

# ARM32 (TV Boxes antigas 32-bit)
GOOS=linux GOARCH=arm GOARM=7 go build -o cipher-scope-arm32 .

# Copiar para a TV Box via SCP
scp cipher-scope-arm64 user@192.168.x.x:/home/user/
```

---

## 🌐 Dashboard

| Painel | Funcionalidade |
|--------|---------------|
| **Gerador de Hash** | Digite uma senha → copia o hash com 1 clique |
| **Configuração** | Hash alvo, tipo, modo de ataque, workers, tamanho máximo, charset |
| **Anel de Progresso** | Progresso visual em tempo real (SVG animado) |
| **Stats** | H/s, ETA, tempo decorrido, tentativas |
| **Resultado** | Banner de sucesso/falha com a senha encontrada |
| **Log** | Histórico de eventos com timestamp |

---

## 🔌 API REST

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `POST` | `/api/hash` | Gera hash de uma senha |
| `POST` | `/api/crack` | Inicia um ataque |
| `POST` | `/api/stop` | Para o ataque em andamento |
| `GET`  | `/api/progress` | SSE — progresso em tempo real |

### Exemplo: gerar hash

```bash
curl -X POST http://localhost:8080/api/hash \
  -H 'Content-Type: application/json' \
  -d '{"password":"hello","type":"md5"}'

# {"hash":"5d41402abc4b2a76b9719d911017c592"}
```

### Exemplo: iniciar ataque

```bash
curl -X POST http://localhost:8080/api/crack \
  -H 'Content-Type: application/json' \
  -d '{
    "hash":      "5d41402abc4b2a76b9719d911017c592",
    "hashType":  "md5",
    "mode":      "auto",
    "workers":   4,
    "maxLength": 5,
    "charset":   "abcdefghijklmnopqrstuvwxyz0123456789",
    "wordlist":  "wordlist.txt"
  }'
```

---

## 🖧 Modo Distribuído (Master / Worker)

Para distribuir o brute force entre várias máquinas na rede:

```bash
# Máquina principal (master) — aguarda workers e distribui tarefas
go run ./master -addr :9000

# Cada TV Box / máquina worker
go run ./worker -master 192.168.0.10:9000 -workers 4
```

O master divide o espaço de combinações entre os workers automaticamente.

---

## 🗂️ Estrutura do Projeto

```
cipher-scope/
├── main.go           # Servidor HTTP + embed da UI
├── ui/               # Dashboard (HTML/CSS/JS) — embutido no binário
├── server/           # Handlers HTTP, SSE e estado do ataque
├── attacks/          # Dictionary e Brute Force
├── core/             # Engine de workers, validação, tipos
├── crypto/           # MD5, SHA1, SHA256
├── master/           # Nó master do modo distribuído
├── worker/           # Nó worker do modo distribuído
└── utils/            # Logger
```

---

## 🔧 Hashes suportados

| Tipo | Tamanho |
|------|---------|
| MD5 | 32 chars |
| SHA1 | 40 chars |
| SHA256 | 64 chars |

---

## 📄 Licença

MIT © [JonathanMar](https://github.com/JonathanMar)

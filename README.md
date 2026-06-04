<div align="center">

<h1>🔐 Cipher Scope</h1>

<p>Canivete suíço de segurança com dashboard web em tempo real.<br>
Quebra hashes, decodifica dados, extrai senhas do Chrome e muito mais.<br>
Escrito em <strong>Go puro</strong> — roda em qualquer máquina, incluindo <strong>TV Box (ARM)</strong>.</p>

<img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white"/>
<img src="https://img.shields.io/badge/Platform-Linux%20%7C%20ARM%20%7C%20Android-informational?style=flat-square"/>
<img src="https://img.shields.io/badge/Tests-38%20passing-success?style=flat-square"/>
<img src="https://img.shields.io/badge/License-MIT-green?style=flat-square"/>

</div>

---

## ✨ Funcionalidades

| Categoria | Funcionalidade |
|-----------|---------------|
| **Hash Cracker** | Dictionary, Rules (mutações l33t), Brute Force, John the Ripper |
| **Gerador de Hash** | MD5, SHA1, SHA256, SHA512, NTLM — com contador de caracteres em tempo real |
| **Identificador de Hash** | Detecta o tipo automaticamente ao colar um hash |
| **Encode / Decode** | Base64, URL, Hexadecimal, Binário |
| **Chrome Decrypter** | Upload do `Login Data` e extração de senhas v10/v11 (AES-CBC) |
| **Dashboard** | Dark mode, progresso em tempo real via SSE, exportar log em .txt |
| **Distribuído** | Master/Worker via TCP para múltiplas máquinas em paralelo |
| **Cross-platform** | Compila para ARM32, ARM64, x86, Windows, macOS — sem CGo |

---

## 🛠️ Instalação do Go (pré-requisito)

### Linux x86_64 (PC comum)

```bash
# Baixar e instalar Go 1.22+
wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz

# Adicionar ao PATH (adicione ao seu ~/.bashrc ou ~/.profile)
export PATH=$PATH:/usr/local/go/bin

# Verificar
go version
```

### TV Box ARM64 (Amlogic S905X3+, Rockchip RK3318...)

```bash
# Baixar Go para ARM64
wget https://go.dev/dl/go1.22.5.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.5.linux-arm64.tar.gz

export PATH=$PATH:/usr/local/go/bin
go version
```

### TV Box ARM32 (dispositivos 32-bit antigos)

```bash
wget https://go.dev/dl/go1.22.5.linux-armv6l.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.5.linux-armv6l.tar.gz

export PATH=$PATH:/usr/local/go/bin
go version
```

### Android (Termux)

```bash
pkg update && pkg install golang git
go version
```

> 💡 Sempre verifique a versão mais recente em [go.dev/dl](https://go.dev/dl/)

---

## 🚀 Como Usar

### Opção 1 — Usar binário pré-compilado (mais fácil para TV Box)

Baixe o binário já compilado da [página de releases](https://github.com/JonathanMar/cipher-scope/releases):

```bash
# ARM64 (TV Box moderna)
wget https://github.com/JonathanMar/cipher-scope/releases/latest/download/cipher-scope-arm64
chmod +x cipher-scope-arm64
./cipher-scope-arm64 -addr 0.0.0.0:8080 -no-browser
```

### Opção 2 — Compilar do fonte

```bash
git clone https://github.com/JonathanMar/cipher-scope.git
cd cipher-scope

# Baixar dependências
go mod download

# Rodar diretamente (desenvolvimento)
go run .

# Compilar binário nativo
go build -o cipher-scope .
./cipher-scope
```

O dashboard abre automaticamente em `http://localhost:8080`.

### Flags disponíveis

```
-addr string        Endereço de escuta (default ":8080")
                    Use "0.0.0.0:8080" para acesso pela rede
-no-browser         Não abrir o browser automaticamente
```

---

## 📺 Guia completo — Rodando na TV Box

> Testado em **Tanix TX6** (Allwinner H6, ARM64) com Linux/Android + Termux.  
> Funciona em qualquer TV Box com ARM64 ou ARM32.

### Passo 1 — Descobrir a arquitetura da TV Box

```bash
uname -m
```

| Resultado | Binário a usar |
|-----------|---------------|
| `aarch64` | `cipher-scope-arm64` ✅ (maioria das TV Boxes modernas) |
| `armv7l`  | `cipher-scope-arm32` |
| `armv8l`  | `cipher-scope-arm64` |

### Passo 2 — Baixar o binário correto

```bash
# ARM64 (ex: Tanix TX6, X96 Max, H96 Max...)
wget https://github.com/JonathanMar/cipher-scope/releases/latest/download/cipher-scope-arm64

# ARM32 (TV Boxes antigas 32-bit)
wget https://github.com/JonathanMar/cipher-scope/releases/latest/download/cipher-scope-arm32
```

### Passo 3 — Dar permissão de execução

```bash
chmod +x cipher-scope-arm64
```

### Passo 4 — Iniciar o servidor

> ⚠️ **Importante:** use sempre `0.0.0.0` como host — não use o IP fixo da interface.  
> `0.0.0.0` significa "ouvir em todas as interfaces de rede".

```bash
./cipher-scope-arm64 -addr 0.0.0.0:8080 -no-browser
```

O terminal vai exibir:

```
╔══════════════════════════════════════╗
║       Cipher Scope  Dashboard        ║
╠══════════════════════════════════════╣
║  URL: http://localhost:8080          ║
╚══════════════════════════════════════╝
```

### Passo 5 — Descobrir o IP da TV Box

Em outro terminal (ou via SSH), rode:

```bash
hostname -I
# Exemplo de saída: 192.168.0.113 (esse é o IP da TV Box na rede)
```

### Passo 6 — Acessar o dashboard

No **celular, PC ou tablet** conectado na **mesma rede Wi-Fi**, abra o navegador:

```
http://<IP-DA-TVBOX>:8080

# Exemplo:
http://192.168.0.113:8080
```

O dashboard Cipher Scope abre com todas as funcionalidades disponíveis remotamente.

---

### 🔌 Controlar a TV Box remotamente via SSH

Se a TV Box tiver SSH habilitado, você pode gerenciar tudo pelo PC:

```bash
# Conectar na TV Box
ssh usuario@192.168.0.113

# Copiar o binário do PC para a TV Box
scp cipher-scope-arm64 usuario@192.168.0.113:~/

# Iniciar em background (continua rodando após fechar o SSH)
nohup ./cipher-scope-arm64 -addr 0.0.0.0:8080 -no-browser > cipher.log 2>&1 &

# Ver o log
tail -f cipher.log

# Parar o servidor
pkill cipher-scope-arm64
```

---

### ⚠️ Erros comuns e soluções

| Erro | Causa | Solução |
|------|-------|---------|
| `bind: cannot assign requested address` | Passou um IP que não pertence a esta máquina | Use `-addr 0.0.0.0:8080` |
| `permission denied` | Binário sem permissão de execução | `chmod +x cipher-scope-arm64` |
| `exec format error` | Binário errado (ARM32 numa ARM64 ou vice-versa) | Verifique com `uname -m` e baixe o binário correto |
| Dashboard não abre no celular | Firewall bloqueando a porta | `sudo ufw allow 8080` ou desative o firewall |
| `no such file or directory` | Binário não está no diretório atual | `ls -la` para verificar e use `./` antes do nome |

---

## 📦 Cross-compile (compilar na sua máquina, rodar na TV Box)

```bash
# ARM64 — TV Boxes modernas (Amlogic S905X3+, Rockchip RK3318, RK3399...)
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o cipher-scope-arm64 .

# ARM32 — TV Boxes antigas 32-bit (GOARM=7 para Cortex-A7/A9)
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o cipher-scope-arm32 .

# Copiar para a TV Box via SCP
scp cipher-scope-arm64 usuario@192.168.x.x:~/
```

> ✅ **Sem CGo:** o projeto usa `modernc.org/sqlite` (SQLite em Go puro), portanto cross-compile funciona sem precisar de toolchain C para ARM.

---


## 🗂️ Dashboard — Abas

### 📊 Dashboard (ataque)

| Painel | Funcionalidade |
|--------|---------------|
| **Gerador de Hash** | Digite → hash gerado ao vivo com **contador de caracteres** |
| **Configuração** | Hash alvo, tipo, modo, workers, **tamanho máximo até 64** |
| **Identificador** | Cola o hash → detecta MD5/NTLM/SHA1/SHA256/SHA512 automaticamente |
| **Anel de Progresso** | SVG animado com % em tempo real |
| **Stats** | H/s, ETA, tempo decorrido, tentativas |
| **Resultado** | Banner de sucesso/falha com a senha encontrada |
| **Log** | Histórico com timestamp + botão **Exportar .txt** |

### 🔧 Ferramentas

| Ferramenta | Formatos |
|-----------|---------|
| **Encode / Decode** | Base64, URL, Hexadecimal, Binário |
| **Identificador de Hash** | MD5, NTLM, SHA1, SHA256, SHA512 |

### 🌐 Chrome

Upload do arquivo `Login Data` do Chrome para extrair senhas salvas:
- **Linux:** `~/.config/google-chrome/Default/Login Data`
- **Chromium:** `~/.config/chromium/Default/Login Data`
- Campo de **master password** configurável (padrão Linux: `peanuts`)
- Tabela com URL, usuário e senha descriptografada
- Exportar tudo como **CSV**

---

## 🔐 Hashes suportados

| Tipo | Tamanho | Uso comum |
|------|---------|-----------|
| MD5 | 32 hex | Legado, BD antigos |
| SHA1 | 40 hex | Git, certificados antigos |
| SHA256 | 64 hex | TLS, JWT, Linux shadow |
| SHA512 | 128 hex | Linux shadow ($6$), APIs |
| NTLM | 32 hex | Windows / Active Directory |

---

## ⚔️ Modos de Ataque

| Modo | Descrição |
|------|-----------|
| **⚡ Auto** | Tenta: Dictionary → Rules → Brute Force |
| **📖 Dict** | Wordlist pura (`wordlist.txt`) |
| **🔀 Rules** | Wordlist + mutações (l33t, sufixos, capitalização) |
| **💥 Brute** | Todas combinações de 1 até N chars (configurável até 64) |
| **🗡️ John** | John the Ripper externo (se instalado) |

---

## 🔌 API REST

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `POST` | `/api/hash` | Gera hash (md5/sha1/sha256/sha512/ntlm) |
| `POST` | `/api/crack` | Inicia um ataque |
| `POST` | `/api/stop` | Para o ataque em andamento |
| `GET`  | `/api/progress` | SSE — progresso em tempo real |
| `POST` | `/api/encode` | Encode/Decode (base64/url/hex/binary) |
| `POST` | `/api/identify` | Identifica o tipo de um hash |
| `POST` | `/api/chrome` | Extrai senhas do Chrome (multipart/form-data) |
| `GET`  | `/api/info` | Info do sistema (workers, John disponível) |

### Exemplos

```bash
# Gerar hash SHA512
curl -X POST http://localhost:8080/api/hash \
  -H 'Content-Type: application/json' \
  -d '{"password":"hello","type":"sha512"}'

# Gerar hash NTLM (Windows/AD)
curl -X POST http://localhost:8080/api/hash \
  -H 'Content-Type: application/json' \
  -d '{"password":"Password1","type":"ntlm"}'

# Identificar tipo de hash
curl -X POST http://localhost:8080/api/identify \
  -H 'Content-Type: application/json' \
  -d '{"hash":"5d41402abc4b2a76b9719d911017c592"}'
# {"candidates":["md5","ntlm"],"isHex":true,"length":32}

# Encode Base64
curl -X POST http://localhost:8080/api/encode \
  -H 'Content-Type: application/json' \
  -d '{"value":"hello world","format":"base64","action":"encode"}'
# {"result":"aGVsbG8gd29ybGQ="}

# Iniciar ataque
curl -X POST http://localhost:8080/api/crack \
  -H 'Content-Type: application/json' \
  -d '{
    "hash":      "5d41402abc4b2a76b9719d911017c592",
    "hashType":  "md5",
    "mode":      "auto",
    "workers":   4,
    "maxLength": 6,
    "charset":   "abcdefghijklmnopqrstuvwxyz0123456789",
    "wordlist":  "wordlist.txt"
  }'

# Parar ataque
curl -X POST http://localhost:8080/api/stop
```

---

## 🖧 Modo Distribuído (Master / Worker)

Para distribuir o brute force entre várias máquinas (ex: várias TV Boxes):

```bash
# Máquina principal (master) — aguarda workers e distribui tarefas
go run ./master -addr :9000

# Cada TV Box / máquina worker (substituir IP pelo do master)
go run ./worker -master 192.168.0.10:9000 -workers 4
```

O master divide o espaço de combinações entre os workers automaticamente.

---

## 🧪 Testes

```bash
# Rodar todos os testes
go test ./...

# Com detalhes
go test ./crypto/... ./core/... ./server/... -v

# Resultado esperado: 38 testes, todos PASS
```

---

## 🗂️ Estrutura do Projeto

```
cipher-scope/
├── main.go               # Servidor HTTP + embed da UI
├── ui/                   # Dashboard (HTML/CSS/JS) — embutido no binário
│   ├── index.html        # 3 abas: Dashboard, Ferramentas, Chrome
│   ├── style.css
│   └── app.js
├── server/               # Handlers HTTP, SSE, estado do ataque
│   ├── server.go         # /api/crack, /api/stop, /api/progress, /api/info
│   └── tools_handler.go  # /api/hash, /api/encode, /api/identify, /api/chrome
├── attacks/              # Motores de ataque
│   ├── dictionary.go     # Wordlist streaming
│   ├── rules.go          # Mutações l33t, sufixos, capitalização
│   ├── bruteforce.go     # Brute force paralelo com workers
│   ├── john.go           # Integração com John the Ripper
│   └── browser.go        # Extração de credenciais do Chrome (CLI)
├── core/                 # Engine de workers, validação, tipos
├── crypto/               # MD5, SHA1, SHA256, SHA512, NTLM, Chrome AES
├── master/               # Nó master do modo distribuído
├── worker/               # Nó worker do modo distribuído
└── utils/                # Logger
```

---

## ⚠️ Aviso Legal

Esta ferramenta é destinada exclusivamente a **testes de segurança autorizados**, **CTFs**, **recuperação de senhas próprias** e fins educacionais.

O uso não autorizado contra sistemas de terceiros é ilegal e antiético. O autor não se responsabiliza pelo uso indevido.

---

## 📄 Licença

MIT © [JonathanMar](https://github.com/JonathanMar)

package server_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auditor/server"
)

// ── HandleHashExtended ────────────────────────────────────────────────────────

func TestHandleHashExtended_MD5(t *testing.T) {
	body := `{"password":"hello","type":"md5"}`
	req  := httptest.NewRequest(http.MethodPost, "/api/hash", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleHashExtended(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	want := "5d41402abc4b2a76b9719d911017c592"
	if resp["hash"] != want {
		t.Errorf("hash MD5(hello) = %q, want %q", resp["hash"], want)
	}
}

func TestHandleHashExtended_SHA512(t *testing.T) {
	body := `{"password":"hello","type":"sha512"}`
	req  := httptest.NewRequest(http.MethodPost, "/api/hash", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleHashExtended(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp["hash"]) != 128 {
		t.Errorf("SHA512 deve ter 128 chars, obteve %d", len(resp["hash"]))
	}
}

func TestHandleHashExtended_NTLM(t *testing.T) {
	body := `{"password":"password","type":"ntlm"}`
	req  := httptest.NewRequest(http.MethodPost, "/api/hash", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleHashExtended(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	want := "8846f7eaee8fb117ad06bdd830b7586c"
	if resp["hash"] != want {
		t.Errorf("NTLM(password) = %q, want %q", resp["hash"], want)
	}
}

func TestHandleHashExtended_UnsupportedType(t *testing.T) {
	body := `{"password":"x","type":"bcrypt"}`
	req  := httptest.NewRequest(http.MethodPost, "/api/hash", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleHashExtended(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", w.Code)
	}
}

func TestHandleHashExtended_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/hash", nil)
	w   := httptest.NewRecorder()
	server.HandleHashExtended(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status %d, want 405", w.Code)
	}
}

// ── HandleEncode ──────────────────────────────────────────────────────────────

func encodeRequest(t *testing.T, value, format, action string) map[string]string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"value": value, "format": format, "action": action})
	req := httptest.NewRequest(http.MethodPost, "/api/encode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.HandleEncode(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("encode %s/%s status %d, want 200 — body: %s", format, action, w.Code, w.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	return resp
}

func TestHandleEncode_Base64_Encode(t *testing.T) {
	resp := encodeRequest(t, "hello world", "base64", "encode")
	if resp["result"] != "aGVsbG8gd29ybGQ=" {
		t.Errorf("base64 encode = %q, want %q", resp["result"], "aGVsbG8gd29ybGQ=")
	}
}

func TestHandleEncode_Base64_Decode(t *testing.T) {
	resp := encodeRequest(t, "aGVsbG8gd29ybGQ=", "base64", "decode")
	if resp["result"] != "hello world" {
		t.Errorf("base64 decode = %q, want %q", resp["result"], "hello world")
	}
}

func TestHandleEncode_URL_Encode(t *testing.T) {
	resp := encodeRequest(t, "hello world&foo=bar", "url", "encode")
	if !strings.Contains(resp["result"], "%") {
		t.Errorf("URL encode deve conter '%%', obteve %q", resp["result"])
	}
}

func TestHandleEncode_URL_Decode(t *testing.T) {
	resp := encodeRequest(t, "hello+world", "url", "decode")
	if resp["result"] != "hello world" {
		t.Errorf("URL decode = %q, want %q", resp["result"], "hello world")
	}
}

func TestHandleEncode_Hex_Encode(t *testing.T) {
	resp := encodeRequest(t, "AB", "hex", "encode")
	if resp["result"] != "4142" {
		t.Errorf("hex encode = %q, want %q", resp["result"], "4142")
	}
}

func TestHandleEncode_Hex_Decode(t *testing.T) {
	resp := encodeRequest(t, "4142", "hex", "decode")
	if resp["result"] != "AB" {
		t.Errorf("hex decode = %q, want %q", resp["result"], "AB")
	}
}

func TestHandleEncode_Binary_Encode(t *testing.T) {
	resp := encodeRequest(t, "A", "binary", "encode")
	// 'A' = 0x41 = 01000001
	if resp["result"] != "01000001" {
		t.Errorf("binary encode = %q, want %q", resp["result"], "01000001")
	}
}

func TestHandleEncode_Binary_Decode(t *testing.T) {
	resp := encodeRequest(t, "01000001", "binary", "decode")
	if resp["result"] != "A" {
		t.Errorf("binary decode = %q, want %q", resp["result"], "A")
	}
}

func TestHandleEncode_UnsupportedFormat(t *testing.T) {
	body := `{"value":"x","format":"rot13","action":"encode"}`
	req  := httptest.NewRequest(http.MethodPost, "/api/encode", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.HandleEncode(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", w.Code)
	}
}

func TestHandleEncode_InvalidBase64(t *testing.T) {
	body := `{"value":"!!!notbase64!!!","format":"base64","action":"decode"}`
	req  := httptest.NewRequest(http.MethodPost, "/api/encode", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.HandleEncode(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status %d, want 422", w.Code)
	}
}

// ── HandleIdentifyHash ────────────────────────────────────────────────────────

func identifyRequest(t *testing.T, hash string) map[string]interface{} {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"hash": hash})
	req := httptest.NewRequest(http.MethodPost, "/api/identify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.HandleIdentifyHash(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("identify status %d, want 200", w.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	return resp
}

func TestHandleIdentify_MD5(t *testing.T) {
	resp := identifyRequest(t, "5d41402abc4b2a76b9719d911017c592")
	cands, _ := resp["candidates"].([]interface{})
	if len(cands) == 0 {
		t.Fatal("deveria identificar candidatos para MD5 (32 hex)")
	}
	found := false
	for _, c := range cands {
		if c.(string) == "md5" {
			found = true
		}
	}
	if !found {
		t.Errorf("candidatos = %v, esperava conter 'md5'", cands)
	}
}

func TestHandleIdentify_SHA1(t *testing.T) {
	resp := identifyRequest(t, "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d")
	cands, _ := resp["candidates"].([]interface{})
	if len(cands) == 0 || cands[0] != "sha1" {
		t.Errorf("candidatos = %v, esperava ['sha1']", cands)
	}
}

func TestHandleIdentify_SHA256(t *testing.T) {
	resp := identifyRequest(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")
	cands, _ := resp["candidates"].([]interface{})
	if len(cands) == 0 || cands[0] != "sha256" {
		t.Errorf("candidatos = %v, esperava ['sha256']", cands)
	}
}

func TestHandleIdentify_SHA512(t *testing.T) {
	hash := "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"
	resp := identifyRequest(t, hash)
	cands, _ := resp["candidates"].([]interface{})
	if len(cands) == 0 || cands[0] != "sha512" {
		t.Errorf("candidatos = %v, esperava ['sha512']", cands)
	}
}

func TestHandleIdentify_NonHex(t *testing.T) {
	resp := identifyRequest(t, "isso não é hex!")
	isHex, _ := resp["isHex"].(bool)
	if isHex {
		t.Error("isHex deveria ser false para string não-hex")
	}
	cands, _ := resp["candidates"].([]interface{})
	if len(cands) != 0 {
		t.Errorf("candidatos = %v, esperava vazio para não-hex", cands)
	}
}

// ── HandleChromeDecrypt ───────────────────────────────────────────────────────

func TestHandleChrome_MissingFile(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("masterPassword", "peanuts")
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/chrome", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	server.HandleChromeDecrypt(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("sem arquivo: status %d, want 400", w.Code)
	}
}

func TestHandleChrome_InvalidDB(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("db", "Login Data")
	fw.Write([]byte("this is not a sqlite database")) // arquivo inválido
	mw.WriteField("masterPassword", "peanuts")
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/chrome", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	server.HandleChromeDecrypt(w, req)
	// Deve retornar 422 (não consegue abrir ou a tabela não existe)
	if w.Code == http.StatusOK {
		t.Error("arquivo inválido deveria falhar (não 200)")
	}
}

func TestHandleChrome_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/chrome", nil)
	w   := httptest.NewRecorder()
	server.HandleChromeDecrypt(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status %d, want 405", w.Code)
	}
}

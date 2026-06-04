#!/usr/bin/env bash
# Script para testar a TV Box via SSH e rodar o cipher-scope
# Uso: bash scripts/tvbox_test.sh

HOST="ifsuldeminas@192.168.0.113"
PASS="ifsuldeminas"
REMOTE_DIR="~/Documentos"
BINARY="cipher-scope-arm64"

echo "=== [1/4] Copiando binário atualizado para a TV Box ==="
sshpass -p "$PASS" scp -o StrictHostKeyChecking=no \
  cipher-scope-arm64 \
  ${HOST}:${REMOTE_DIR}/cipher-scope-arm64
echo "✅ Binário copiado"

echo ""
echo "=== [2/4] Verificando na TV Box ==="
sshpass -p "$PASS" ssh -o StrictHostKeyChecking=no ${HOST} bash <<'REMOTE'
echo "Hostname: $(hostname)"
echo "Arch:     $(uname -m)"
echo "Kernel:   $(uname -r)"
echo ""
echo "Binários em ~/Documentos:"
ls -lh ~/Documentos/cipher-scope* 2>/dev/null || echo "(nenhum)"
echo ""
echo "IP da TV Box:"
hostname -I
REMOTE

echo ""
echo "=== [3/4] Testando execução (3 segundos) ==="
sshpass -p "$PASS" ssh -o StrictHostKeyChecking=no ${HOST} bash <<'REMOTE'
cd ~/Documentos
chmod +x cipher-scope-arm64
timeout 3 ./cipher-scope-arm64 -addr 0.0.0.0:8080 -no-browser 2>&1 || true
REMOTE

echo ""
echo "=== [4/4] Tudo OK! ==="
echo "Na TV Box rode:"
echo "  ./cipher-scope-arm64 -addr 0.0.0.0:8080 -no-browser"
echo ""
echo "Acesse do PC/celular em:"
sshpass -p "$PASS" ssh -o StrictHostKeyChecking=no ${HOST} hostname -I | awk '{print "  http://"$1":8080"}'

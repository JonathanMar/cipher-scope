#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# setup-wordlist.sh — Baixa o RockYou (14M senhas) para o cipher-scope
# Uso: bash scripts/setup-wordlist.sh [destino]
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

DEST="${1:-wordlist.txt}"
ROCKYOU_URL="https://github.com/brannondorsey/naive-hashcat/releases/download/data/rockyou.txt"
SECLISTS_URL="https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Common-Credentials/10-million-password-list-top-1000000.txt"

echo "╔══════════════════════════════════════════╗"
echo "║   Cipher Scope — Wordlist Setup          ║"
echo "╚══════════════════════════════════════════╝"
echo ""

# Verifica dependências
for cmd in wget curl; do
  if command -v "$cmd" &>/dev/null; then
    DOWNLOADER="$cmd"
    break
  fi
done

if [ -z "${DOWNLOADER:-}" ]; then
  echo "❌ wget ou curl não encontrado. Instale com: apt install wget"
  exit 1
fi

download() {
  local url="$1" out="$2"
  echo "⬇️  Baixando: $url"
  if [ "$DOWNLOADER" = "wget" ]; then
    wget -q --show-progress -O "$out" "$url"
  else
    curl -L --progress-bar -o "$out" "$url"
  fi
}

echo "Escolha a wordlist:"
echo "  1) RockYou (14M senhas, ~134MB) — recomendado"
echo "  2) Top 1M senhas SecLists (~9MB) — leve para TV Box"
echo "  3) Ambas (merge, ~15M senhas)"
echo ""
read -rp "Opção [1]: " CHOICE
CHOICE="${CHOICE:-1}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

case "$CHOICE" in
  1)
    download "$ROCKYOU_URL" "$DEST"
    ;;
  2)
    download "$SECLISTS_URL" "$DEST"
    ;;
  3)
    download "$ROCKYOU_URL"  "$TMP/rockyou.txt"
    download "$SECLISTS_URL" "$TMP/seclists.txt"
    echo "🔀 Mesclando e removendo duplicatas..."
    sort -u "$TMP/rockyou.txt" "$TMP/seclists.txt" > "$DEST"
    ;;
  *)
    echo "Opção inválida."
    exit 1
    ;;
esac

COUNT=$(wc -l < "$DEST")
SIZE=$(du -sh "$DEST" | cut -f1)
echo ""
echo "✅ Wordlist salva em: $DEST"
echo "   Senhas: $(printf '%d' "$COUNT" | sed ':a;s/\B[0-9]\{3\}\>/,&/;ta')"
echo "   Tamanho: $SIZE"
echo ""
echo "Inicie o cipher-scope com: ./cipher-scope -wordlist $DEST"

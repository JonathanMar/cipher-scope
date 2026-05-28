#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# setup-john-arm64.sh — Instala John the Ripper (Jumbo) em Linux ARM64/ARM32
# Testado em: Amlogic S905X3, Rockchip RK3318 (armbian, ubuntu, debian)
# Uso: bash scripts/setup-john-arm64.sh
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

JOHN_REPO="https://github.com/openwall/john.git"
INSTALL_DIR="$HOME/john"
BIN_LINK="/usr/local/bin/john"

echo "╔══════════════════════════════════════════╗"
echo "║  Cipher Scope — John the Ripper Setup   ║"
echo "║  Plataforma: $(uname -m)                     ║"
echo "╚══════════════════════════════════════════╝"
echo ""

# ── Verificar se já está instalado ───────────────────────────────────────────
if command -v john &>/dev/null; then
  echo "✅ John já instalado: $(john --version 2>&1 | head -1)"
  exit 0
fi

# ── Detectar distro/gerenciador de pacotes ────────────────────────────────────
if command -v apt-get &>/dev/null; then
  PKG_INSTALL="sudo apt-get install -y"
  PKG_UPDATE="sudo apt-get update -qq"
elif command -v pacman &>/dev/null; then
  PKG_INSTALL="sudo pacman -S --noconfirm"
  PKG_UPDATE="sudo pacman -Sy"
else
  echo "❌ Gerenciador de pacotes não reconhecido (precisa de apt ou pacman)"
  exit 1
fi

echo "📦 Instalando dependências de compilação..."
$PKG_UPDATE
$PKG_INSTALL build-essential libssl-dev zlib1g-dev git libgmp-dev

echo ""
echo "📥 Clonando John the Ripper (Jumbo)..."
if [ -d "$INSTALL_DIR" ]; then
  echo "   Diretório $INSTALL_DIR já existe, atualizando..."
  git -C "$INSTALL_DIR" pull --ff-only
else
  git clone --depth=1 "$JOHN_REPO" "$INSTALL_DIR"
fi

echo ""
echo "⚙️  Compilando (pode levar 5-15 min na TV Box)..."
cd "$INSTALL_DIR/src"
./configure --disable-openmp 2>&1 | tail -5
make -j"$(nproc)" 2>&1 | tail -10

echo ""
echo "🔗 Criando link em /usr/local/bin/john..."
sudo ln -sf "$INSTALL_DIR/run/john" "$BIN_LINK"

echo ""
echo "✅ John instalado com sucesso!"
john --version | head -1
echo ""
echo "Reinicie o cipher-scope para que ele detecte o John automaticamente."

#!/usr/bin/env bash
# Script para publicar todos os commits pendentes + release v1.0.1
# Uso: GH_TOKEN=ghp_xxx bash scripts/push_all.sh

set -e
DIR="$(cd "$(dirname "$0")/.." && pwd)"
TOKEN="${GH_TOKEN}"
REPO="JonathanMar/cipher-scope"

if [ -z "$TOKEN" ]; then
  echo "Erro: defina GH_TOKEN antes de rodar este script"
  exit 1
fi

cd "$DIR"

echo "=== [1/3] Configurando remote com token ==="
git remote set-url origin "https://JonathanMar:${TOKEN}@github.com/${REPO}.git"

echo "=== [2/3] Commitando e fazendo push ==="
git add README.md main.go 2>/dev/null || true
git diff --cached --quiet || git commit -m "fix+docs: corrigir URL banner + guia completo TV Box no README"

# Push via cherry-pick em branch limpo para evitar o commit do workflow
REMOTE_SHA=$(git rev-parse origin/main)
LOCAL_SHA=$(git rev-parse HEAD)

if [ "$REMOTE_SHA" = "$LOCAL_SHA" ]; then
  echo "Já está sincronizado com o remoto."
else
  # Criar branch temporário baseado no remote
  git checkout -b _push_tmp origin/main 2>/dev/null

  # Cherry-pick todos os commits novos (exceto o do workflow)
  git log --oneline "${REMOTE_SHA}..${LOCAL_SHA}" --reverse | while read hash msg; do
    echo "Cherry-picking: $hash $msg"
    git cherry-pick "$hash" 2>/dev/null || true
  done

  NEW_SHA=$(git rev-parse HEAD)
  git push origin "${NEW_SHA}:refs/heads/main"
  git checkout main
  git branch -D _push_tmp 2>/dev/null || true
fi

echo ""
echo "=== [3/3] Publicando release v1.0.1 ==="
TAG="v1.0.1"
COMMIT=$(git rev-parse origin/main)

# Criar tag
echo "Criando tag $TAG..."
TAG_RES=$(curl -s -X POST \
  -H "Authorization: token $TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/$REPO/git/refs" \
  -d "{\"ref\":\"refs/tags/$TAG\",\"sha\":\"$COMMIT\"}")
echo "$TAG_RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print('Tag:', d.get('ref', d.get('message','?')))"

# Criar release
echo "Criando release..."
REL=$(curl -s -X POST \
  -H "Authorization: token $TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/$REPO/releases" \
  -d "{\"tag_name\":\"$TAG\",\"name\":\"Cipher Scope v1.0.1\",\"body\":\"### Correções\\n- Fix URL no banner quando -addr tem IP explícito\\n- Guia completo TV Box no README\\n\\n### Downloads\\n- \`cipher-scope-amd64\` — Linux x86_64\\n- \`cipher-scope-arm64\` — TV Box ARM64 (Tanix TX6, X96 Max...)\\n- \`cipher-scope-arm32\` — TV Box ARM32 (32-bit)\",\"draft\":false,\"prerelease\":false}")

REL_ID=$(echo "$REL" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id','ERRO:'+str(d.get('message',''))))")
UPLOAD_URL="https://uploads.github.com/repos/$REPO/releases/$REL_ID/assets"
echo "Release ID: $REL_ID"

upload() {
  echo "Uploading $2..."
  RES=$(curl -s -X POST \
    -H "Authorization: token $TOKEN" \
    -H "Content-Type: application/octet-stream" \
    "$UPLOAD_URL?name=$2" \
    --data-binary @"$DIR/$1")
  echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); url=d.get('browser_download_url'); print('✅ '+url) if url else print('❌', d.get('message','?'))"
}

upload cipher-scope-amd64 cipher-scope-amd64
upload cipher-scope-arm64 cipher-scope-arm64
upload cipher-scope-arm32 cipher-scope-arm32

# Limpar token do remote
git remote set-url origin "git@github.com:${REPO}.git"

echo ""
echo "✅ Tudo publicado!"
echo "   https://github.com/$REPO/releases/tag/$TAG"

#!/usr/bin/env sh
# Writes helper scripts for a throwaway test project under test-proj/.
set -eu

mkdir -p test-proj
cd test-proj || exit 1

cat > run.sh <<'EOF'
clear
sh scripts/build.sh
cp "build/gitagger-windows-amd64.exe" "test-proj/gitagger.exe"
cd test-proj
touch test.txt
git init
git add .
git commit -m "init"
./gitagger init --force
cat .gitagger.yml
EOF

cat > delete.sh <<'EOF'
clear
cd test-proj
rm -rf -- * .[!.]* ..?*
EOF

cat > test.sh <<'EOF'
clear
cd test-proj
sh run.sh >> log.txt
EOF

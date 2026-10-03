mkdir -p test-proj
cd test-proj

echo "clear
"scripts/build-windows.bat"
cp "build/gitagger-windows-amd64.exe" "test-proj/gitagger.exe"
cd test-proj
touch test.txt
git init
git add .
git commit -m "init"
./gitagger init --force
./gitagger workflow init gh --force
./gitagger workflow init jenkins --force
cat .gitagger.yml
" > run.sh

echo "clear
cd test-proj
rm -rf -- * .[!.]* ..?*" > delete.sh

echo "clear
cd test-proj
sh run.sh >> log.txt
" > test.sh
.PHONY: clean, test

test: cgp
	make compile && lit -v tests

cgp: 
	go build -o cgp.exe

compile:
	C:/Program\ Files/Git/usr/bin/find.exe tests -type f -exec ./cgp.exe -c {} \; &>/dev/null

clean:
	C:/Program\ Files/Git/usr/bin/find.exe tests -type f -name '*.ll' -delete

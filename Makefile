.PHONY: test, compile, clean

test: cgp
	make compile && lit -v tests

cgp: 
	go build -o cgp.exe

compile:
	find tests -type f -exec ./cgp.exe -c {} \; &>/dev/null

clean:
	find tests -type f -name "*.ll" -delete
	rm -f cgp.exe

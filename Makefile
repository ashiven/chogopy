.PHONY: test, compile, clean

test: cgp compile
	lit -v tests

cgp: 
	go build -o cgp.exe

compile:
	find tests -type f -name "*.choc" -exec ./cgp.exe -c {} \; >/dev/null 2>&1

clean:
	find tests -type f -name "*.ll" -delete
	rm -f cgp.exe

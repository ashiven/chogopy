.PHONY: test, compile, clean

test: clean compile
	lit -v tests

cgp: 
	go build -tags=llvm18 -o cgp

compile: cgp
	find tests -type f -name "*.choc" -exec ./cgp {} \; >/dev/null 2>&1

clean:
	find tests -type f ! \( -name "*.choc" -o -name "lit.cfg" \) -delete
	rm -f cgp

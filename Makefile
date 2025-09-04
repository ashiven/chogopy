.PHONY: test, compile, clean

test: clean cgp compile
	lit -v tests

cgp: 
	go build -tags=llvm18 -o cgp

compile:
	find tests -type f -name "*.choc" -exec ./cgp -c {} \; >/dev/null 2>&1

clean:
	find tests -type f -name "*.ll" -delete
	rm -f cgp

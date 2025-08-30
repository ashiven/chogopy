.PHONY: test, compile, clean

test: clean cgp compile
	lit -v tests

cgp: 
	go build -o cgp

compile:
	find tests -type f -name "*.choc" -exec ./cgp -c {} \; >/dev/null 2>&1

clean:
	find tests -type f -name "*.ll" -delete
	rm -f cgp

install:
	python3 -m pip install -r requirements.txt
.PHONY: install

test:
	python3 -m unittest test_ping_pong.py
.PHONY: test

lint:
	flake8 ping_pong.py test_ping_pong.py
	black --check ping_pong.py test_ping_pong.py
.PHONY: lint

format:
	black ping_pong.py test_ping_pong.py
.PHONY: format

clean:
	rm -rf __pycache__
.PHONY: clean

run:
	python3 ping_pong.py
.PHONY: run

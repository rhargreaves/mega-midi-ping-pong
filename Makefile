install:
	python3 -m pip install -r requirements.txt
.PHONY: install

test:
	PYTHONPATH=. python3 test/test_ping_pong.py
.PHONY: test

lint:
	flake8 ping_pong/*.py test/*.py
	black --check ping_pong/*.py test/*.py
.PHONY: lint

format:
	black ping_pong/*.py test/*.py
.PHONY: format

clean:
	rm -rf __pycache__
.PHONY: clean

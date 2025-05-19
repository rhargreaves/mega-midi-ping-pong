.PHONY: install test lint clean verify

install:
	python3 -m pip install -r requirements.txt

test:
	python3 -m unittest test_ping_pong.py

lint:
	flake8 ping_pong.py test_ping_pong.py
	black --check ping_pong.py test_ping_pong.py

format:
	black ping_pong.py test_ping_pong.py

clean:
	rm -rf __pycache__

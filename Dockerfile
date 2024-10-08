# syntax=docker/dockerfile:1@sha256:865e5dd094beca432e8c0a1d5e1c465db5f998dca4e439981029b3b81fb39ed5

### convert poetry.lock to requirements.txt ###
FROM python:3.10-slim@sha256:1eb5d76bf3e9e612176ebf5eadf8f27ec300b7b4b9a99f5856f8232fd33aa16e AS poetry

WORKDIR /app

ENV PIP_ROOT_USER_ACTION=ignore

COPY requirements-poetry.txt ./
RUN pip install -r requirements-poetry.txt

COPY pyproject.toml poetry.lock ./
RUN poetry export -f requirements.txt --output requirements.txt

### final image ###
FROM python:3.10-slim@sha256:1eb5d76bf3e9e612176ebf5eadf8f27ec300b7b4b9a99f5856f8232fd33aa16e

WORKDIR /app

ENV PYTHONPATH=/app

COPY --from=poetry /app/requirements.txt ./requirements.txt

ENV PIP_ROOT_USER_ACTION=ignore

RUN pip install -U pip && \
    pip install -r requirements.txt

ENTRYPOINT [ "python", "./main.py" ]

COPY . ./

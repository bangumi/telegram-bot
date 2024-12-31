# syntax=docker/dockerfile:1@sha256:93bfd3b68c109427185cd78b4779fc82b484b0b7618e36d0f104d4d801e66d25

### convert poetry.lock to requirements.txt ###
FROM python:3.10-slim@sha256:af6f1b19eae3400ea3a569ba92d4819a527be4662971d51bb798c923bba30a81 AS poetry

WORKDIR /app

ENV PIP_ROOT_USER_ACTION=ignore

COPY requirements-poetry.txt ./
RUN pip install -r requirements-poetry.txt

COPY pyproject.toml poetry.lock ./
RUN poetry export -f requirements.txt --output requirements.txt

### final image ###
FROM python:3.10-slim@sha256:af6f1b19eae3400ea3a569ba92d4819a527be4662971d51bb798c923bba30a81

WORKDIR /app

ENV PYTHONPATH=/app

COPY --from=poetry /app/requirements.txt ./requirements.txt

ENV PIP_ROOT_USER_ACTION=ignore

RUN pip install -U pip && \
    pip install -r requirements.txt

ENTRYPOINT [ "python", "./main.py" ]

COPY . ./

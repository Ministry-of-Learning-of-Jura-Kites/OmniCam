FROM python:3.12-slim-trixie

COPY --from=ghcr.io/astral-sh/uv:0.10 /uv /uvx /bin/

RUN apt update \
  && apt install -y --no-install-recommends \
  protobuf-compiler \
  python3-vtk9 \
  && rm -rf /var/lib/apt/lists/*

# Install deps for python
WORKDIR /app/algo

COPY predictive-algorithm/pyproject.toml predictive-algorithm/uv.lock ./

RUN uv sync --locked --no-dev

WORKDIR /app

COPY protobufs/ ./protobufs/

RUN mkdir -p /app/algo/src/messages/protobufs && cd protobufs && \
  protoc --python_out=/app/algo/src/messages/protobufs \
  --pyi_out=/app/algo/src/messages/protobufs \
  *.proto

WORKDIR /app/algo

COPY . .

ENV PATH="/app/.venv/bin:$PATH"

CMD ["python","src/main.py"]

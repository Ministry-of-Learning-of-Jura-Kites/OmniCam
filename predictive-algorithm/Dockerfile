FROM python:3.12-slim-trixie

COPY --from=ghcr.io/astral-sh/uv:0.10 /uv /uvx /bin/

RUN apt-get update && apt-get install -y --no-install-recommends python3-vtk9 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY pyproject.toml uv.lock ./

RUN uv sync --locked --no-dev

COPY . .

WORKDIR /app

ENV PATH="/app/.venv/bin:$PATH"

# CMD ["tail", "-f", "/dev/null"]

CMD ["uvicorn", "main:app", "--app-dir", "src", "--host", "0.0.0.0","--port", "8081"]

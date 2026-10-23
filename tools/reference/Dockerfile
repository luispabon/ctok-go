FROM python:3.13.17-slim@sha256:d879ed88565a7a3c614b980b369d4a5025a9d057b71f93d09fc6d0b2ca823a18

COPY tag.tar.gz /tmp/ctok.tar.gz
COPY regex-2026.8.31-cp313-cp313-manylinux2014_x86_64.manylinux_2_17_x86_64.manylinux_2_28_x86_64.whl /tmp/
RUN echo '80c051e1ffeebdb8c58d024366bde69f84f57b7d230ab976475b6d6c67252646  /tmp/ctok.tar.gz' | sha256sum -c - \
    && echo '15e9e862c6e905ef66ea5f019deb5ac5fdeebf8fc134ea4c7b5d5c2eb7bdcdd8  /tmp/regex-2026.8.31-cp313-cp313-manylinux2014_x86_64.manylinux_2_17_x86_64.manylinux_2_28_x86_64.whl' | sha256sum -c - \
    && mkdir /opt/reference \
    && tar -xzf /tmp/ctok.tar.gz -C /opt/reference --strip-components=1 \
    && pip install --no-index --no-deps /tmp/*.whl \
    && rm /tmp/ctok.tar.gz /tmp/*.whl
ENV PYTHONPATH=/opt/reference
WORKDIR /work

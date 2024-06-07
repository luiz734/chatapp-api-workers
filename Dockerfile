# FROM python:3.12.3-alpine
FROM python:3.12.3-slim
# RUN apk add --no-cache git
# RUN apt install git-all
RUN apt update
RUN apt install -y git libgl1 libglib2.0-0 libsm6 libxrender1 libxext6
RUN git clone https://github.com/luiz734/chatapp-api-workers

WORKDIR chatapp-api-workers
RUN python -m ensurepip --upgrade
RUN python -m pip install --upgrade setuptools
# RUN pip install --upgrade virtualenv
# RUN python -m venv venv && chmod +x venv/bin/activate && source venv/bin/activate
RUN python -m pip install -r requirements.txt
CMD python main.py

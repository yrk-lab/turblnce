FROM python:alpine
ADD ./scheduler.py /usr/local/bin/kube-scheduler

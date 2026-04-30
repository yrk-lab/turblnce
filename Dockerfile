FROM python:alpine
ADD ./scheduler.py /usr/local/bin/kube-scheduler
RUN pip install kubernetes && chmod +x /usr/local/bin/kube-scheduler

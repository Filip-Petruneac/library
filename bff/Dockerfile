FROM python:3

WORKDIR /bff/

COPY . .

RUN pip install --no-cache-dir -r requirements.txt

EXPOSE 8081

CMD ["python", "main.py"]


FROM golang

# Выбираем рабочую папку


ADD . /go/src/Concurs
WORKDIR /go/src/Concurs

RUN mkdir /tmp/data
#COPY ./data/data.zip ./data/options.txt /tmp/data/


# Копируем наш исходный main.go внутрь контейнера, в папку go/src/dumb
RUN go get github.com/valyala/fasthttp
RUN go get github.com/buaazp/fasthttprouter
RUN go get github.com/buger/jsonparser
# Компилируем и устанавливаем наш сервер
RUN  go install Concurs
# Открываем 80-й порт наружу
EXPOSE 80
# Запускаем наш сервер
CMD /go/bin/Concurs
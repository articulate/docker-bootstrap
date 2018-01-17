FROM ruby:2.4

RUN mkdir -p /service
ADD Gemfile* /service
WORKDIR /service
RUN bundle install

ADD spec /service/spec
ADD Makefile /service/Makefile

CMD rspec

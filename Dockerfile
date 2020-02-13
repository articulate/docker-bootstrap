FROM ruby:2.5

RUN mkdir -p /service
ADD Gemfile* /service/
WORKDIR /service
RUN bundle install

ADD spec /service/spec

CMD rspec

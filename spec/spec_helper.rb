require "dockerspec/serverspec"
require 'diplomat'

def get_config(key)
  Specinfra.configuration.send(key)
end

def docker_build_template(opts = {})
  git_commit = ENV["TRAVIS_COMMIT"] || `git rev-parse HEAD`.strip
  opts.merge!(context: { git_commit: git_commit})
  docker_build(opts)
end

def add_env_vars(env_vars_hash={})
  set :env, get_config(:env) + [env_vars_hash.to_a]
end

def entrypoint_command(cmd)
  command("/entrypoint.sh #{cmd}")
end

def set_consul(scope, key, value, new=true)
  service_name = get_config(:service_name)
  service_product = get_config(:service_product)
  service_env = get_config(:service_env)

  consul_key = case scope
               when :service
                 if new
                   "services/#{service_name}/env_vars/#{key}"
                 else
                   "apps/#{service_name}/#{service_env}/env_vars/#{key}"
                 end
               when :global
                 if new
                   "global/env_vars/#{key}"
                 else
                   "global/#{service_env}/env_vars/#{key}"
                 end
               when :product
                 "products/#{service_product}/env_vars/#{key}"
               end

  puts "Setting '#{consul_key}' to value '#{value}'"

  before(:each) do
    Diplomat::Kv.put(consul_key, value)
  end

  after(:each) do
    Diplomat::Kv.delete(consul_key)
  end
end

Diplomat.configure do |diplomat_config|
  diplomat_config.url = "http://consul:8500"
end

RSpec.configure do |config|
  set :docker_container_create_options, HostConfig: { NetworkMode: "container:#{`hostname`.strip}" }

  set :service_name, "service-name"
  set :service_env, "stage"
  set :service_product, "product-name"

  set :env, [
    ["CONSUL_ADDR", "consul:8500"],
    ["APP_ENV", get_config(:service_env)],
    ["APP_NAME", get_config(:service_name)],
    ["APP_PRODUCT", get_config(:service_product)]
  ]

  config.before(:suite) do
    Diplomat::Kv.delete("", recurse: true)
  end
end

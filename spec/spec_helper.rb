require "dockerspec/serverspec"
require 'diplomat'
require 'vault'

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

def set_var(store_type, scope, key, value, new=true)
  service_name = get_config(:service_name)
  service_product = get_config(:service_product)
  service_env = get_config(:service_env)

  full_key = case scope
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
  if store_type == :vault
    full_key = "secret/#{full_key}"
  end

  puts "Setting '#{full_key}' to value '#{value}' in #{store_type}"

  before(:each) do
    if store_type == :vault
      Vault.logical.write(full_key, value: value)
    else
      Diplomat::Kv.put(full_key, value)
    end
  end

  after(:each) do
    if store_type == :vault
      Vault.logical.delete(full_key)
    else
      Diplomat::Kv.delete(full_key)
    end
  end
end

Diplomat.configure do |diplomat_config|
  diplomat_config.url = "http://#{ENV.fetch("CONSUL_ADDR")}"
end

RSpec::Matchers.define :include_env do |env_var, env_val|
  match do |stdout|
    stdout.include?("#{env_var}=#{env_val}\n")
  end
  failure_message_when_negated do |actual|
    "expected that #{env_var}=#{env_val} is in #{stdout}"
  end
end

RSpec.configure do |config|
  set :docker_container_create_options, HostConfig: { NetworkMode: "container:#{`hostname`.strip}" }

  set :service_name, "service-name"
  set :service_env, "stage"
  set :service_product, "product-name"

  set :env, [
    ["CONSUL_ADDR", ENV.fetch("CONSUL_ADDR")],
    ["VAULT_ADDR", ENV.fetch("VAULT_ADDR")],
    ["VAULT_TOKEN", ENV.fetch("VAULT_TOKEN")],
    ["VAULT_RENEW_TOKEN", "false"],
    ["SERVICE_ENV", get_config(:service_env)],
    ["SERVICE_NAME", get_config(:service_name)],
    ["SERVICE_PRODUCT", get_config(:service_product)]
  ]

  config.before(:suite) do
    Vault.logical.delete("secret")
    Diplomat::Kv.delete("", recurse: true)
  end
end

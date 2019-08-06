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

def docker_run_with_envs(name, env_vars_hash={})
  global_env_vars = {
    "CONSUL_ADDR": ENV.fetch("CONSUL_ADDR"),
    "VAULT_ADDR": ENV.fetch("VAULT_ADDR"),
    "VAULT_TOKEN": ENV.fetch("VAULT_TOKEN"),
    "VAULT_RENEW_TOKEN": "false",
    "SERVICE_ENV": $service_env,
    "SERVICE_NAME": $service_name,
    "SERVICE_PRODUCT": $service_product
  }
  use_env_vars = global_env_vars.merge(env_vars_hash)
  use_env_vars = use_env_vars.to_a

  docker_run(name, env: use_env_vars)
end

def entrypoint_command(cmd)
  command("/entrypoint.sh #{cmd}")
end

def set_var(store_type, scope, key, value, opts={})
  around(:each) do |example|
    service_name = $service_name
    service_product = $service_product
    service_env = opts[:service_env] || $service_env
    peer_id = opts[:peer_id]

    full_key = case scope
               when :service
                 if opts[:old_keys]
                   "apps/#{service_name}/#{service_env}/env_vars/#{key}"
                 else
                   "services/#{service_name}/env_vars/#{key}"
                 end
               when :global
                 if opts[:old_keys]
                   "global/#{service_env}/env_vars/#{key}"
                 else
                   "global/env_vars/#{key}"
                 end
               when :product
                 "products/#{service_product}/env_vars/#{key}"
               when :peer
                "services/#{service_name}/peer/#{peer_id}/env_vars/#{key}"
               end

    if store_type == :vault
      full_key = "secret/#{full_key}"
    end

    #puts "Setting '#{full_key}' to value '#{value}' in #{store_type}"
    if store_type == :vault
      Vault.logical.write(full_key, value: value)
    else
      Diplomat::Kv.put(full_key, value)
    end

    example.run

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
  failure_message_when_negated do |stdout|
    "expected that #{env_var}=#{env_val} is not in #{stdout}"
  end
  failure_message do |stdout|
    "expected that #{env_var}=#{env_val} is in #{stdout}"
  end
end

RSpec.configure do |config|
  config.before(:suite) do
    Vault.logical.delete("secret")
    Diplomat::Kv.delete("", recurse: true)
  end

  $service_name = "service-name"
  $service_env = "stage"
  $service_product = "product-name"

  set :docker_container_create_options, "HostConfig": { "NetworkMode": "container:#{`hostname`.strip}" }
end

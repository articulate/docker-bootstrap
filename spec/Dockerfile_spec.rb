require "dockerspec/serverspec"
require 'diplomat'

container_name = `hostname`.strip

Diplomat.configure do |config|
  config.url = "http://consul:8500"
end

base_env_vars = [
  ["CONSUL_ADDR", "consul:8500"],
  ["APP_ENV", "stage"],
  ["APP_NAME", "project-name"],
  ["APP_SERVICE", "service-name"]
]

def entrypoint_command(cmd)
  command("/entrypoint.sh #{cmd}")
end

describe "Dockerfile" do
  set :docker_container_create_options, HostConfig: { NetworkMode: "container:#{container_name}" }

  describe docker_build(path: "spec/dockerfiles/Dockerfile.debian", tag: "consul_template_bootstrap_debian") do
    describe "Service path env vars" do
      set :env, base_env_vars
      before(:each) do
        Diplomat::Kv.put("services/project-name/env_vars/SOME_SERVICE_VAR", "service-var")
      end

      after(:each) do
        Diplomat::Kv.delete("services/project-name/env_vars/SOME_SERVICE_VAR")
      end

      describe docker_run("consul_template_bootstrap_debian") do
        describe entrypoint_command("env") do
          its(:stdout) { should include("SOME_SERVICE_VAR") }
          its(:stderr) { should be_empty }
        end
      end
    end
  end
end

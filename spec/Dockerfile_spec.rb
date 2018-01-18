require "spec_helper"

describe "Dockerfile" do
  describe docker_build_template(template: "spec/dockerfiles/Dockerfile.debian.erb", tag: "consul_template_bootstrap_debian") do
    describe "Service path env vars" do
      set_consul(:service, "SOME_SERVICE_VAR", "service-var")

      describe docker_run("consul_template_bootstrap_debian") do
        describe entrypoint_command("env") do
          its(:stdout) { should include("SOME_SERVICE_VAR") }
          its(:stderr) { should be_empty }
        end
      end
    end
  end
end

# services override apps
# new global overrides old global
# apps overrides any global
# SERVICE_NAME and APP_NAME both work.

# global sets work
# products sets work
# services sets work
#
# products overrides global
# services overrides products
# services overrides global
# services overrides products and global
#
# dev consul doesn't override set envs
# peer uses stage consul

require "spec_helper"

distros = ["debian", "centos", "alpine"]

distros.each do |distro|
  describe "Dockerfile" do
    describe docker_build_template(template: "spec/dockerfiles/Dockerfile.#{distro}.erb", tag: "consul_template_bootstrap_#{distro}") do
      describe "General sets work" do
        set_consul(:global, "GLOBAL_VAR", "global-var")
        set_consul(:product, "PRODUCT_VAR", "product-var")
        set_consul(:service, "SERVICE_VAR", "service-var")

        describe docker_run("consul_template_bootstrap_#{distro}") do
          describe entrypoint_command("env") do
            its(:stdout) { should include_env "GLOBAL_VAR", "global-var" }
            its(:stdout) { should include_env "PRODUCT_VAR", "product-var" }
            its(:stdout) { should include_env "SERVICE_VAR", "service-var" }
            its(:stderr) { should be_empty }
          end
        end
      end

      describe "Products override global" do
        set_consul(:global, "TEST_VAR", "global-var")
        set_consul(:product, "TEST_VAR", "product-var")

        describe docker_run("consul_template_bootstrap_#{distro}") do
          describe entrypoint_command("env") do
            its(:stdout) { should include_env "TEST_VAR", "product-var" }
            its(:stderr) { should be_empty }
          end
        end
      end

      describe "Services override products" do
        set_consul(:product, "TEST_VAR", "product-var")
        set_consul(:service, "TEST_VAR", "service-var")

        describe docker_run("consul_template_bootstrap_#{distro}") do
          describe entrypoint_command("env") do
            its(:stdout) { should include_env "TEST_VAR", "service-var" }
            its(:stderr) { should be_empty }
          end
        end
      end

      describe "Services override products" do
        set_consul(:global, "TEST_VAR", "global-var")
        set_consul(:service, "TEST_VAR", "service-var")

        describe docker_run("consul_template_bootstrap_#{distro}") do
          describe entrypoint_command("env") do
            its(:stdout) { should include_env "TEST_VAR", "service-var" }
            its(:stderr) { should be_empty }
          end
        end
      end

      describe "Services override globals" do
        set_consul(:global, "TEST_VAR", "global-var")
        set_consul(:service, "TEST_VAR", "service-var")

        describe docker_run("consul_template_bootstrap_#{distro}") do
          describe entrypoint_command("env") do
            its(:stdout) { should include_env "TEST_VAR", "service-var" }
            its(:stderr) { should be_empty }
          end
        end
      end

      describe "Services override products and globals" do
        set_consul(:global, "TEST_VAR", "global-var")
        set_consul(:product, "TEST_VAR", "product-var")
        set_consul(:service, "TEST_VAR", "service-var")

        describe docker_run("consul_template_bootstrap_#{distro}") do
          describe entrypoint_command("env") do
            its(:stdout) { should include_env "TEST_VAR", "service-var" }
            its(:stderr) { should be_empty }
          end
        end
      end
    end
  end
end

# services override apps
# new global overrides old global
# apps overrides any global
# SERVICE_NAME and APP_NAME both work.

# √ global sets work
# √ products sets work
# √ services sets work
#
# products overrides global
# services overrides products
# services overrides global
# services overrides products and global
#
# dev consul doesn't override set envs
# peer uses stage consul
# vault overrides consul

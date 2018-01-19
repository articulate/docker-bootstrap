require "spec_helper"

distros = ["debian", "centos", "alpine"]

distros.each do |distro|
  describe "Dockerfile" do
    describe docker_build_template(template: "spec/dockerfiles/Dockerfile.#{distro}.erb", tag: "consul_template_bootstrap_#{distro}") do
      [:consul, :vault].each do |backend_type|
        describe backend_type do
          describe "General sets work" do
            set_var(backend_type, :global, "GLOBAL_VAR", "global-var")
            set_var(backend_type, :product, "PRODUCT_VAR", "product-var")
            set_var(backend_type, :service, "SERVICE_VAR", "service-var")

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
            set_var(backend_type, :global, "TEST_VAR", "global-var")
            set_var(backend_type, :product, "TEST_VAR", "product-var")

            describe docker_run("consul_template_bootstrap_#{distro}") do
              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "product-var" }
                its(:stderr) { should be_empty }
              end
            end
          end

          describe "Services override products" do
            set_var(backend_type, :product, "TEST_VAR", "product-var")
            set_var(backend_type, :service, "TEST_VAR", "service-var")

            describe docker_run("consul_template_bootstrap_#{distro}") do
              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end
          end

          describe "Services override products" do
            set_var(backend_type, :global, "TEST_VAR", "global-var")
            set_var(backend_type, :service, "TEST_VAR", "service-var")

            describe docker_run("consul_template_bootstrap_#{distro}") do
              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end
          end

          describe "Services override globals" do
            set_var(backend_type, :global, "TEST_VAR", "global-var")
            set_var(backend_type, :service, "TEST_VAR", "service-var")

            describe docker_run("consul_template_bootstrap_#{distro}") do
              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end
          end

          describe "Services override products and globals" do
            set_var(backend_type, :global, "TEST_VAR", "global-var")
            set_var(backend_type, :product, "TEST_VAR", "product-var")
            set_var(backend_type, :service, "TEST_VAR", "service-var")

            describe docker_run("consul_template_bootstrap_#{distro}") do
              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end
          end

          ## LEGACY BACKWARDS COMPATIBLE
          describe "New services space overrides old apps" do
            set_var(backend_type, :service, "TEST_VAR", "service-var")
            set_var(backend_type, :service, "TEST_VAR", "app-var", false)

            describe docker_run("consul_template_bootstrap_#{distro}") do
              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end
          end

          describe "New global space overrides old global" do
            set_var(backend_type, :global, "TEST_VAR", "global-var")
            set_var(backend_type, :global, "TEST_VAR", "old-global-var", false)

            describe docker_run("consul_template_bootstrap_#{distro}") do
              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "global-var" }
                its(:stderr) { should be_empty }
              end
            end
          end

          describe "Old apps override both global spaces" do
            set_var(backend_type, :global, "TEST_VAR", "global-var")
            set_var(backend_type, :global, "TEST_VAR", "old-global-var", false)
            set_var(backend_type, :service, "TEST_VAR", "app-var", false)

            describe docker_run("consul_template_bootstrap_#{distro}") do
              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "app-var" }
                its(:stderr) { should be_empty }
              end
            end
          end

          describe "dev" do
            add_env_vars({
              "SERVICE_ENV": "dev"
            })
            set :service_env, "dev"

            describe "General sets work" do
              set_var(backend_type, :global, "GLOBAL_VAR", "global-var")
              set_var(backend_type, :product, "PRODUCT_VAR", "product-var")
              set_var(backend_type, :service, "SERVICE_VAR", "service-var")

              describe docker_run("consul_template_bootstrap_#{distro}") do
                describe entrypoint_command("env") do
                  its(:stdout) { should include_env "GLOBAL_VAR", "global-var" }
                  its(:stdout) { should include_env "PRODUCT_VAR", "product-var" }
                  its(:stdout) { should include_env "SERVICE_VAR", "service-var" }
                  its(:stderr) { should be_empty }
                end
              end
            end
          end
        end
      end
    end
  end
end

# dev consul doesn't override set envs
# peer uses stage consul
# vault overrides consul

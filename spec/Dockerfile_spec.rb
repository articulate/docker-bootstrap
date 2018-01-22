require "spec_helper"

distros = ["debian", "centos", "alpine"]

distros.each do |distro|
  describe "Dockerfile" do
    describe docker_build_template(template: "spec/dockerfiles/Dockerfile.#{distro}.erb", tag: "consul_template_bootstrap_#{distro}") do
      describe docker_run_with_envs("consul_template_bootstrap_#{distro}") do
        [:consul, :vault].each do |backend_type|
          describe backend_type do
            describe "General sets work" do
              set_var(backend_type, :global, "GLOBAL_VAR", "global-var")
              set_var(backend_type, :product, "PRODUCT_VAR", "product-var")
              set_var(backend_type, :service, "SERVICE_VAR", "service-var")

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "GLOBAL_VAR", "global-var" }
                its(:stdout) { should include_env "PRODUCT_VAR", "product-var" }
                its(:stdout) { should include_env "SERVICE_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end

            describe "Products override global" do
              set_var(backend_type, :global, "TEST_VAR", "global-var")
              set_var(backend_type, :product, "TEST_VAR", "product-var")

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "product-var" }
                its(:stderr) { should be_empty }
              end
            end

            describe "Services override products" do
              set_var(backend_type, :product, "TEST_VAR", "product-var")
              set_var(backend_type, :service, "TEST_VAR", "service-var")

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end

            describe "Services override products" do
              set_var(backend_type, :global, "TEST_VAR", "global-var")
              set_var(backend_type, :service, "TEST_VAR", "service-var")

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end

            describe "Services override globals" do
              set_var(backend_type, :global, "TEST_VAR", "global-var")
              set_var(backend_type, :service, "TEST_VAR", "service-var")

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end

            describe "Services override products and globals" do
              set_var(backend_type, :global, "TEST_VAR", "global-var")
              set_var(backend_type, :product, "TEST_VAR", "product-var")
              set_var(backend_type, :service, "TEST_VAR", "service-var")

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end

            ## LEGACY BACKWARDS COMPATIBLE
            describe "New services space overrides old apps" do
              set_var(backend_type, :service, "TEST_VAR", "service-var")
              set_var(backend_type, :service, "TEST_VAR", "app-var", old_keys: true)

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "service-var" }
                its(:stderr) { should be_empty }
              end
            end

            describe "New global space overrides old global" do
              set_var(backend_type, :global, "TEST_VAR", "global-var")
              set_var(backend_type, :global, "TEST_VAR", "old-global-var", old_keys: true)

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "global-var" }
                its(:stderr) { should be_empty }
              end
            end

            describe "Old apps override both global spaces" do
              set_var(backend_type, :global, "TEST_VAR", "global-var")
              set_var(backend_type, :global, "TEST_VAR", "old-global-var", old_keys: true)
              set_var(backend_type, :service, "TEST_VAR", "app-var", old_keys: true)

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "TEST_VAR", "app-var" }
                its(:stdout) { should_not include_env "TEST_VAR", "old-global-var" }
                its(:stdout) { should_not include_env "TEST_VAR", "global-var" }
                its(:stderr) { should be_empty }
              end
            end
          end
        end
      end

      describe docker_run_with_envs("consul_template_bootstrap_#{distro}", SERVICE_ENV: "dev", ALREADY_SET: "true") do
        [:consul, :vault].each do |backend_type|
          describe backend_type do
            describe "General sets work" do
              set_var(backend_type, :global, "GLOBAL_VAR", "global-var", service_env: "dev")
              set_var(backend_type, :global, "OLD_GLOBAL_VAR", "old-global-var", service_env: "dev", old_keys: true)
              set_var(backend_type, :product, "PRODUCT_VAR", "product-var", service_env: "dev")
              set_var(backend_type, :service, "SERVICE_VAR", "service-var", service_env: "dev")
              set_var(backend_type, :service, "OLD_SERVICE_VAR", "old-service-var", service_env: "dev", old_keys: true)

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "SERVICE_ENV", "dev" }
                its(:stdout) { should include_env "GLOBAL_VAR", "global-var" }
                its(:stdout) { should include_env "PRODUCT_VAR", "product-var" }
                its(:stdout) { should include_env "SERVICE_VAR", "service-var" }
                its(:stdout) { should include_env "OLD_GLOBAL_VAR", "old-global-var" }
                its(:stdout) { should include_env "OLD_SERVICE_VAR", "old-service-var" }
                its(:stderr) { should be_empty }
              end
            end

            describe "Already set envs are not overriden" do
              set_var(backend_type, :global, "ALREADY_SET", "false", service_env: "dev")

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "SERVICE_ENV", "dev" }
                its(:stdout) { should_not include_env "ALREADY_SET", "false" }
                its(:stdout) { should include_env "ALREADY_SET", "true" }
                its(:stderr) { should be_empty }
              end
            end
          end
        end
      end

      describe docker_run_with_envs("consul_template_bootstrap_#{distro}", SERVICE_ENV: "peer-rise-runtime-1768", ALREADY_SET: "true") do
        [:consul, :vault].each do |backend_type|
          describe backend_type do
            describe "General sets work" do
              set_var(backend_type, :global, "GLOBAL_VAR", "global-var")
              set_var(backend_type, :product, "PRODUCT_VAR", "product-var")
              set_var(backend_type, :service, "SERVICE_VAR", "service-var")

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "GLOBAL_VAR", "global-var" }
                its(:stdout) { should include_env "PRODUCT_VAR", "product-var" }
                its(:stdout) { should include_env "SERVICE_VAR", "service-var" }
                its(:stdout) { should include_env "SERVICE_ENV", "peer-rise-runtime-1768" }
                its(:stderr) { should be_empty }
              end
            end

            describe "Already set envs are not overriden" do
              set_var(backend_type, :global, "ALREADY_SET", "false")

              describe entrypoint_command("env") do
                its(:stdout) { should include_env "SERVICE_ENV", "peer-rise-runtime-1768" }
                its(:stdout) { should_not include_env "ALREADY_SET", "false" }
                its(:stdout) { should include_env "ALREADY_SET", "true" }
                its(:stderr) { should be_empty }
              end
            end
          end
        end
      end
    end
  end
end

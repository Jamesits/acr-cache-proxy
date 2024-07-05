job "acr-cache-proxy" {
  type      = "system"
  namespace = "system"
  node_pool = "all"
  priority  = 10 # this thing is failsafe in most cases

  update {
    max_parallel = 1
    stagger      = "60s"
    auto_revert  = true
  }

  group "acr-cache-proxy" {
    restart {
      mode = "delay"
    }

    network {
      mode = "bridge"
      port "http" {
        static = 8080 # please change this
        to     = 80
      }
    }

    task "acr-cache-proxy" {
      driver = "docker"
      config {
        image = "jamesits/acr-cache-proxy:latest"
        ports = ["http"]
        args = [
          "--upstream-domain",
          "example.azurecr.io", # please change this
          "--upstream-prefix",
          "hub",                # please change this
        ]

        # mount SSL certificates inside the container
        # assuming Debian or derivatives; path for other distros might vary
        volumes = [ "/usr/share/ca-certificates:/usr/share/ca-certificates:ro", "/etc/ssl/certs:/etc/ssl/certs:ro" ]
      }

      resources {
        cpu        = 100
        memory     = 320
        memory_max = 640
      }

      env {
        # hardcoded user-managed identity
        AZURE_CLIENT_ID = "00000000-0000-0000-000000000000"
      }

// # automatic environment variables via Nomad variables
//       template {
//         change_mode          = "restart"
//         destination          = "${NOMAD_SECRETS_DIR}/.env"
//         env                  = true
//         error_on_missing_key = true
//         splay                = "60s"
//         data                 = <<EOT
// {{- with nomadVar (print "nomad/jobs/" (env "NOMAD_JOB_NAME")) }}
// {{- range .Tuples }}
// {{ .K }}={{ .V }}
// {{- end }}
// {{- end }}
// EOT
//       }
    }
  }
}

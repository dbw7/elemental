[Unit]
Description=Elemental Registry
After=network.target
{{- if .PayloadScript }}
Wants=network-online.target
After=network-online.target
{{- end }}
Before=k8s-config-installer.service rke2-server.service rke2-agent.service

[Service]
Type=exec
{{- if .PayloadScript }}
ExecStartPre=/bin/bash {{ .PayloadScript }}
{{- end }}
ExecStart={{ .Hauler }} store serve registry --store {{ .StoreDir }} --directory {{ .Backend }} --port {{ .Port }}
ExecStartPost=/usr/bin/curl --fail --silent --retry 30 --retry-delay 1 --retry-connrefused http://127.0.0.1:{{ .Port }}/v2/
TimeoutStartSec={{ .TimeoutStartSec }}
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target

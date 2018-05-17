package internal

var APIServerDefaultArgs = []string{
	"-v=6",
	"--etcd-servers={{ if .EtcdURL }}{{ .EtcdURL.String }}{{ end }}",
	"--cert-dir={{ .CertDir }}",
	"--secure-port={{ if .URL }}{{ .URL.Port }}{{ end }}",
	"--bind-address={{ if .URL }}{{ .URL.Hostname }}{{ end }}",
}

func DoAPIServerArgDefaulting(args []string) []string {
	if len(args) != 0 {
		return args
	}

	return APIServerDefaultArgs
}

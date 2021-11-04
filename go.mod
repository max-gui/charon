module github.com/max-gui/charon

go 1.15

require (
	github.com/gin-gonic/gin v1.7.4
	github.com/hashicorp/consul/api v1.11.0
	github.com/max-gui/consulagent v0.0.0-00010101000000-000000000000
	github.com/max-gui/logagent v0.0.0-00010101000000-000000000000
	github.com/max-gui/regagent v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	github.com/zsais/go-gin-prometheus v0.1.0
)

// replace github.com/max-gui/regagent => ../Regagent
replace github.com/max-gui/logagent => ../logagent

replace github.com/max-gui/regagent => ../regagent

replace github.com/max-gui/consulagent => ../consulagent

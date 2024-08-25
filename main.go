// Code generated by hertz generator.

package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/tundrawork/DG-citrus/config"
)

func main() {
	config.Init()

	h := server.Default(server.WithHostPorts(":" + config.Conf.Port))
	// https://github.com/cloudwego/hertz/issues/121
	h.NoHijackConnPool = true
	h.LoadHTMLGlob("resources/views/*")
	register(h)
	h.Spin()
}

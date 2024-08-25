package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/tundrawork/DG-citrus/config"
)

func HomeHandler(ctx context.Context, c *app.RequestContext) {
	c.HTML(http.StatusOK, "index.tmpl", utils.H{
		"host": config.Conf.HostName,
	})
}

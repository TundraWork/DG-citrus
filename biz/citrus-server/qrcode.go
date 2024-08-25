package citrus_server

import (
	"fmt"
	"io"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"golang.org/x/image/colornames"
)

const (
	DGAppWebsiteLink  = "https://www.dungeon-lab.com/app-download.php"
	DGAppWebsocketTag = "DGLAB-SOCKET"
)

type qrcodeWriteCloser struct {
	io.Writer
}

func (wc qrcodeWriteCloser) Close() error {
	// do nothing here as the closing operation should be executed by hertz
	return nil
}

// newQrcodeWriteCloser wraps an io.Writer with a Close method, returning an io.WriteCloser.
func newQrcodeWriteCloser(w io.Writer) io.WriteCloser {
	return &qrcodeWriteCloser{Writer: w}
}

func sendDGAppBindingCode(bodyWriter io.Writer, host string, secureId ClientSecureId) error {
	payload := fmt.Sprintf("%s#%s#wss://%s/app/%s", DGAppWebsiteLink, DGAppWebsocketTag, host, secureId)
	qrc, err := qrcode.New(payload)
	if err != nil {
		return err
	}
	writer := standard.NewWithWriter(
		newQrcodeWriteCloser(bodyWriter),
		standard.WithBgColor(colornames.Lightpink),
		standard.WithCircleShape(),
	)
	err = qrc.Save(writer)
	if err != nil {
		return err
	}
	return nil
}

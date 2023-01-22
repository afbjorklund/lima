package progressbar

import (
	"fmt"
	"os"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/mattn/go-isatty"
)

var OutputJSON bool

var elementInt pb.ElementFunc = func(state *pb.State, args ...string) string {
        if len(args) == 0 {
                return ""
        }
        var v int64
        switch args[0] {
        case "current":
                v = state.Current()
        case "total":
                v = state.Total()
        default:
                v = 0
        }
        return fmt.Sprint(v)
}

func New(size int64) (*pb.ProgressBar, error) {
	bar := pb.New64(size)

	bar.Set(pb.Bytes, true)
	if OutputJSON {
                bar.Set(pb.Terminal, false)
                bar.Set(pb.ReturnSymbol, "")
                pb.RegisterElement("int", elementInt, false)
                bar.SetTemplateString(`{"current":{{int . "current"}},"total":{{int . "total"}},` +
                        `"percent":"{{percent .}}","speed":"{{speed . "%s/s"}}",` +
                        `"url":"{{string . "url"}}"}` + "\n")
                bar.SetRefreshRate(1 * time.Second)
	} else if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		bar.SetTemplateString(`{{counters . }} {{bar . | green }} {{percent .}} {{speed . "%s/s"}}`)
		bar.SetRefreshRate(200 * time.Millisecond)
	} else {
		bar.Set(pb.Terminal, false)
		bar.Set(pb.ReturnSymbol, "\n")
		bar.SetTemplateString(`{{counters . }} ({{percent .}}) {{speed . "%s/s"}}`)
		bar.SetRefreshRate(5 * time.Second)
	}
	bar.SetWidth(80)
	if err := bar.Err(); err != nil {
		return nil, err
	}

	return bar, nil
}

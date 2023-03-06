package ssaviz

// This file borrows a lot from https://github.com/golang/go/blob/master/src/cmd/cover/html.go

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"

	"github.com/skratchdot/open-golang/open"
)

// HTML is a HTML report that contains multiple [Graph]s.
type HTML struct {
	html []byte
	path string
}

func buildHTML(graphs []*Graph) (*HTML, error) {
	data := templateData{}

	for _, g := range graphs {
		svg, err := g.Render(SVG)
		if err != nil {
			return nil, fmt.Errorf("failed to render %s of %s: %w", g.Name, SVG, err)
		}
		data.Graphs = append(data.Graphs, &templateGraph{
			Name: g.Name,
			SVG:  template.HTML(svg),
		})
	}

	var buf bytes.Buffer
	if err := htmlTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to exec template: %w", err)
	}

	return &HTML{html: buf.Bytes()}, nil
}

// Save saves HTML report to given path.
func (r *HTML) Save(path string) error {
	// Record recently path then it can be viewed by [HTML.View].
	if filepath.IsAbs(path) {
		r.path = path
	} else {
		wd, _ := os.Getwd()
		r.path, _ = filepath.Rel(wd, path)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer f.Close()
	_, err = f.Write(r.html)
	return err
}

// View opens the HTML report in system default application.
//
// NOTE: This function will leave a HTML file named "go-ssaviz-*.html"
// in the default temporary directory if the HTML has never been saved.
func (r *HTML) View() error {
	if _, err := os.Stat(r.path); errors.Is(err, os.ErrNotExist) {
		file, err := os.CreateTemp("", "go-ssaviz-*.html")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		r.path = file.Name()
		file.Write(r.html)
		file.Close()
	}

	url := &url.URL{Scheme: "file", Path: r.path}
	if err := open.Run(url.String()); err != nil {
		return fmt.Errorf("failed to open URL %s: %w", url, err)
	}

	return nil
}

// TODO:
// func (r *HTML) Serve(ctx context.Context, addr string) error {
// }

type templateData struct {
	Graphs []*templateGraph
}

type templateGraph struct {
	Name string
	SVG  template.HTML
}

var htmlTemplate = template.Must(template.New("html").Parse(tmplHTML))

const tmplHTML = `
<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<title>go-ssaviz</title>
		<style>
			#topbar {
				background: black;
				position: fixed;
				top: 0; left: 0; right: 0;
				height: 42px;
				border-bottom: 1px solid rgb(80, 80, 80);
			}
			#content {
				margin-top: 50px;
			}
			#nav {
				float: left;
				margin-left: 10px;
				margin-top: 10px;
			}
		</style>
	</head>
	<body>
		<div id="topbar">
			<div id="nav">
				<select id="files">
				{{range $i, $f := .Graphs}}
				<option value="file{{$i}}">{{$f.Name}}</option>
				{{end}}
				</select>
			</div>
		</div>
		<div id="content">
		{{range $i, $f := .Graphs}}
		<div class="file" id="file{{$i}}" style="display: none">{{$f.SVG}}</div>
		{{end}}
		</div>
	</body>
	<script>
	// Switch graphs.
	(function() {
		var files = document.getElementById('files');
		var visible;
		files.addEventListener('change', onChange, false);
		function select(part) {
			if (visible)
				visible.style.display = 'none';
			visible = document.getElementById(part);
			if (!visible)
				return;
			files.value = part;
			visible.style.display = 'block';
			location.hash = part;
		}
		function onChange() {
			select(files.value);
			window.scrollTo(0, 0);
		}
		if (location.hash != "") {
			select(location.hash.substr(1));
		}
		if (!visible) {
			select("file0");
		}
	})();

	// Drag to move.
	// ref: https://stackoverflow.com/a/45831670
	var divOverlay = document.getElementById ("content");
	var isDown = false;
	divOverlay.addEventListener('mousedown', function(e) {
		isDown = true;
	}, true);

	document.addEventListener('mouseup', function() {
	  isDown = false;
	}, true);

	document.addEventListener('mousemove', function(event) {
	   event.preventDefault();
	   if (isDown) {
	   var deltaX = event.movementX;
	   var deltaY = event.movementY;
	  var rect = divOverlay.getBoundingClientRect();
	  divOverlay.style.left = rect.x + deltaX + 'px';
	  divOverlay.style.top  = rect.x + deltaX + 'px';
	 }
	}, true);
	</script>
</html>
`

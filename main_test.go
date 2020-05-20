package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_statusSwitch(t *testing.T) {

	assertOK := func(name, in string, expect *Reply) {
		t.Helper()
		got := statusSwitch(in)
		if !reflect.DeepEqual(got, expect) {
			t.Errorf("[%v] expect %#v, got %#v",
				name,
				expect,
				got,
			)
		}
	}

	t.Run("TestArbitraryRouter", func(t *testing.T) {
		tests := []struct {
			name   string
			in     string
			expect Reply
		}{
			{"any1", "qwrqqzsg", Reply{200, robot + " " + "qwrqqzsg"}},
			{"any2", "w45tye5h34w324r", Reply{200, robot + " " + "w45tye5h34w324r"}},
			{"any3", "fshgws", Reply{200, robot + " " + "fshgws"}},
			{"any4", "fshg   ws", Reply{200, robot + " " + "fshg   ws"}},
			{"empty", "", Reply{200, emptyCall}},
		}
		for _, test := range tests {
			assertOK(test.name, test.in, &test.expect)
		}
	})

	t.Run("StandardReply", func(t *testing.T) {
		tests := []struct {
			name   string
			in     string
			expect Reply
		}{
			{"test 200", "200", Reply{200, resp200}},
			{"test 400", "400", Reply{400, resp400}},
			{"test 500", "500", Reply{500, resp500}},
		}
		for _, test := range tests {
			assertOK(test.name, test.in, &test.expect)
		}
	})
}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

const mockResp = "MockResource"

type MockResource struct {
	addr string
}

func (r MockResource) API() string { return r.addr }
func (r MockResource) Get() *Reply {
	return &Reply{http.StatusOK, mockResp}
}

func Test_Routes(t *testing.T) {

	assertEqual := func(name string, got, expect Reply) {
		t.Helper()
		if !reflect.DeepEqual(got, expect) {
			t.Errorf("[%v] expect %#v, got %#v",
				name,
				expect,
				got,
			)
		}
	}

	assertStatus := func(name string, got, expect Reply) {
		t.Helper()
		if got.status != expect.status {
			t.Errorf("[%v] expect %#v, got %#v",
				name,
				expect,
				got,
			)
		}
	}

	t.Run("Basic Routes", func(t *testing.T) {
		tests := []struct {
			name     string
			endpoint string
			expect   Reply
		}{
			{"", "/", Reply{http.StatusOK, emptyCall}},
			{"", "/200", Reply{http.StatusOK, resp200}},
			{"", "/400", Reply{http.StatusBadRequest, resp400}},
			{"", "/500", Reply{http.StatusInternalServerError, resp500}},
		}

		for _, test := range tests {
			r := setupDefaults()
			resp := performRequest(r, "GET", test.endpoint)
			replyWrap := Reply{resp.Code, resp.Body.String()}
			assertEqual(test.name, replyWrap, test.expect)
		}
	})

	t.Run("Builtin Services", func(t *testing.T) {
		fwds := []Forwarder{
			*makeForwarder("/getip", MockResource{"ipify"}),
		}
		r := setupDefaults()
		r = setupForwardings(r, fwds)

		tests := []struct {
			name     string
			endpoint string
			expect   Reply
		}{
			{"", "/getip", Reply{http.StatusOK, ""}},
		}
		for _, test := range tests {
			resp := performRequest(r, "GET", test.endpoint)
			replyWrap := Reply{resp.Code, resp.Body.String()}
			assertStatus(test.name, replyWrap, test.expect)
		}
	})
}

func Test_fwdOption(t *testing.T) {

	t.Run("AcceptableCases", func(t *testing.T) {
		tests := []struct {
			name   string
			option string
			expect Forwarders
		}{
			{"OK 1", "/foo:1.2.3.4:8080/bar", []Forwarder{Forwarder{"/foo", Resource{"http://1.2.3.4:8080/bar"}}}},
			{"OK 2", "/foo:http://1.2.3.4:8080/bar", []Forwarder{Forwarder{"/foo", Resource{"http://1.2.3.4:8080/bar"}}}},
			{"OK 3", "/foo:https://1.2.3.4:8080/bar", []Forwarder{Forwarder{"/foo", Resource{"https://1.2.3.4:8080/bar"}}}},
		}
		for _, test := range tests {
			var fwds Forwarders
			err := fwds.Set(test.option)
			if err != nil {
				t.Errorf("[%v] got err: %v, expect %#v", test.name, err, test.expect)
			}
		}
	})

	t.Run("UnacceptableCases", func(t *testing.T) {
		tests := []struct {
			name   string
			option string
		}{
			{"NG 1", "foo:1.2.3.4:8080/bar"},
			{"NG 2", "/foo:http:///"},
			{"NG 3", "/foo:https:/1.2.3.4:8080/bar"},
			{"NG 4", "/foo:https//1.2.3.4:8080/bar"},
			{"NG 5", "/foo:/1.2.3.4:8080/bar"},
			{"NG 6", "/foo"},
			{"NG 7", "/foo:"},
		}
		for _, test := range tests {
			var fwds Forwarders
			err := fwds.Set(test.option)
			if err == nil {
				t.Errorf("[%v] expect an error, got %#v", test.name, fwds)
			} else {
				t.Logf("[PASS][%v] got the error: '%v'", test.name, err)
			}
		}
	})
}

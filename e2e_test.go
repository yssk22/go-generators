package gqlgen

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/yssk22/go-generators/enum"
	"github.com/yssk22/go-generators/enum/entgo"
	"github.com/yssk22/go-generators/enum/gqlgen"
	"github.com/yssk22/go-generators/graphql"
	graphqlgqlgen "github.com/yssk22/go-generators/graphql/gqlgen"
)

func TestE2E(t *testing.T) {
	// cleanup gqlgen dir
	os.RemoveAll("./testdata/e2e/gqlgen")
	// generate enums
	err := enum.Generate("./testdata/e2e/models", gqlgen.NewGenerator())
	if err != nil {
		t.Fatalf("failed to generate enum for gqlgen: %v", err)
	}
	err = enum.Generate("./testdata/e2e/models", entgo.NewGenerator())
	if err != nil {
		t.Fatalf("failed to generate enum for entgo: %v", err)
	}
	err = graphql.Generate("./testdata/e2e/models", graphqlgqlgen.NewGenerator("./testdata/e2e/gqlgen"))
	if err != nil {
		t.Fatalf("failed to generate a server code: %v", err)
	}
	// launch a new server process
	var cmdErr error
	cmd := exec.Command("go", "run", "./")
	cmd.Dir = "./testdata/e2e/"
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	go func() {
		cmdErr = cmd.Run()
	}()
	defer func() {
		if cmdErr != nil {
			t.Fatalf("cannot start a process: %s", cmdErr)
		}
		if cmd.Process != nil {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	}()
	time.Sleep(5 * time.Second)

	// now request a query and make sure the server can respond to it.
	query := `{
query_example {
	field_string
	field_user_defined_scalar
	field_user_defined_enum
  }
}`
	expect := map[string]interface{}{
		"query_example": map[string]interface{}{
			"field_string":              "strValue",
			"field_user_defined_scalar": "no",
			"field_user_defined_enum":   "ValueA",
		},
	}
	requestBody, _ := json.Marshal(map[string]interface{}{
		"query": query,
	})
	var resp *http.Response
	var retry = 0
	for retry < 10 {
		resp, err = http.Post("http://localhost:8080/query", "application/json", bytes.NewReader(requestBody))
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			retry++
		} else {
			break
		}
	}
	if err != nil {
		t.Fatalf("cannot connect server: %v", err)
	}
	defer resp.Body.Close()
	var v = make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&v)
	t.Log(v)
	got := v["data"].(map[string]interface{})
	if !reflect.DeepEqual(expect, got) {
		t.Fatalf("expected: %v, got: %v", expect, got)
	}
}
